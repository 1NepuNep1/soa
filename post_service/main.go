package main

import (
	"log"
	"net"
	"os"
	"postservice/database"
	"postservice/handlers"
	"postservice/models"

	pb "postservice/proto"

	"google.golang.org/grpc"
)

func main() {
	database.ConnectDB()
	database.DB.AutoMigrate(&models.Post{})

	lis, err := net.Listen("tcp", ":"+os.Getenv("GRPC_PORT"))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()

	pb.RegisterPostServiceServer(s, &handlers.Server{})

	log.Printf("gRPC server running on port :%s", os.Getenv("GRPC_PORT"))
	if err := s.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
