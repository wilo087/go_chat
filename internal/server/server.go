package server

import (
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/wilo087/go_chat/package/models"
)

type Server struct {
	subscribersMu           sync.Mutex
	subscribers             map[Subscriber]struct{}
	subscriberMessageBuffer int
	maxMessages             int
	messages                []models.Message
}

func NewServer(bufferSize int, maxMsg int) *Server {
	return &Server{
		subscribers:             make(map[Subscriber]struct{}),
		messages:                make([]models.Message, 0),
		subscriberMessageBuffer: bufferSize,
		maxMessages:             maxMsg,
	}
}

// Server starter
func (s *Server) Start() {
	s.setupRoutes()

	// Example bot messages
	// go s.broadcastBotMessage()

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Server error:", err)
		os.Exit(1)
	}
}

// Server routes
func (s *Server) setupRoutes() {
	http.Handle("/", http.FileServer(http.Dir("./htmx")))
	http.HandleFunc("/ws", s.subscribeHandler)
}

// Handler websocket connection
func (s *Server) subscribeHandler(w http.ResponseWriter, r *http.Request) {
	err := s.Subscribe(r.Context(), w, r)

	if err != nil {
		fmt.Println("WebSocket error:", err)
	}
}
