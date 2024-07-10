package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/AswinJose14/Voting-system/controllers"
	"github.com/go-redis/redis"
)

func StartVoteServer(redisClient *redis.Client) {
	//Initialising conntroller
	controller := controllers.NewVoteController(redisClient)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("Welcome")) })
	http.HandleFunc("/create", controller.CreateSession)
	http.HandleFunc("/join", controller.JoinSession)
	http.HandleFunc("/vote", controller.CastVote)
	http.HandleFunc("/results", controller.GetResults)
	fmt.Println("Listening and serving at 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
