package main

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"lectiongrpc/cmd/grpc/server/app"
	eventV1Pb "lectiongrpc/pkg/event/v1"
	"log"
	"net"
	"os"
)

const defaultPort = "9999"
const defaultHost = "0.0.0.0"
const defaultCertificatePath = "./tls/certificate.pem"
const defaultPrivateKeyPath = "./tls/key.pem"

func main() {
	port, ok := os.LookupEnv("APP_PORT")
	if !ok {
		port = defaultPort
	}

	host, ok := os.LookupEnv("APP_HOST")
	if !ok {
		host = defaultHost
	}

	certificatePath, ok := os.LookupEnv("APP_CERT_PATH")
	if !ok {
		certificatePath = defaultCertificatePath
	}

	privateKeyPath, ok := os.LookupEnv("APP_PRIVATE_KEY_PATH")
	if !ok {
		privateKeyPath = defaultPrivateKeyPath
	}

	if err := execute(net.JoinHostPort(host, port), certificatePath, privateKeyPath); err != nil {
		log.Print(err)
		os.Exit(1)
	}
}

func execute(addr string, certificatePath string, privateKeyPath string) (err error) {
	creds, err := credentials.NewServerTLSFromFile(certificatePath, privateKeyPath)
	if err != nil {
		return err
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer(grpc.Creds(creds))
	server := app.NewServer()
	eventV1Pb.RegisterEventServiceServer(grpcServer, server)

	return grpcServer.Serve(listener)
}
