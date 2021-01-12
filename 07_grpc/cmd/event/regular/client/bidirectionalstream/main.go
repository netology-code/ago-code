package main

import (
	"context"
	"google.golang.org/grpc"
	"io"
	eventV1Pb "lectiongrpc/pkg/event/v1"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
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
	ctx, cancel := context.WithCancel(context.Background())
	stream, err := client.BidirectionalStream(ctx)
	if err != nil {
		return err
	}

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT)

	wg := sync.WaitGroup{}
	wg.Add(2)

	// sender
	go func() {
		defer wg.Done()
		i := 0
		for {
			select {
			case <-ctx.Done():
				log.Print("sender ctx done")
				err := stream.CloseSend()
				if err != nil {
					log.Print(err)
				}
				return
			case <-time.After(time.Second):
				i++
				err := stream.Send(&eventV1Pb.EventRequest{
					Id:      int64(i),
					Payload: "Request",
				})
				if err != nil {
					log.Print(err)
					cancel()
					return
				}
			}
		}
	}()

	// receiver
	go func() {
		defer func() {
			cancel()
			wg.Done()
		}()
		for {
			response, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					return
				}
				log.Printf("recv error: %v", err)
				return
			}
			log.Print(response)
		}
	}()

	// wait
	select {
	case <-ch:
		log.Print("got SIGINT")
		cancel()
	case <-ctx.Done():
		log.Print("ctx done")
	}
	wg.Wait()

	return err
}
