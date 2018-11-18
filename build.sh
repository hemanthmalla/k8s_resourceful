#!/usr/bin/env bash

protoc -I rpc rpc/resourceupdate.proto --go_out=plugins=grpc:rpc

dep ensure

env GOOS=linux GOARCH=amd64 go build -o bin/server server.go
env GOOS=linux GOARCH=amd64 go build -o bin/resourceful cmd.go

docker build -t hemanthmalla/k8s_resourceful ./

if [ "$1" ]; then
    docker tag hemanthmalla/k8s_resourceful:latest hemanthmalla/k8s_resourceful:$1
    docker push hemanthmalla/k8s_resourceful:$1
fi