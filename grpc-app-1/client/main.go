package main

import (
	"log"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	pb "github.com/amitsaha/apigateway-demo/grpc-app-1/verify"
)

const (
	address     = "localhost:50051"
)

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewUserVerifyClient(conn)

    r, err := c.VerifyUser(context.Background(), &pb.VerifyRequest{Id: 12321, Token: "$kasdasa"})
	if err != nil {
		log.Fatalf("Could not verify: %v", err)
	}
	log.Printf("Result: %s", r.Message)
}
