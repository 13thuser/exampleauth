package main

import (
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"

	datastore "github.com/13thuser/exampleauth/datastore"
	pb "github.com/13thuser/exampleauth/grpc"
)

var GRPC_SERVER_PORT = "50051"

func main() {
	// Create a new gRPC server with an interceptor
	server := grpc.NewServer(
		grpc.UnaryInterceptor(validateTokenUnaryInterceptor),
		grpc.StreamInterceptor(validateTokenStreamInterceptor),
	)

	// Create a new instance of the datastore
	db := datastore.NewDatastore()

	// Register the gRPC server
	pb.RegisterBookingServiceServer(server, NewBookingServer(db))

	// Start the gRPC server
	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", GRPC_SERVER_PORT))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	log.Println("Server started on port", GRPC_SERVER_PORT)
	// Start serving requests
	if err := server.Serve(listener); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
