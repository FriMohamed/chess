package server

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"chess-backend/game"
)

func (s *Server) quickGame(w http.ResponseWriter, r *http.Request) {
	var request gameRequest

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

	session := s.createSession(room.Game.ID, player.ID)

	response := quickGameResponse{
		GameID:    room.Game.ID,
		SessionID: session.ID,
		PlayerID:  session.PlayerID,
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(
			w,
			"internal server error",
			http.StatusInternalServerError,
		)
		return
	}
}

func (s *Server) quitGame(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("sessionId")

	if sessionID == "" {

		http.Error(
			w,
			"game_id and session_id are required",
			http.StatusBadRequest,
		)
		return
	}

	session := s.getSession(sessionID)
	if session == nil {

		http.Error(w, "invalid session", http.StatusForbidden)
		return
	}

	removed := s.manager.RemovePlayer(session.GameID, session.PlayerID)

	if !removed {

		http.Error(w, "game or player not found", http.StatusNotFound)
		return
	}

	s.deleteSession(sessionID)

	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) createPrivateGame(w http.ResponseWriter, r *http.Request) {
	var request gameRequest

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

	room := s.manager.CreatePrivateGame(player)
	session := s.createSession(room.Game.ID, player.ID)

	response := privateGameResponse{
		GameID:    room.Game.ID,
		SessionID: session.ID,
		PlayerID:  session.PlayerID,
		Code:      room.Code,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(
			w,
			"internal server error",
			http.StatusInternalServerError,
		)
		return
	}
}

func (s *Server) joinPrivateGame(w http.ResponseWriter, r *http.Request) {
	var request joinPrivateGameRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if request.Nickname == "" || request.Code == "" {
		http.Error(
			w,
			"nickname and code are required",
			http.StatusBadRequest,
		)
		return
	}

	player := &game.Player{
		ID:       uuid.NewString(),
		Nickname: request.Nickname,
	}

	room, ok := s.manager.JoinPrivateGame(request.Code, player)
	if !ok {

		http.Error(
			w,
			"private game not found or full",
			http.StatusNotFound,
		)
		return
	}

	session := s.createSession(room.Game.ID, player.ID)

	response := privateGameResponse{
		GameID:    room.Game.ID,
		SessionID: session.ID,
		PlayerID:  session.PlayerID,
		Code:      room.Code,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(
			w,
			"internal server error",
			http.StatusInternalServerError,
		)
		return
	}
}
