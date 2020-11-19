package main

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	eventV1Pb "lectiongrpc/pkg/event/v1"
	"log"
	"net"
	"os"
	"time"
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
		log.Print(err)
		os.Exit(1)
	}
}

func execute(addr string) (err error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer func() {
		if cerr := conn.Close(); cerr != nil {
			if err == nil {
				err = cerr
				return
			}
			log.Print(err)
		}
	}()

	client := eventV1Pb.NewEventServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	stream, err := client.ClientStream(ctx)
	if err != nil {
		return err
	}

	go func() {
		<-time.After(time.Second)
		cancel()
	}()

	for i := 1; i <= 5; i++ {
		err := stream.Send(&eventV1Pb.EventRequest{
			Id:      int64(i),
			Payload: "Request",
		})
		if err != nil {
			if status.Code(err) == codes.Canceled {
				log.Print("canceled")
			}
			if ctx.Err() == context.Canceled {
				log.Print("context canceled")
			}
			return err
		}
		time.Sleep(time.Second)
	}
	response, err := stream.CloseAndRecv()
	if err != nil {
		return err
	}
	log.Print(response)
	log.Print("finished")
	return nil
}
