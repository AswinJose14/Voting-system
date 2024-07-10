package main

import (
	"github.com/AswinJose14/Voting-system/server"
	"github.com/AswinJose14/Voting-system/utils"
)

func main() {
	//Load env files
	utils.LoadEnv()
	//Initialise Redis
	redisClient := utils.InitializeRedisClient()
	//Starting gRPC server for User Authentication
	go server.StartAuth(redisClient)
	//Starting server for the voting system
	server.StartVoteServer(redisClient)
}
