package server

import (
	"encoding/json"
	"errors"

	"chess-backend/game"
)

func (s *Server) sendToPlayer(playerID string, message Message) {
	s.mu.RLock()
	client := s.clients[playerID]
	s.mu.RUnlock()

	if client == nil {
		return
	}

	if err := client.Send(message); err != nil {
		logger.Printf(
			"[WS] failed to send message player=%s type=%s: %v",
			playerID,
			message.Type,
			err,
		)
	}
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

func (s *Server) sendGameError(client *Client, err error) {
	switch {
	case errors.Is(err, game.ErrGameNotStarted):
		s.sendError(client, ErrorGameNotStarted, err.Error())

	case errors.Is(err, game.ErrNotYourTurn):
		s.sendError(client, ErrorNotYourTurn, err.Error())

	case errors.Is(err, game.ErrInvalidMove):
		s.sendError(client, ErrorInvalidMove, err.Error())

	case errors.Is(err, game.ErrPlayerNotFound):
		s.sendError(client, ErrorPlayerNotFound, err.Error())

	case errors.Is(err, game.ErrDrawAlreadyOffered):
		s.sendError(client, ErrorDrawAlreadyOffered, err.Error())

	case errors.Is(err, game.ErrNoDrawOffer):
		s.sendError(client, ErrorNoDrawOffer, err.Error())

	case errors.Is(err, game.ErrInvalidDrawResponse):
		s.sendError(client, ErrorInvalidDrawResponse, err.Error())

	default:
		s.sendError(
			client,
			ErrorInvalidMessage,
			"game operation failed",
		)
	}
}

func (s *Server) broadcastState(room *game.Room, messageType MessageType) {
	snapshot := room.Snapshot()

	state := GameState{
		GameID:    snapshot.ID,
		White:     snapshot.White,
		Black:     snapshot.Black,
		FEN:       snapshot.FEN,
		WhiteTime: snapshot.WhiteTime,
		BlackTime: snapshot.BlackTime,
		Active:    snapshot.Active,
		Status:    snapshot.Status,
		Result:    snapshot.Result,
		EndReason: snapshot.EndReason,
		Check:     snapshot.Check,
		DrawOfferedBy: snapshot.DrawOfferedBy,
	}

	

	data, err := json.Marshal(state)
	if err != nil {
		logger.Printf("[ERROR] failed to marshal game state game=%s: %v", snapshot.ID, err)
		return
	}

	message := Message{
		Type: messageType,
		Data: data,
	}

	clients := s.roomClients(room)

	for _, client := range clients {
		if err := client.Send(message); err != nil {
			logger.Printf(
				"[ERROR] websocket send failed game=%s player=%s: %v",
				snapshot.ID,
				client.Player.ID,
				err,
			)
		}
	}
}
