package server

import (
	"encoding/json"

	"chess-backend/game"
)

type gameRequest struct {
	Nickname string `json:"nickname"`
}

type joinPrivateGameRequest struct {
	Nickname string `json:"nickname"`
	Code     string `json:"code"`
}

type quickGameResponse struct {
	GameID   string `json:"game_id"`
	SessionID string `json:"session_id"`
	PlayerID string `json:"player_id"`
}

type privateGameResponse struct {
    GameID    string `json:"game_id"`
    SessionID string `json:"session_id"`
    PlayerID  string `json:"player_id"`
    Code      string `json:"code"`
}

type Session struct {
	ID       string
	GameID   string
	PlayerID string
}

type MessageType string

const (
	MessageGameStarted MessageType = "game_started"
	MessageGameState   MessageType = "game_state"

	MessageMove        MessageType = "move"
	MessageResign       MessageType = "resign"
	MessageOfferDraw    MessageType = "offer_draw"
	MessageRespondDraw  MessageType = "respond_draw"

	MessageDrawOffered  MessageType = "draw_offered"
	MessageDrawDeclined MessageType = "draw_declined"
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

type DrawOfferData struct {
	PlayerID string `json:"player_id"`
}

type DrawResponseCommand struct {
	Accepted bool `json:"accepted"`
}

type ErrorCode string

const (
	ErrorInvalidMessage ErrorCode = "invalid_message"
	ErrorGameNotStarted ErrorCode = "game_not_started"
	ErrorNotYourTurn    ErrorCode = "not_your_turn"
	ErrorInvalidMove    ErrorCode = "invalid_move"
	ErrorPlayerNotFound    ErrorCode = "player_not_found"
	ErrorDrawAlreadyOffered ErrorCode = "draw_already_offered"
	ErrorNoDrawOffer        ErrorCode = "no_draw_offer"
	ErrorInvalidDrawResponse ErrorCode = "invalid_draw_response"
)

type ErrorData struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message,omitempty"`
}

type GameState struct {
	GameID     string           `json:"game_id"`
	White      *game.Player     `json:"white"`
	Black      *game.Player     `json:"black"`
	FEN        string           `json:"fen"`
	WhiteTime  int64            `json:"white_time"`
	BlackTime  int64            `json:"black_time"`
	Active     game.ActiveColor `json:"active"`

	Status     game.GameStatus  `json:"status"`
	Result     game.GameResult  `json:"result"`
	EndReason  game.EndReason   `json:"end_reason"`
	Check      bool             `json:"check"`
	DrawOfferedBy string           `json:"draw_offered_by"`
}