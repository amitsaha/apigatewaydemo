package main

import (
    "fmt"
	"log"
	"net"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
    consulapi "github.com/hashicorp/consul/api"
	pb "github.com/amitsaha/apigatewaydemo/grpc-app-1/verify"
    "os"
)

const (
    port = "127.0.0.1:50051"
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

    consulConfig := consulapi.DefaultConfig()
	consulClient, err := consulapi.NewClient(consulConfig)
	if err != nil {
        log.Fatalf("err", err)
        os.Exit(1)
	}

    agent := consulClient.Agent()
	reg := &consulapi.AgentServiceRegistration{
		Name: "verification",
		Port: 50051,
	}
	if err := agent.ServiceRegister(reg); err != nil {
		log.Fatalf("err: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterUserVerifyServer(s, &server{})
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
