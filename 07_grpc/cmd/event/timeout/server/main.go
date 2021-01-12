package main

import (
	"google.golang.org/grpc"
	"lectiongrpc/cmd/event/timeout/server/app"
	eventV1Pb "lectiongrpc/pkg/event/v1"
	"net"
	"os"
)

const defaultPort = "9999"
const defaultHost = "0.0.0.0"

func main() {
	port, ok := os.LookupEnv("APP_PORT")
	if !ok {
		port = defaultPort
	}

	host, ok := os.LookupEnv("APP_HOST")
	if !ok {
		host = defaultHost
	}

	if err := execute(net.JoinHostPort(host, port)); err != nil {
		os.Exit(1)
	}
}

func execute(addr string) (err error) {
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	server := app.NewServer()
	eventV1Pb.RegisterEventServiceServer(grpcServer, server)

	return grpcServer.Serve(listener)
}
