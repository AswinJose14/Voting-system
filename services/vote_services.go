package services

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/AswinJose14/Voting-system/models"
	"github.com/go-redis/redis"
	"github.com/gorilla/websocket"
)

type VoteServiceImpl struct {
	Rdb               *redis.Client
	ClientConnections map[string]*websocket.Conn
}

func InitializeRedisClient() {

}
func (s *VoteServiceImpl) CreateSession(sessionID string) (*models.Session, error) {
	fmt.Println("Creating session")
	session := &models.Session{
		ID:      sessionID,
		Options: make(map[string]int),
		Clients: make(map[string]*models.Client),
	}
	err := s.SaveSession(session)
	if err != nil {
		return nil, err
	}
	return session, nil
}

func (s *VoteServiceImpl) GetSession(sessionID string) (*models.Session, bool, error) {
	session := &models.Session{}
	fmt.Println("Fetching session from Redis")

	// Get the Redis key for the session
	sessionKey := GetSessionKey(sessionID)
	fmt.Println(sessionKey)

	// Fetch the session data from Redis
	sessionData, err := s.Rdb.Get(sessionKey).Result()
	fmt.Println(sessionData)
	if err == redis.Nil {
		fmt.Println("no session found")
		return nil, false, nil
	} else if err != nil {
		fmt.Printf("Error fetching session from Redis: %v\n", err)
		return nil, false, err
	}

	// Unmarshal the session data
	err = json.Unmarshal([]byte(sessionData), session)
	if err != nil {
		return nil, false, err
	}

	return session, true, nil
}

func (s *VoteServiceImpl) SaveSession(session *models.Session) error {
	fmt.Println("saving session")
	sessionData, err := json.Marshal(session)
	if err != nil {
		return err
	}
	fmt.Println(string(sessionData))
	sessionKey := GetSessionKey(session.ID)
	err = s.Rdb.Set(sessionKey, string(sessionData), time.Hour*24).Err()
	if err != nil {
		fmt.Printf("could not set session to redis")
		return err
	}
	return nil
}

func (s *VoteServiceImpl) DeleteSession(sessionID string) error {
	err := s.Rdb.Del(sessionID).Err()
	if err != nil {
		return err
	}
	return nil
}

func (s *VoteServiceImpl) AddClient(session *models.Session, username string, conn *websocket.Conn) error {
	session.Clients[username] = &models.Client{
		Username: username,
		Voted:    false,
	}
	return s.SaveSession(session)
}

func (s *VoteServiceImpl) RemoveClient(session *models.Session, username string) error {
	delete(session.Clients, username)
	return s.SaveSession(session)
}

func (s *VoteServiceImpl) AddVote(session *models.Session, username string, vote models.Vote) (bool, error) {
	client, ok := session.Clients[username]
	if !ok || client.Voted {
		return false, nil
	}
	fmt.Println("adding vote for client:", username, " Option: ", vote.Option)
	session.Options[vote.Option]++
	client.Voted = true
	return true, s.SaveSession(session)
}

func (s *VoteServiceImpl) BroadcastResults(session *models.Session) {
	fmt.Println("broadcasting results")

	results := map[string]interface{}{
		"session_id": session.ID,
		"results":    session.Options,
	}

	for _, client := range session.Clients {
		// For simplicity, we assume the clients have a Conn field (you can store the actual connections elsewhere)
		if clientConn, ok := s.ClientConnections[client.Username]; ok {
			err := clientConn.WriteJSON(results)
			if err != nil {
				clientConn.Close()
				delete(s.ClientConnections, client.Username)
			}
		}
	}
}

func GetSessionKey(sessionID string) string {
	sessionKey := "sessionData:" + sessionID
	return sessionKey
}
