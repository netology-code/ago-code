package app

import (
	"context"
	"io"
	eventV1Pb "lectiongrpc/pkg/event/v1"
	"log"
	"time"
)

type Server struct {
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) Unary(
	ctx context.Context,
	request *eventV1Pb.EventRequest,
) (*eventV1Pb.EventResponse, error) {
	log.Print(request)
	return &eventV1Pb.EventResponse{
		Id:      1,
		Payload: "Response",
	}, nil
}

func (s *Server) ServerStream(
	request *eventV1Pb.EventRequest,
	server eventV1Pb.EventService_ServerStreamServer,
) error {
	log.Print(request)
	for i := 1; i <= 5; i++ {
		time.Sleep(time.Second)
		if err := server.Send(&eventV1Pb.EventResponse{
			Id:      int64(i),
			Payload: "Response",
		}); err != nil {
			log.Print(err)
			return err
		}
	}
	return nil
}

func (s *Server) ClientStream(
	server eventV1Pb.EventService_ClientStreamServer,
) error {
	for {
		request, err := server.Recv()
		if err != nil {
			if err == io.EOF {
				return server.SendAndClose(&eventV1Pb.EventResponse{
					Id:      1,
					Payload: "Response",
				})
			}
			log.Print(err)
			return err
		}
		log.Print(request)
	}
}

func (s *Server) BidirectionalStream(
	server eventV1Pb.EventService_BidirectionalStreamServer,
) error {
	count := 0
	for {
		count++
		request, err := server.Recv()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		log.Print(request)
		err = server.Send(&eventV1Pb.EventResponse{
			Id:      request.Id,
			Payload: "Response",
		})
		if err != nil {
			return err
		}

		if count == 10 { // пример выхода по условию
			return nil
		}
	}
}
