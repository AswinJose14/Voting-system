package models

type Vote struct {
	Option string `json:"option"`
}

type Client struct {
	Username string
	Voted    bool
}
type Session struct {
	ID      string
	Options map[string]int
	Clients map[string]*Client
	// Mux     sync.Mutex
}
