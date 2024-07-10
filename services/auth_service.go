package services

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"time"

	pb "github.com/AswinJose14/Voting-system/auth"
	"github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis"
)

type Server struct {
	pb.UnimplementedUserServiceServer
	RedisClient *redis.Client
}

func (s *Server) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	hashedPassword := fmt.Sprintf("%x", sha256.Sum256([]byte(req.GetPassword())))
	err := s.RedisClient.Set(req.GetUsername(), hashedPassword, 0).Err()
	if err != nil {
		return nil, err
	}

	return &pb.RegisterResponse{Message: "User registered successfully"}, nil
}

func (s *Server) Login(ctx context.Context, req *pb.LoginRequest) (*pb.LoginResponse, error) {
	hashedPassword := fmt.Sprintf("%x", sha256.Sum256([]byte(req.GetPassword())))
	storedPassword, err := s.RedisClient.Get(req.GetUsername()).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("user does not exist")
	} else if err != nil {
		return nil, err
	}

	if storedPassword != hashedPassword {
		return nil, fmt.Errorf("invalid credentials")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": req.GetUsername(),
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(os.Getenv("JWT_SECRET_KEY")))
	if err != nil {
		return nil, fmt.Errorf("could not generate token")
	}

	return &pb.LoginResponse{Token: tokenString}, nil
}

func (s *Server) Logout(ctx context.Context, req *pb.LogoutRequest) (*pb.LogoutResponse, error) {
	return &pb.LogoutResponse{Message: "User logged out successfully"}, nil
}
