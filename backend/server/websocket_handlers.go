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

    // Broadcast the updated state (containing draw_offered_by) to both players
    s.broadcastState(room, MessageGameState)
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

    err := room.RespondDraw(
        client.Player.ID,
        command.Accepted,
    )
    if err != nil {
        s.sendGameError(client, err)
        return
    }

    // 1. Broadcast the state change immediately to both clients:
    // - If accepted: Status = Finished, Result = Draw, EndReason = DrawAgreement, draw_offered_by = ""
    // - If declined: Status = Playing, draw_offered_by = ""
    s.broadcastState(room, MessageGameState)

    // 2. If accepted, trigger room completion tasks (stop timers, persist game outcome, etc.)
    if command.Accepted {
        room.NotifyFinished(false)
    }
}