package controllers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/AswinJose14/Voting-system/auth"
	"github.com/AswinJose14/Voting-system/models"
	"github.com/AswinJose14/Voting-system/services"
	"github.com/AswinJose14/Voting-system/utils"
	"github.com/go-redis/redis"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type VoteController struct {
	service *services.VoteServiceImpl
}

func NewVoteController(redis *redis.Client) *VoteController {
	voteController := &VoteController{
		service: &services.VoteServiceImpl{
			Rdb:               redis,
			ClientConnections: make(map[string]*websocket.Conn),
		},
	}
	return voteController
}

var Mux = &sync.Mutex{}

func (c *VoteController) CreateSession(w http.ResponseWriter, r *http.Request) {
	sessionID := uuid.New().String()
	session, err := c.service.CreateSession(sessionID)
	if err != nil {
		http.Error(w, "Failed to create session", http.StatusInternalServerError)
		return
	}
	response := map[string]string{"session_id": session.ID}
	jsonResponse(w, response)
}

func (c *VoteController) JoinSession(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session")
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}
	username, err := auth.ValidateToken(token)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	session, ok, err := c.service.GetSession(sessionID)
	fmt.Println(session)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Error(w, "Invalid session id", http.StatusNotFound)
		return
	}
	// Upgrade the HTTP connection to a WebSocket connection
	conn, err := utils.Upgrade(w, r)
	if err != nil {
		fmt.Printf("Error upgrading to WebSocket: %v\n", err)
		http.Error(w, "Error upgrading to WebSocket", http.StatusInternalServerError)
		return
	}

	// Add the client to the session
	err = c.service.AddClient(session, username, conn)
	if err != nil {
		http.Error(w, "Failed to add client", http.StatusInternalServerError)
		return
	}

	c.service.ClientConnections[username] = conn

	// Handle messages in a separate goroutine
	go c.handleMessages(conn, session, username)
}

func (c *VoteController) CastVote(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session")

	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "missing token", http.StatusUnauthorized)
		return
	}
	
	username, err := auth.ValidateToken(token)
	if err != nil {
		http.Error(w, "invalid token", http.StatusUnauthorized)
		return
	}

	session, ok, err := c.service.GetSession(sessionID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Error(w, "Invalid session ID", http.StatusNotFound)
		return
	}
	//to check if the user had made the connection
	_, ok = session.Clients[username]
	if !ok {
		http.Error(w, "user connection not found", http.StatusNotFound)
		return
	}

	var vote models.Vote
	err = json.NewDecoder(r.Body).Decode(&vote)
	if err != nil {
		http.Error(w, "Invalid vote data", http.StatusBadRequest)
		return
	}

	success, err := c.service.AddVote(session, username, vote)
	if err != nil {
		http.Error(w, "Failed to add vote", http.StatusInternalServerError)
		return
	}
	if !success {
		http.Error(w, "Client has already voted or not found", http.StatusForbidden)
		return
	}
	c.service.BroadcastResults(session)
	jsonResponse(w, map[string]string{"status": "vote received"})

}

func (c *VoteController) GetResults(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("session")
	session, ok, err := c.service.GetSession(sessionID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
	if !ok {
		http.Error(w, "Invalid session ID", http.StatusNotFound)
		return
	}

	jsonResponse(w, session.Options)
}

func (c *VoteController) handleMessages(conn *websocket.Conn, session *models.Session, username string) {
	defer func() {
		conn.Close()
		// services.RemoveClient(session, username)
	}()

	for {
		var vote models.Vote
		err := conn.ReadJSON(&vote)
		if err != nil {
			log.Println("ReadJSON:", err)
			break
		}

		c.service.AddVote(session, username, vote)
		c.service.BroadcastResults(session)
	}
}

func jsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
