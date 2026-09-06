	package server

	import (
		"net/http"
		"sync"

		"chess-backend/game"
	)

	type Server struct {
		manager *game.Manager

		mu      sync.RWMutex
		clients map[string]*Client
	}

	func New(manager *game.Manager) *Server {
		return &Server{
			manager: manager,
			clients: make(map[string]*Client),
		}
	}

	func (s *Server) Handler() http.Handler {
		mux := http.NewServeMux()

		mux.HandleFunc("GET /games", s.openGames)
		mux.HandleFunc("POST /games/quick", s.quickGame)
		mux.HandleFunc("GET /games/{gameID}/ws", s.gameWebSocket)

		return mux
	}
