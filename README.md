## Resourceful - K8S in-place resource updates

Disclaimer : This is a POC / Hack. Production usage has not been tested.

[![asciicast](https://asciinema.org/a/212425.svg)](https://asciinema.org/a/212425)

#### Installation

Resourceful has a GRPC based server component and a command line tool.

The server component needs to be present on all the worker nodes and can be installed with :

```
kubectl create -f k8s/deamonset.yaml
```

CLI is designed to be a kubectl plugin, the following script builds the binary and places it in your path. This step requires a working installation of go, dep and protoc compiler. 

```
bash install.sh
```

#### Usage

```
kubectl resourceful update nginx -n resourceful-test -p nginx --memory 200000000 --cpu 40000

kubectl resourceful update --help

NAME:
   kubectl-resourceful update - Update resource limits of a running container

USAGE:
   kubectl-resourceful update [command options] CONTAINER-NAME

OPTIONS:
   --cpu value                   Impose a CPU CFS quota on the container. The number of microseconds per cpu-period that the container is limited to before throttled (default: 0)
   --memory value                Memory limit (in bytes) (default: 0)
   --minikube                    Set this flag if you're running k8s in minikube
   --namespace value, -n value   Namespace of the container
   --pod value, -p value         Name of the Pod
   --singlenode value, -s value  IP address of single node K8S cluster on which resourceful is running
```


#### How does it work ?

Docker announced support for resource updates in [1.10](https://blog.docker.com/2016/02/docker-1-10/).
Kubernetes does not currently(as of 1.13) support in-place restart free resource updates. This project is only a hack to be able update resource limits from kubectl. More streamlined methods to natively support this are [WIP](https://github.com/kubernetes/kubernetes/issues/5774).

![CRI](https://cl.ly/3I2p0D1V0T26/Image%202016-12-19%20at%2017.13.16.png "CRI Arch.")

Kubernetes(kubelet) already has support for CRI, which is an abstraction over serveral container runtimes like docker, CRI-O, rkt, etc. 

Interface to the container runtime is implemented with GRPC and unix domain sockets for IPC. Domain sockets cannot be used for communicating with other machines. Hence, resourceful implements a thin wrapper on top of the CRI interface to allow external access.

This service/wrapper is placed on all worker nodes using a deamonset.

The CLI component then determines the host machine of a given container and routes the update request to the appropriate worker node, which is finally delegated to CRI to perform the update operation.


#### Limitations

* Memory limits can only be set to values less than the swap memory allocated for the container [https://github.com/kubernetes/kubernetes/issues/69793]
* The machine on which CLI is executed should have routes to worker nodes.
* kubectl and kubeconfig should be configured and available in path.


#### TODO

* Health checks for GRPC service
* Notify kuberentes of the changes to container's resources without triggering a restart.


