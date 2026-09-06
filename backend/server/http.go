package server

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"chess-backend/game"
)

func (s *Server) quickGame(w http.ResponseWriter, r *http.Request) {
	var request quickGameRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if request.Nickname == "" {
		http.Error(w, "nickname is required", http.StatusBadRequest)
		return
	}

	player := &game.Player{
		ID:       uuid.NewString(),
		Nickname: request.Nickname,
	}

	room := s.manager.QuickGame(player)

	response := quickGameResponse{
		GameID:   room.GameID(),
		PlayerID: player.ID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (s *Server) openGames(w http.ResponseWriter, r *http.Request) {
    games := s.manager.OpenGames()

    w.Header().Set("Content-Type", "application/json")

    json.NewEncoder(w).Encode(games)
}


