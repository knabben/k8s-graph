package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/goccy/go-graphviz"
	"github.com/goccy/go-graphviz/cgraph"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/restmapper"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	kubeconfig *string

	group    *string
	version  *string
	resource *string

	name      *string
	namespace *string
)

type Graphz struct {
	client dynamic.Interface
	mapper meta.RESTMapper
	graph  *cgraph.Graph
}

func init() {
	kubeconfig = flag.String("kubeconfig", "kubeconfig", "(optional) absolute path to the kubeconfig file")

	// object GVR flags
	group = flag.String("group", "", "Group of the object")
	version = flag.String("version", "v1", "Version of the object")
	resource = flag.String("resource", "pods", "Resource of the object")

	// object details flags
	name = flag.String("name", "coredns", "Name of the object")
	namespace = flag.String("namespace", "default", "Namespace of the object")
}

// fetchOnwerReferences return the references to the owners of a particular object
func (g *Graphz) fetchOnwerReferences(group, version, resource, namespace, name string) ([]interface{}, error) {
	scheme := schema.GroupVersionResource{Group: group, Version: version, Resource: resource}
	res, err := g.client.Resource(scheme).Namespace(namespace).Get(context.TODO(), name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	metadata, exists, err := unstructured.NestedMap(res.Object, "metadata")
	if err != nil && !exists {
		panic(err)
	}

	owners := metadata["ownerReferences"]
	if owners != nil { // typecase on owners existence
		return owners.([]interface{}), nil
	}
	return nil, nil
}

// fetchRecursiveOwners traverse with bfs and create edges between nodes
func (g *Graphz) fetchRecursiveOwners(owners []interface{}, node *cgraph.Node) {
	for _, n := range owners {
		owner := n.(map[string]interface{})

		apiVersion := owner["apiVersion"].(string)
		avSplit := strings.Split(apiVersion, "/")
		ownerKind := owner["kind"].(string)
		ownerName := owner["name"].(string)

		// create owner node
		gnode, err := g.graph.CreateNode(generateNodeName(ownerName, apiVersion, ownerKind))
		if err != nil {
			panic(err.Error())
		}

		// create owner node and node edge
		_, err = g.graph.CreateEdge("", node, gnode)
		if err != nil {
			panic(err.Error())
		}

		// convert GVK to GVR, so it's possible to fetch object owners
		gk := schema.ParseGroupKind(fmt.Sprintf("%s.%s", ownerKind, avSplit[0]))
		mapping, err := g.mapper.RESTMapping(gk)
		if err != nil {
			panic(err)
		}

		owners, err := g.fetchOnwerReferences(mapping.Resource.Group, mapping.Resource.Version, mapping.Resource.Resource, *namespace, ownerName)
		if err != nil {
			panic(err.Error())
		}
		if owners != nil {
			g.fetchRecursiveOwners(owners, gnode)
		}
	}
}

// NewClients return dynamic and discovery clients
func NewClients(kubeconfig *string) (*discovery.DiscoveryClient, dynamic.Interface, error) {
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, nil, err
	}

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}

	return discovery.NewDiscoveryClientForConfigOrDie(config), client, nil
}

// generateNodeName return GRV/GRK formatted name
func generateNodeName(name, group, resource string) string {
	return fmt.Sprintf("%s\n%s\n%s", resource, group, name)
}

func main() {
	var (
		owners []interface{}
		buf    bytes.Buffer
	)

	flag.Parse()

	// create graphviz setup
	g := graphviz.New()
	graph, err := g.Graph()
	if err != nil {
		log.Fatal(err)
	}

	defer func() {
		if err := graph.Close(); err != nil {
			log.Fatal(err)
		}
		g.Close()
	}()

	// return discovery and dynamic clients
	discovery, client, err := NewClients(kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// setup initial mapper from gvr -> gvk
	groupResources, err := restmapper.GetAPIGroupResources(discovery)
	if err != nil {
		panic(err.Error())
	}
	mapper := restmapper.NewDiscoveryRESTMapper(groupResources)

	// add initial node from passed parameter
	gz := &Graphz{
		client: client,
		mapper: mapper,
		graph:  graph,
	}
	node, err := graph.CreateNode(generateNodeName(*name, *group, *resource))
	if err != nil {
		log.Fatal(err)
	}

	// fetch initial parent owners of the object
	if owners, err = gz.fetchOnwerReferences(*group, *version, *resource, *namespace, *name); err != nil {
		panic(err.Error())
	}
	gz.fetchRecursiveOwners(owners, node)

	// print out diagram
	if err := g.Render(graph, "dot", &buf); err != nil {
		log.Fatalf("%+v", err)
	}
	fmt.Println(buf.String())
}
