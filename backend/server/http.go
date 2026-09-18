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
		logger.Printf("[HTTP] quick game invalid body: %v", err)
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if request.Nickname == "" {
		logger.Printf("[HTTP] quick game missing nickname")
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
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Printf(
			"[HTTP] quick game response failed game=%s session=%s: %v",
			room.Game.ID,
			session.ID,
			err,
		)
		return
	}

	logger.Printf(
		"[HTTP] quick game created game=%s player=%s session=%s",
		room.Game.ID,
		player.ID,
		session.ID,
	)
}



func (s *Server) quitGame(w http.ResponseWriter, r *http.Request) {
	sessionID := r.URL.Query().Get("sessionId")

	if sessionID == "" {
		logger.Printf(
			"[HTTP] quit game missing session session=%s",
			sessionID,
		)

		http.Error(
			w,
			"game_id and session_id are required",
			http.StatusBadRequest,
		)
		return
	}

	session := s.getSession(sessionID)
	if session == nil {
		logger.Printf(
			"[HTTP] quit game invalid session session=%s",
			sessionID,
		)

		http.Error(w, "invalid session", http.StatusForbidden)
		return
	}

	removed := s.manager.RemovePlayer(session.GameID, session.PlayerID)

	if !removed {
		logger.Printf(
			"[HTTP] quit game player not found game=%s player=%s session=%s",
			session.GameID,
			session.PlayerID,
			sessionID,
		)

		http.Error(w, "game or player not found", http.StatusNotFound)
		return
	}

	s.deleteSession(sessionID)

	logger.Printf(
		"[HTTP] game quit game=%s player=%s session=%s",
		session.GameID,
		session.PlayerID,
		sessionID,
	)

	w.WriteHeader(http.StatusNoContent)
}


func (s *Server) createPrivateGame(w http.ResponseWriter, r *http.Request) {
	var request gameRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		logger.Printf("[HTTP] private game invalid body: %v", err)
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
		Code:      room.Code,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Printf(
			"[HTTP] private game response failed game=%s session=%s: %v",
			room.Game.ID,
			session.ID,
			err,
		)
		return
	}

	logger.Printf(
		"[HTTP] private game created game=%s code=%s session=%s",
		room.Game.ID,
		room.Code,
		session.ID,
	)
}


func (s *Server) joinPrivateGame(w http.ResponseWriter, r *http.Request) {
	var request joinPrivateGameRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		logger.Printf("[HTTP] join private game invalid body: %v", err)
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

	room, ok := s.manager.JoinPrivateGame(request.Code, player, logger)
	if !ok {
		logger.Printf(
			"[HTTP] private game not found or full code=%s",
			request.Code,
		)

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
		Code:      room.Code,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		logger.Printf(
			"[HTTP] join private game response failed game=%s session=%s: %v",
			room.Game.ID,
			session.ID,
			err,
		)
		return
	}

	logger.Printf(
		"[HTTP] private game joined game=%s code=%s session=%s",
		room.Game.ID,
		room.Code,
		session.ID,
	)
}

// a helper for diagnostic
func (s *Server) openGames(w http.ResponseWriter, r *http.Request) {
	games := s.manager.OpenGames()

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(games)
}
