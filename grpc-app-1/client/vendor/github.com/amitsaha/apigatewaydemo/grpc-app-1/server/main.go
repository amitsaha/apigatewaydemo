package main

import (
	"fmt"
	pb "github.com/amitsaha/apigatewaydemo/grpc-app-1/verify"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"net"
)

type server struct{}

func (s *server) VerifyUser(ctx context.Context, in *pb.VerifyRequest) (*pb.VerifyReply, error) {
	msg := fmt.Sprintf("Verified: %d", in.Id)
	return &pb.VerifyReply{Message: msg}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":6000")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterUserVerifyServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
