package main

import (
    "fmt"
	"log"
	"net"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pb "github.com/amitsaha/apigateway-demo/grpc-app-1/verify"
)

const (
	port = ":50051"
)

type server struct{}

func (s *server) VerifyUser(ctx context.Context, in *pb.VerifyRequest) (*pb.VerifyReply, error) {
    msg := fmt.Sprintf("Verified: %d", in.Id)
    return &pb.VerifyReply{Message: msg}, nil
}

func main() {
	lis, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterUserVerifyServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
