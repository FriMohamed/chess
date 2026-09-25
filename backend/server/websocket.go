package server

import (
	"encoding/json"
	"log"
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
	sessionID := r.URL.Query().Get("sessionId")

	session := s.getSession(sessionID)
	if session == nil {
		http.Error(w, "invalid session", http.StatusForbidden)
		return
	}

	if session.GameID != gameID {
		http.Error(w, "invalid session", http.StatusForbidden)
		return
	}

	room := s.manager.GetRoom(gameID)
	if room == nil {
		http.Error(w, "game not found", http.StatusNotFound)
		return
	}

	playerID := session.PlayerID

	player := room.Player(playerID)
	if player == nil {
		http.Error(w, "player not found in game", http.StatusForbidden)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf(
			"[WS] upgrade failed game=%s player=%s: %v",
			gameID,
			playerID,
			err,
		)
		return
	}

	client := NewClient(player, conn)

	oldClient := s.addClient(client)

	if oldClient != nil {
		_ = oldClient.Close()
	}

	if !room.PlayerConnected(playerID) {
		s.removeClient(client)
		_ = client.Close()
		return
	}

	cleanUpFn := func(broadcast bool) {
		if broadcast {
			s.broadcastState(room, MessageGameState)
		}
		s.manager.ScheduleRoomCleanup(room.Game.ID)
	}

	if room.TryStart(cleanUpFn) {
		room.StartClockWatcher()
		s.broadcastState(room, MessageGameStarted)
	} else {
		s.broadcastState(room, MessageGameState)
	}

	s.readClient(room, client)
}

func (s *Server) readClient(room *game.Room, client *Client) {
	defer func() {
		if s.isCurrentClient(client) {
			log.Printf(
				"[WS] active connection closed game=%s player=%s",
				room.Game.ID,
				client.Player.ID,
			)

			s.removeClient(client)
			room.PlayerDisconnected(client.Player.ID)
		} else {
			log.Printf(
				"[WS] replaced connection closed game=%s player=%s",
				room.Game.ID,
				client.Player.ID,
			)
		}

		_ = client.Close()
	}()

	for {
		_, data, err := client.Conn.ReadMessage()
		if err != nil {
			log.Printf(
				"[WS] read failed game=%s player=%s: %v",
				room.Game.ID,
				client.Player.ID,
				err,
			)
			return
		}

		var message Message

		if err := json.Unmarshal(data, &message); err != nil {
			log.Printf(
				"[WS] invalid message game=%s player=%s: %v",
				room.Game.ID,
				client.Player.ID,
				err,
			)

			s.sendError(
				client,
				ErrorInvalidMessage,
				"invalid message",
			)
			continue
		}

		s.handleMessage(room, client, message)
	}
}

func (s *Server) addClient(client *Client) *Client {
	s.mu.Lock()
	defer s.mu.Unlock()

	oldClient := s.clients[client.Player.ID]

	s.clients[client.Player.ID] = client

	if oldClient != nil {
		log.Printf(
			"[WS] client replaced player=%s",
			client.Player.ID,
		)
	} else {
		log.Printf(
			"[WS] client registered player=%s",
			client.Player.ID,
		)
	}

	return oldClient
}

func (s *Server) isCurrentClient(client *Client) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.clients[client.Player.ID] == client
}

func (s *Server) removeClient(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	current, ok := s.clients[client.Player.ID]

	if !ok {
		log.Printf(
			"[WS] client already removed player=%s",
			client.Player.ID,
		)
		return
	}

	if current != client {
		log.Printf(
			"[WS] client removal skipped player=%s reason=connection_replaced",
			client.Player.ID,
		)
		return
	}

	delete(s.clients, client.Player.ID)

	log.Printf(
		"[WS] client removed player=%s",
		client.Player.ID,
	)
}

func (s *Server) roomClients(room *game.Room) []*Client {
	playerIDs := room.PlayerIDs()

	s.mu.RLock()
	defer s.mu.RUnlock()

	clients := make([]*Client, 0, len(playerIDs))

	for _, playerID := range playerIDs {
		client := s.clients[playerID]

		if client != nil {
			clients = append(clients, client)
		}
	}

	return clients
}
