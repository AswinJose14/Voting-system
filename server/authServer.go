package server

import (
	"fmt"
	"log"
	"net"

	pb "github.com/AswinJose14/Voting-system/auth"
	"github.com/AswinJose14/Voting-system/services"
	"github.com/go-redis/redis"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func StartAuth(redisClient *redis.Client) {
	// Set up gRPC server
	lis, err := net.Listen("tcp", "127.0.0.1:50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	fmt.Println("Listening on port 50051")

	s := grpc.NewServer()
	pb.RegisterUserServiceServer(s, &services.Server{RedisClient: redisClient})
	fmt.Println("UserServiceServer registered")

	reflection.Register(s)
	fmt.Println("Reflection registered")

	// Start serving
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
	fmt.Println("Listening and serving at port 50051")

}
