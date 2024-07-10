package main

import (
	"github.com/AswinJose14/Voting-system/server"
	"github.com/AswinJose14/Voting-system/utils"
)

func main() {
	utils.LoadEnv()
	redisClient := utils.InitializeRedisClient()
	go server.StartAuth(redisClient)
	server.StartVoteServer(redisClient)

}
