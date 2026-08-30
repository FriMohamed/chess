package game

import (
	"errors"
	"math/rand"

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
	ID         string
	Players    [2]*Player
	WhiteIndex int
	Status     GameStatus
	Chess      *chess.Game
	Clock      *GameClock
}

func NewGame(player *Player) *Game {
	return &Game{
		ID:      uuid.NewString(),
		Players: [2]*Player{player, nil},
		Status:  Waiting,
		Chess:   chess.NewGame(),
	}
}

func (g *Game) AddPlayer(player *Player) bool {
	for i := range g.Players {
		if g.Players[i] == nil {
			g.Players[i] = player
			return true
		}
	}

	return false
}

func (g *Game) GetPlayer(id string) *Player {
	for _, player := range g.Players {
		if player != nil && player.ID == id {
			return player
		}
	}

	return nil
}

func (g *Game) RemovePlayer(id string) bool {
	for i, player := range g.Players {
		if player != nil && player.ID == id {
			g.Players[i] = nil
			return true
		}
	}

	return false
}

func (g *Game) Start() bool {
	if g.Status != Waiting || !g.IsFull() {
		return false
	}

	g.WhiteIndex = rand.Intn(2)
	g.Clock = newGameClock()
	g.Status = Playing

	return true
}



func (g *Game) White() *Player {
	if !g.IsFull() {
		return nil
	}

	return g.Players[g.WhiteIndex]
}

func (g *Game) Black() *Player {
	if !g.IsFull() {
		return nil
	}

	return g.Players[1-g.WhiteIndex]
}

func (g *Game) Move(player *Player, notation string) error {
	if g.Status != Playing {
		return ErrGameNotStarted
	}

	if !g.IsFull() {
		return ErrGameNotStarted
	}

	turn := g.Chess.Position().Turn()

	if g.White().ID == player.ID && turn != chess.White {
		return ErrNotYourTurn
	}

	if g.Black().ID == player.ID && turn != chess.Black {
		return ErrNotYourTurn
	}

	if err := g.Chess.MoveStr(notation); err != nil {
		return ErrInvalidMove
	}

	g.Clock.updateClock()

	return nil
}

func (g *Game) IsEmpty() bool {
	return g.Players[0] == nil && g.Players[1] == nil
}

func (g *Game) IsFull() bool {
	return g.Players[0] != nil && g.Players[1] != nil
}
