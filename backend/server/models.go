package server

import (
	"encoding/json"

	"chess-backend/game"
)

type quickGameRequest struct {
	Nickname string `json:"nickname"`
}

type quickGameResponse struct {
	GameID   string `json:"game_id"`
	PlayerID string `json:"player_id"`
}

type MessageType string

const (
	MessageGameStarted MessageType = "game_started"
	MessageGameState   MessageType = "game_state"
	MessageMove        MessageType = "move"
	MessageError       MessageType = "error"
)

type Message struct {
	Type MessageType     `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

type MoveCommand struct {
	From      string `json:"from"`
	To        string `json:"to"`
	Promotion string `json:"promotion,omitempty"`
}

type ErrorCode string

const (
	ErrorInvalidMessage ErrorCode = "invalid_message"
	ErrorGameNotStarted ErrorCode = "game_not_started"
	ErrorNotYourTurn    ErrorCode = "not_your_turn"
	ErrorInvalidMove    ErrorCode = "invalid_move"
)

type ErrorData struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message,omitempty"`
}

type GameState struct {
	GameID string       `json:"game_id"`
	White  *game.Player `json:"white"`
	Black  *game.Player `json:"black"`
	FEN    string       `json:"fen"`
	WhiteTime  int64           `json:"white_time"`
	BlackTime  int64           `json:"black_time"`
	Active     int             `json:"active"`
}