package server

import (
	"net/http"

	"github.com/gorilla/websocket"

	"chess-backend/game"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (s *Server) gameWebSocket(w http.ResponseWriter, r *http.Request) {
	gameID := r.PathValue("gameID")
	playerID := r.URL.Query().Get("playerId")

	room := s.manager.GetRoom(gameID)
	if room == nil {
		http.Error(w, "game not found", http.StatusNotFound)
		return
	}

	player := room.Game.GetPlayer(playerID)
	if player == nil {
		http.Error(w, "player not found in game", http.StatusForbidden)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := &game.Client{
		Player: player,
		Conn:   conn,
	}

	room.AddClient(client)

	go client.Listen(room)
}
