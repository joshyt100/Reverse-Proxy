package main

import (
	"context"
	"log"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"reverse-proxy/testgrpc/echopb"
)

type server struct {
	echopb.UnimplementedEchoServiceServer
}

func (s *server) Echo(ctx context.Context, req *echopb.EchoRequest) (*echopb.EchoResponse, error) {
	log.Printf("received: %q", req.GetMessage())
	return &echopb.EchoResponse{
		Message: "echo: " + req.GetMessage(),
	}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}

	grpcServer := grpc.NewServer()
	echopb.RegisterEchoServiceServer(grpcServer, &server{})
	reflection.Register(grpcServer)

	log.Println("gRPC backend listening on :50051")
	log.Fatal(grpcServer.Serve(lis))
}
