package main

import (
	"context"
	"fmt"
	"github.com/hemanthmalla/vscale/rpc"
	rpcpb "github.com/hemanthmalla/vscale/rpc"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	pb "k8s.io/kubernetes/pkg/kubelet/apis/cri/runtime/v1alpha2"
	"k8s.io/kubernetes/pkg/kubelet/util"
	"log"
	"net"
	"time"
)

const (
	port = ":50051"
)

type updateOptions struct {
	// CPU CFS (Completely Fair Scheduler) period. Default: 0 (not specified).
	CPUPeriod int64
	// CPU CFS (Completely Fair Scheduler) quota. Default: 0 (not specified).
	CPUQuota int64
	// CPU shares (relative weight vs. other containers). Default: 0 (not specified).
	CPUShares int64
	// Memory limit in bytes. Default: 0 (not specified).
	MemoryLimitInBytes int64
	// OOMScoreAdj adjusts the oom-killer score. Default: 0 (not specified).
	OomScoreAdj int64
	// CpusetCpus constrains the allowed set of logical CPUs. Default: "" (not specified).
	CpusetCpus string
	// CpusetMems constrains the allowed set of memory nodes. Default: "" (not specified).
	CpusetMems string
}

type vscaleServer struct {

}

func getRuntimeClientConnection() (*grpc.ClientConn, error) {
	var RuntimeEndpoint = "unix:///var/run/dockershim.sock"

	addr, dialer, err := util.GetAddressAndDialer(RuntimeEndpoint)
	if err != nil {
		return nil, err
	}

	conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock(), grpc.WithTimeout(2 * time.Second), grpc.WithDialer(dialer))
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %v", err)
	}
	return conn, nil
}

func getContainerID(runtimeClient pb.RuntimeServiceClient, podName string, containerName string, namespace string) (id string, err error){

	filter := &pb.ContainerFilter{}
	filter.LabelSelector = map[string]string{
		"io.kubernetes.pod.name": podName,
		"io.kubernetes.container.name": containerName,
		"io.kubernetes.pod.namespace": namespace,
	}
	r, err := runtimeClient.ListContainers(context.Background(), &pb.ListContainersRequest{Filter: filter})
	logrus.Debugf("ListContainerResponse: %v", r)
	if err != nil {
		return "", err
	}
	if len(r.Containers) < 1 {
		return "", fmt.Errorf("container not found")
	}
	return r.Containers[0].Id, nil
}

func UpdateContainerResources(client pb.RuntimeServiceClient, ID string, opts updateOptions) error {
	if ID == "" {
		return fmt.Errorf("ID cannot be empty")
	}
	request := &pb.UpdateContainerResourcesRequest{
		ContainerId: ID,
		Linux: &pb.LinuxContainerResources{
			CpuPeriod:          opts.CPUPeriod,
			CpuQuota:           opts.CPUQuota,
			CpuShares:          opts.CPUShares,
			CpusetCpus:         opts.CpusetCpus,
			CpusetMems:         opts.CpusetMems,
			MemoryLimitInBytes: opts.MemoryLimitInBytes,
			OomScoreAdj:        opts.OomScoreAdj,
		},
	}
	logrus.Debugf("UpdateContainerResourcesRequest: %v", request)
	r, err := client.UpdateContainerResources(context.Background(), request)
	logrus.Debugf("UpdateContainerResourcesResponse: %v", r)
	if err != nil {
		return err
	}
	return nil
}

func (*vscaleServer) UpdateContainerResource(ctx context.Context, req *resourceupdate.UpdateRequest) (*resourceupdate.UpdateResponse, error) {

	/* Steps needed to update container resources :
	 *
	 * 1. Get a handle to CRI runtime client.
	 * 2. Find container ID based on namespace and pod name tags
	 * 3. Issue update request to CRI
	 */

	var conn *grpc.ClientConn
	var err error
	var runtimeClient pb.RuntimeServiceClient
	conn, err = getRuntimeClientConnection()
	runtimeClient = pb.NewRuntimeServiceClient(conn)

	 id, err := getContainerID(runtimeClient, req.PodName, req.ContainerName, req.Namespace)
	 if err != nil {
	 	logrus.Debugf("Pod : %v not found in the given namespace", req.PodName)
		 res := resourceupdate.UpdateResponse{
	 		Msg: fmt.Sprintf("Pod : %v not found in the given namespace", req.PodName),
	 		Success:false,
		 }
		 return &res, nil
	 }

	 options := updateOptions{
	 	MemoryLimitInBytes: req.Memory,
	 }

	 if req.Cpu != 0{
	 	options.CPUQuota = req.Cpu
	 }
	 err = UpdateContainerResources(runtimeClient, id, options)
	 if err != nil{
		 res := resourceupdate.UpdateResponse{
			 Msg: fmt.Sprintf("Unable to update resources ! %v", err.Error()),
			 Success:false,
		 }
		 return &res, nil
	 } else {
		 res := resourceupdate.UpdateResponse{
			 Msg: fmt.Sprint("Successfully updated resources !"),
			 Success:true,
		 }
		 return &res, nil
	 }
}

func main()  {
	logrus.SetLevel(logrus.DebugLevel)
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	fmt.Printf("Listening on port %s ...\n", port)
	s := grpc.NewServer()
	rpcpb.RegisterUpdaterServer(s, &vscaleServer{})
	// Register reflection service on gRPC server.
	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}



