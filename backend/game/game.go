package game

import (
	"errors"

	"github.com/corentings/chess"
	"github.com/google/uuid"
)

type GameStatus string

const (
	Waiting  GameStatus = "waiting"
	Playing  GameStatus = "playing"
	Finished GameStatus = "finished"
)

var (
	ErrPlayerNotFound = errors.New("player not found")
	ErrGameNotStarted = errors.New("game not started")
	ErrNotYourTurn    = errors.New("not your turn")
	ErrInvalidMove    = errors.New("invalid move")
)

type Game struct {
	ID     string
	White  *Player
	Black  *Player
	Status GameStatus
	Chess  *chess.Game
}

func NewGame(player *Player) *Game {
	return &Game{
		ID:     uuid.NewString(),
		White:  player,
		Status: Waiting,
		Chess:  chess.NewGame(),
	}
}

func (g *Game) AddPlayer(player *Player) bool {
	if g.Black != nil {
		return false
	}

	g.Black = player
	return true
}

func (g *Game) GetPlayer(id string) *Player {
	if g.White != nil && g.White.ID == id {
		return g.White
	}

	if g.Black != nil && g.Black.ID == id {
		return g.Black
	}

	return nil
}

func (g *Game) RemovePlayer(id string) bool {
	if g.White != nil && g.White.ID == id {
		g.White = nil
		return true
	}

	if g.Black != nil && g.Black.ID == id {
		g.Black = nil
		return true
	}

	return false
}

func (g *Game) Move(player *Player, from, to string) error {
	if g.Status != Playing {
		return ErrGameNotStarted
	}

	if !g.IsFull() {
		return ErrGameNotStarted
	}

	turn := g.Chess.Position().Turn()

	if g.White.ID == player.ID && turn != chess.White {
		return ErrNotYourTurn
	}

	if g.Black.ID == player.ID && turn != chess.Black {
		return ErrNotYourTurn
	}

	if err := g.Chess.MoveStr(from + to); err != nil {
		return ErrInvalidMove
	}

	return nil
}

func (g *Game) IsEmpty() bool {
	return g.White == nil && g.Black == nil
}

func (g *Game) IsFull() bool {
	return g.White != nil && g.Black != nil
}
