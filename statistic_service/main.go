package main

import (
	"log"
	"net"
	"os"
	"statisticservice/database"
	"statisticservice/handlers"
	"statisticservice/kafka"
	"statisticservice/rest"

	pb "statisticservice/proto"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

func main() {
	database.ConnectDB()
	database.EnsureSchema()

	kafka.StartConsumers()

	go func() {
		lis, err := net.Listen("tcp", os.Getenv("GRPC_SERVER_ADDR"))
		if err != nil {
			log.Fatalf("Failed to listen gRPC: %v", err)
		}

		s := grpc.NewServer()
		pb.RegisterStatisticsServiceServer(s, &handlers.Server{})
		log.Println("gRPC StatisticsService running on", os.Getenv("GRPC_SERVER_ADDR"))
		if err := s.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	rest.InitGRPCClient()

	router := gin.Default()
	rest.RegisterRoutes(router)

	log.Println("REST server running on", os.Getenv("REST_SERVER_ADDR"))
	if err := router.Run(os.Getenv("REST_SERVER_ADDR")); err != nil {
		log.Fatalf("Failed to run REST server: %v", err)
	}
}
