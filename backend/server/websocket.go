package server

import (
	"encoding/json"
	"errors"
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

	player := room.Player(playerID)
	if player == nil {
		http.Error(w, "player not found in game", http.StatusForbidden)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}	

	client := NewClient(player, conn)

	s.addClient(client)
	room.PlayerConnected(playerID)

	if len(s.roomClients(room)) == 2 && room.Start() {
		s.broadcastState(room, MessageGameStarted)
	}

	s.readClient(room, client)
}

func (s *Server) readClient(room *game.Room, client *Client) {
	defer func() {
		s.removeClient(client.Player.ID)
		room.PlayerDisconnected(client.Player.ID)
		client.Close()
	}()

	for {
		_, data, err := client.Conn.ReadMessage()
		if err != nil {
			return
		}

		var message Message

		if err := json.Unmarshal(data, &message); err != nil {
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

func (s *Server) handleMessage(
	room *game.Room,
	client *Client,
	message Message,
) {
	switch message.Type {
	case MessageMove:
		s.handleMove(room, client, message)

	default:
		s.sendError(
			client,
			ErrorInvalidMessage,
			"unknown message type",
		)
	}
}

func (s *Server) handleMove(
	room *game.Room,
	client *Client,
	message Message,
) {
	var command MoveCommand

	if err := json.Unmarshal(message.Data, &command); err != nil {
		s.sendError(
			client,
			ErrorInvalidMessage,
			"invalid move command",
		)
		return
	}

	notation := command.From + command.To + command.Promotion
	err := room.Move(client.Player.ID, notation)

	if err != nil {
		s.sendGameError(client, err)
		return
	}

	s.broadcastState(room, MessageGameState)
}

func (s *Server) sendError(
	client *Client,
	code ErrorCode,
	message string,
) {
	data, err := json.Marshal(ErrorData{
		Code:    code,
		Message: message,
	})
	if err != nil {
		return
	}

	_ = client.Send(Message{
		Type: MessageError,
		Data: data,
	})
}

func (s *Server) addClient(client *Client) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.clients[client.Player.ID] = client
}

func (s *Server) removeClient(playerID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.clients, playerID)
}

func (s *Server) roomClients(room *game.Room) []*Client {
	players := room.Players()

	s.mu.RLock()
	defer s.mu.RUnlock()

	clients := make([]*Client, 0, len(players))

	for _, player := range players {
		if player == nil {
			continue
		}

		if client := s.clients[player.ID]; client != nil {
			clients = append(clients, client)
		}
	}

	return clients
}

func (s *Server) broadcastState(
	room *game.Room,
	messageType MessageType,
) {
	state := GameState{
		GameID:    room.Game.ID,
		White:     room.Game.White(),
		Black:     room.Game.Black(),
		FEN:       room.Game.Chess.Position().String(),
		WhiteTime: room.Game.Clock.TimeLeft[game.White].Milliseconds(),
		BlackTime: room.Game.Clock.TimeLeft[game.Black].Milliseconds(),
		Active:    room.Game.Clock.Active,
	}

	data, err := json.Marshal(state)
	if err != nil {
		println("broadcast: failed to marshal state:", err.Error())
		return
	}

	message := Message{
		Type: messageType,
		Data: data,
	}

	clients := s.roomClients(room)

	for _, client := range clients {
		if err := client.Send(message); err != nil {
			println("broadcast: send failed:", err.Error())
		} else {
			println("broadcast: sent to", client.Player.ID)
		}
	}
}

func (s *Server) sendGameError(client *Client, err error) {
	switch {
	case errors.Is(err, game.ErrGameNotStarted):
		s.sendError(client, ErrorGameNotStarted, err.Error())

	case errors.Is(err, game.ErrNotYourTurn):
		s.sendError(client, ErrorNotYourTurn, err.Error())

	case errors.Is(err, game.ErrInvalidMove):
		s.sendError(client, ErrorInvalidMove, err.Error())

	case errors.Is(err, game.ErrPlayerNotFound):
		s.sendError(client, ErrorInvalidMessage, err.Error())

	default:
		s.sendError(
			client,
			ErrorInvalidMessage,
			"game operation failed",
		)
	}
}
