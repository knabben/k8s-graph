# k8s-graph

Kubernetes ownerReferences dependency graph.

## Running

Suppose you have an `vSphereMachine` object on your cluster named `tkg-windows-wl-windows-containerd-jmjht`

```
go build ./...
./k8s-graph --name "tkg-windows-wl-windows-containerd-jmjht" --resource=vspheremachines --group "infrastructure.cluster.x-k8s.io" --version "v1beta1"
```

## Checking the GVR 

You can check the API resources available on your cluster with:

```
$ kubectl api-resources

clusterclasses                    cc           cluster.x-k8s.io/v1beta1                             true         ClusterClass
clusters                          cl           cluster.x-k8s.io/v1beta1                             true         Cluster
machinedeployments                md           cluster.x-k8s.io/v1beta1                             true         MachineDeployment
machinehealthchecks               mhc,mhcs     cluster.x-k8s.io/v1beta1                             true         MachineHealthCheck
machinepools                      mp           cluster.x-k8s.io/v1beta1                             true         MachinePool
machines                          ma           cluster.x-k8s.io/v1beta1                             true         Machine
machinesets                       ms           cluster.x-k8s.io/v1beta1                             true         MachineSet
...
vsphereclusteridentities                       infrastructure.cluster.x-k8s.io/v1beta1              false        VSphereClusterIdentity
vsphereclusters                                infrastructure.cluster.x-k8s.io/v1beta1              true         VSphereCluster
vsphereclustertemplates                        infrastructure.cluster.x-k8s.io/v1beta1              true         VSphereClusterTemplate
vspheredeploymentzones                         infrastructure.cluster.x-k8s.io/v1beta1              false        VSphereDeploymentZone
vspherefailuredomains                          infrastructure.cluster.x-k8s.io/v1beta1              false        VSphereFailureDomain
vspheremachines                                infrastructure.cluster.x-k8s.io/v1beta1              true         VSphereMachine
vspheremachinetemplates                        infrastructure.cluster.x-k8s.io/v1beta1              true         VSphereMachineTemplate
vspherevms                                     infrastructure.cluster.x-k8s.io/v1beta1              true         VSphereVM
```

## Generating graph

After running the command, uou can use some online graphic generator website: https://dreampuf.github.io/GraphvizOnline.


![Capture](https://user-images.githubusercontent.com/1223213/152265637-c5ac8542-7dd3-48ab-9c66-d47bd7cb0746.PNG)
