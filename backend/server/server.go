package server

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"chess-backend/game"
)

type Server struct {
	manager *game.Manager
}

func New(manager *game.Manager) *Server {
	return &Server{
		manager: manager,
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /games/quick", s.quickGame)
	mux.HandleFunc("GET /games/{gameID}/ws", s.gameWebSocket)

	return mux
}

func (s *Server) quickGame(w http.ResponseWriter, r *http.Request) {
	var req quickGameRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.Nickname == "" {
		http.Error(w, "nickname is required", http.StatusBadRequest)
		return
	}

	player := &game.Player{
		ID:       uuid.NewString(),
		Nickname: req.Nickname,
	}

	room := s.manager.QuickGame(player)

	response := quickGameResponse{
		GameID:   room.Game.ID,
		PlayerID: player.ID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(response)
}


