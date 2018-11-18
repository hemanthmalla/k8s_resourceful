package main

import (
	"encoding/json"
	"fmt"
	pb "github.com/hemanthmalla/vscale/rpc"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"
)

const (
	GRPCPort = "50051"
)

func makeReq(request *pb.UpdateRequest, ip string) {
	var addr string
	if ip != ""{
		addr = ip + ":" + GRPCPort
	}else {
		addr = "localhost:" + GRPCPort
	}
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	c := pb.NewUpdaterClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	status, err := c.UpdateContainerResource(ctx, request)
	if err != nil {
		fmt.Errorf(err.Error())
	} else {
		fmt.Println(status.Msg)
	}
}

func getWorkerIP(context *cli.Context) string{
	singleNode := context.String("singlenode")
	if singleNode != ""{
		return singleNode
	}
	if context.Bool("minikube"){
		output, err := exec.Command("minikube", "ip").CombinedOutput()
		if err != nil {
			fmt.Println("Minikube IP couldn't be determined")
		}else{
			return strings.Trim(string(output), "\n")
		}
	}
	ns := context.String("namespace")
	pod := context.String("pod")

	type PodResponse struct {
		Status struct {
			HostIP    string    `json:"hostIP"`
		} `json:"status"`
	}

	cmd := fmt.Sprintf("get pod %s -o json -n %s", pod, ns)
	output, err := exec.Command("kubectl", strings.Split(cmd, " ")...).CombinedOutput()
	if err != nil {
		fmt.Errorf(err.Error())
		return ""
	}
	var data PodResponse
	json.Unmarshal(output, &data)
	return data.Status.HostIP
}

var updateContainerCommand = cli.Command{
	Name:      "update",
	Usage:     "Update resource limits of a running container",
	ArgsUsage: "CONTAINER-NAME",
	Flags: []cli.Flag{
		cli.Int64Flag{
			Name:  "cpu",
			Usage: "Impose a CPU CFS quota on the container. The number of microseconds per cpu-period that the container is limited to before throttled",
		},
		cli.Int64Flag{
			Name:  "memory",
			Usage: "Memory limit (in bytes)",
		},
		cli.StringFlag{
			Name:  "namespace, n",
			Usage: "Namespace of the container",
		},
		cli.StringFlag{
			Name:  "pod, p",
			Usage: "Name of the Pod",
		},
		cli.BoolFlag{
			Name: "minikube",
			Usage: "Set this flag if you're running k8s in minikube",
		},
		cli.StringFlag{
			Name: "singlenode, s",
			Usage: "IP address of single node K8S cluster on which resourceful is running",
		},
	},
	Action: func(context *cli.Context) error {
		if context.NArg() == 0 {
			return cli.ShowSubcommandHelp(context)
		}
		ip := getWorkerIP(context)
		makeReq(&pb.UpdateRequest{
			Namespace:     context.String("namespace"),
			PodName:       context.String("pod"),
			ContainerName: context.Args().Get(0),
			Memory:        context.Int64("memory"),
			Cpu:           context.Int64("cpu"),
		}, ip)

		return nil
	},
}

func main() {
	app := cli.NewApp()
	app.Name = "resourceful"
	app.Usage = "Kubectl plugin to update container resource limits"
	app.Version = "0.1"

	app.Commands = []cli.Command{
		updateContainerCommand,
	}

	// sort all flags
	for _, cmd := range app.Commands {
		sort.Sort(cli.FlagsByName(cmd.Flags))
	}
	sort.Sort(cli.FlagsByName(app.Flags))

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}

}
