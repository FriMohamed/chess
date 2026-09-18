package server

import (
	"net/http"
	"sync"

	"chess-backend/game"
)

type Server struct {
	manager *game.Manager

	mu       sync.RWMutex
	clients  map[string]*Client
	sessions map[string]*Session
}

func New(manager *game.Manager) *Server {
	return &Server{
		manager: manager,
		clients: make(map[string]*Client),
		sessions: make(map[string]*Session),
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /games", s.openGames)
	mux.HandleFunc("POST /games/quick", s.quickGame)

	mux.HandleFunc("POST /games/private", s.createPrivateGame)
	mux.HandleFunc("POST /games/private/join", s.joinPrivateGame)

	mux.HandleFunc("DELETE /games", s.quitGame)

	mux.HandleFunc("GET /games/{gameID}/ws", s.gameWebSocket)

	return mux
}
