package game

import (
	"errors"
	"fmt"
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

type GameResult string

const (
	NoResult  GameResult = "none"
	WhiteWins GameResult = "white_wins"
	BlackWins GameResult = "black_wins"
	Draw      GameResult = "draw"
)

type EndReason string

const (
	NoEndReason          EndReason = "none"
	Checkmate            EndReason = "checkmate"
	Stalemate            EndReason = "stalemate"
	ThreefoldRepetition  EndReason = "threefold_repetition"
	FivefoldRepetition   EndReason = "fivefold_repetition"
	FiftyMoveRule        EndReason = "fifty_move_rule"
	SeventyFiveMoveRule  EndReason = "seventy_five_move_rule"
	InsufficientMaterial EndReason = "insufficient_material"
	Resignation          EndReason = "resignation"
	Timeout              EndReason = "timeout"
)

type Game struct {
	ID         string
	Players    [2]*Player
	WhiteIndex int
	Status     GameStatus
	Result     GameResult
	EndReason  EndReason
	Check      bool

	Chess *chess.Game
	Clock *GameClock
}

func NewGame(player *Player) *Game {
	return &Game{
		ID:        uuid.NewString(),
		Players:   [2]*Player{player, nil},
		Status:    Waiting,
		Result:    NoResult,
		EndReason: NoEndReason,
		Check:     false,
		Chess:     chess.NewGame(chess.UseNotation(chess.UCINotation{})),
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

	fmt.Printf("MOVE STRING RECEIVED: %q\n", notation)

	if err := g.Chess.MoveStr(notation); err != nil {
		return ErrInvalidMove
	}

	g.Clock.updateClock()
	g.updateGameState()

	return nil
}

func (g *Game) updateGameState() {
	g.Check = false

	outcome := g.Chess.Outcome()

	switch outcome {
	case chess.WhiteWon:
		g.Status = Finished
		g.Result = WhiteWins

	case chess.BlackWon:
		g.Status = Finished
		g.Result = BlackWins

	case chess.Draw:
		g.Status = Finished
		g.Result = Draw

	default:
		g.Result = NoResult
	}

	if g.Status != Finished {
		g.Check = g.Chess.Position().Status() == chess.Checkmate
		return
	}

	g.EndReason = mapEndReason(g.Chess.Method())
}

func (g *Game) IsEmpty() bool {
	return g.Players[0] == nil && g.Players[1] == nil
}

func (g *Game) IsFull() bool {
	return g.Players[0] != nil && g.Players[1] != nil
}

func mapEndReason(method chess.Method) EndReason {
	switch method {
	case chess.Checkmate:
		return Checkmate

	case chess.Stalemate:
		return Stalemate

	case chess.ThreefoldRepetition:
		return ThreefoldRepetition

	case chess.FivefoldRepetition:
		return FivefoldRepetition

	case chess.FiftyMoveRule:
		return FiftyMoveRule

	case chess.SeventyFiveMoveRule:
		return SeventyFiveMoveRule

	case chess.InsufficientMaterial:
		return InsufficientMaterial

	case chess.Resignation:
		return Resignation

	default:
		return NoEndReason
	}
}
