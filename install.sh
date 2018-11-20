#!/usr/bin/env bash
set -x
protoc -I rpc rpc/resourceupdate.proto --go_out=plugins=grpc:rpc

dep ensure

go build -o bin/server server.go
go build -o bin/kubectl-resourceful cmd.go

cp bin/kubectl-resourceful /usr/local/bin/

chmod +x /usr/local/bin/kubectl-resourceful

kubectl resourceful update --help

echo $'\n Invoke the above info with "kubectl resourceful update --help" \n Make sure resourceful\'s server component has been installed, if not it can be installed with kubectl "create -f k8s/deamonset.yaml"\n'
