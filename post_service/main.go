package main

import (
	"log"
	"net"
	"os"
	"postservice/database"
	"postservice/handlers"
	"postservice/models"
	"postservice/rest"

	pb "postservice/proto"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

func main() {
	database.ConnectDB()
	database.DB.AutoMigrate(&models.Post{})

	go func() {
		lis, err := net.Listen("tcp", os.Getenv("GRPC_SERVER_ADDR"))
		if err != nil {
			log.Fatalf("Failed to listen gRPC: %v", err)
		}

		s := grpc.NewServer()
		pb.RegisterPostServiceServer(s, &handlers.Server{})
		log.Println("gRPC server running on", os.Getenv("GRPC_SERVER_ADDR"))
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	rest.InitGRPCClient()
	router := gin.Default()
	rest.RegisterRoutes(router)

	log.Println("REST server running on", os.Getenv("REST_SERVER_ADDR"))
	router.Run(os.Getenv("REST_SERVER_ADDR"))
}
