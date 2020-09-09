package app

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	fineV1Pb "lectiongrpc/pkg/fine/v1"
)

type Server struct {
}

func NewServer() *Server {
	return &Server{}
}

func (s *Server) FindByUserId(ctx context.Context, request *fineV1Pb.FinesRequest) (*fineV1Pb.FinesResponse, error) {
	if request.UserId == 1 {
		return &fineV1Pb.FinesResponse{
			UserId: 1,
			Items:  []*fineV1Pb.Fine{},
		}, nil
	}

	return nil, status.Errorf(codes.NotFound, "user with id %d not found", request.GetUserId())
}
