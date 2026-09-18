package server

import (
	"encoding/json"

	"chess-backend/game"
)

func (s *Server) handleMessage(
	room *game.Room,
	client *Client,
	message Message,
) {
	switch message.Type {
	case MessageMove:
		s.handleMove(room, client, message)

	case MessageResign:
		s.handleResign(room, client)

	case MessageOfferDraw:
		s.handleOfferDraw(room, client)

	case MessageRespondDraw:
		s.handleRespondDraw(room, client, message)

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

	if err := room.Move(client.Player.ID, notation); err != nil {
		s.sendGameError(client, err)
		return
	}

	s.broadcastState(room, MessageGameState)
	if room.Snapshot().Status == game.Finished {
		room.NotifyFinished(false)
	}
}

func (s *Server) handleResign(
	room *game.Room,
	client *Client,
) {
	if err := room.Resign(client.Player.ID); err != nil {
		s.sendGameError(client, err)
		return
	}

	s.broadcastState(room, MessageGameState)
	room.NotifyFinished(false)
}

func (s *Server) handleOfferDraw(
	room *game.Room,
	client *Client,
) {
	if err := room.OfferDraw(client.Player.ID); err != nil {
		s.sendGameError(client, err)
		return
	}

	opponent := room.Opponent(client.Player.ID)
	if opponent == nil {
		return
	}

	data, err := json.Marshal(DrawOfferData{
		PlayerID: client.Player.ID,
	})
	if err != nil {
		return
	}

	s.sendToPlayer(
		opponent.ID,
		Message{
			Type: MessageDrawOffered,
			Data: data,
		},
	)
}

func (s *Server) handleRespondDraw(
	room *game.Room,
	client *Client,
	message Message,
) {
	var command DrawResponseCommand

	if err := json.Unmarshal(message.Data, &command); err != nil {
		s.sendError(
			client,
			ErrorInvalidMessage,
			"invalid draw response",
		)
		return
	}

	offererID, err := room.RespondDraw(
		client.Player.ID,
		command.Accepted,
	)
	if err != nil {
		s.sendGameError(client, err)
		return
	}

	if command.Accepted {
		s.broadcastState(room, MessageGameState)
		room.NotifyFinished(false)
		return
	}

	s.sendToPlayer(
		offererID,
		Message{
			Type: MessageDrawDeclined,
		},
	)
}
