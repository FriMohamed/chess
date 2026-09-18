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
	ErrPlayerNotFound      = errors.New("player not found")
	ErrGameNotStarted      = errors.New("game not started")
	ErrNotYourTurn         = errors.New("not your turn")
	ErrInvalidMove         = errors.New("invalid move")
	ErrTimeout             = errors.New("time expired")
	ErrDrawAlreadyOffered  = errors.New("draw already offered")
	ErrNoDrawOffer         = errors.New("no draw offer")
	ErrInvalidDrawResponse = errors.New("invalid draw response")
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
	Stalemate             EndReason = "stalemate"
	ThreefoldRepetition  EndReason = "threefold_repetition"
	FivefoldRepetition   EndReason = "fivefold_repetition"
	FiftyMoveRule        EndReason = "fifty_move_rule"
	SeventyFiveMoveRule  EndReason = "seventy_five_move_rule"
	InsufficientMaterial EndReason = "insufficient_material"
	Resignation          EndReason = "resignation"
	DrawAgreement        EndReason = "draw_agreement"
	EndReasonDisconnect  EndReason = "disconnect"
	EndReasonQuit        EndReason = "quit"
	Timeout              EndReason = "timeout"
)

type Game struct {
	ID         string
	Players    [2]*Player
	WhiteIndex int
	Status     GameStatus
	Result     GameResult
	EndReason  EndReason

	Chess *chess.Game
	Clock *GameClock

	drawOfferPlayerID string
}

func NewGame(player *Player) *Game {
	return &Game{
		ID:        uuid.NewString(),
		Players:   [2]*Player{player, nil},
		Status:    Waiting,
		Result:    NoResult,
		EndReason: NoEndReason,
		Chess:     chess.NewGame(chess.UseNotation(chess.UCINotation{})),
	}
}

func (g *Game) AddPlayer(player *Player) bool {
	if player == nil {
		return false
	}

	// Do not add the same player twice.
	if g.GetPlayer(player.ID) != nil {
		return false
	}

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

func (g *Game) Snapshot() GameSnapshot {
	var white, black *Player

	if g.IsFull() {
		white = g.White()
		black = g.Black()
	}

	var whiteTime, blackTime int64
	var active ActiveColor

	if g.Clock != nil {
		whiteTime = g.Clock.TimeLeft[White].Milliseconds()
		blackTime = g.Clock.TimeLeft[Black].Milliseconds()
		active = g.Clock.Active
	}

	return GameSnapshot{
		ID:        g.ID,
		White:     white,
		Black:     black,
		FEN:       g.Chess.FEN(),
		WhiteTime: whiteTime,
		BlackTime: blackTime,
		Active:    active,
		Status:    g.Status,
		Result:    g.Result,
		EndReason: g.EndReason,
	}
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

	if player == nil || !g.IsFull() {
		return ErrPlayerNotFound
	}

	if g.Clock == nil {
		return ErrGameNotStarted
	}

	white := g.White()
	black := g.Black()

	if white == nil || black == nil {
		return ErrPlayerNotFound
	}

	// The clock is authoritative when a move arrives.
	// This prevents a player from moving after their time has expired
	// but before the clock watcher ticked.
	g.Clock.updateElapsed()

	if g.Clock.TimeLeft[g.Clock.Active] <= 0 {
		var playerID string

		if g.Clock.Active == White {
			playerID = white.ID
		} else {
			playerID = black.ID
		}

		g.Finish(playerID, Timeout)
		return ErrTimeout
	}

	turn := g.Chess.Position().Turn()

	if player.ID == white.ID && turn != chess.White {
		return ErrNotYourTurn
	}

	if player.ID == black.ID && turn != chess.Black {
		return ErrNotYourTurn
	}

	if player.ID != white.ID && player.ID != black.ID {
		return ErrPlayerNotFound
	}

	if err := g.Chess.MoveStr(notation); err != nil {
		return ErrInvalidMove
	}

	// Only switch the clock after a successful move.
	g.Clock.switchClock()

	// A move invalidates any pending draw offer.
	g.drawOfferPlayerID = ""

	g.updateGameState()

	return nil
}

func (g *Game) Resign(player *Player) error {
	if g.Status != Playing {
		return ErrGameNotStarted
	}

	if player == nil || !g.IsFull() {
		return ErrPlayerNotFound
	}

	white := g.White()
	black := g.Black()

	if white == nil || black == nil {
		return ErrPlayerNotFound
	}

	if player.ID != white.ID && player.ID != black.ID {
		return ErrPlayerNotFound
	}

	g.Status = Finished

	if player.ID == white.ID {
		g.Result = BlackWins
	} else {
		g.Result = WhiteWins
	}

	g.EndReason = Resignation
	g.drawOfferPlayerID = ""

	return nil
}

func (g *Game) OfferDraw(player *Player) error {
	if g.Status != Playing {
		return ErrGameNotStarted
	}

	if player == nil || !g.IsFull() {
		return ErrPlayerNotFound
	}

	white := g.White()
	black := g.Black()

	if white == nil || black == nil {
		return ErrPlayerNotFound
	}

	if player.ID != white.ID && player.ID != black.ID {
		return ErrPlayerNotFound
	}

	if g.drawOfferPlayerID != "" {
		return ErrDrawAlreadyOffered
	}

	g.drawOfferPlayerID = player.ID

	return nil
}

func (g *Game) RespondDraw(player *Player, accepted bool) (string, error) {
	if g.Status != Playing {
		return "", ErrGameNotStarted
	}

	if player == nil || !g.IsFull() {
		return "", ErrPlayerNotFound
	}

	white := g.White()
	black := g.Black()

	if white == nil || black == nil {
		return "", ErrPlayerNotFound
	}

	if player.ID != white.ID && player.ID != black.ID {
		return "", ErrPlayerNotFound
	}

	if g.drawOfferPlayerID == "" {
		return "", ErrNoDrawOffer
	}

	if g.drawOfferPlayerID == player.ID {
		return "", ErrInvalidDrawResponse
	}

	offererID := g.drawOfferPlayerID

	if accepted {
		g.Status = Finished
		g.Result = Draw
		g.EndReason = DrawAgreement
	}

	g.drawOfferPlayerID = ""

	return offererID, nil
}

func (g *Game) updateGameState() {
	g.Result = NoResult
	g.EndReason = NoEndReason

	switch g.Chess.Outcome() {
	case chess.WhiteWon:
		g.Status = Finished
		g.Result = WhiteWins
		g.EndReason = mapEndReason(g.Chess.Method())

	case chess.BlackWon:
		g.Status = Finished
		g.Result = BlackWins
		g.EndReason = mapEndReason(g.Chess.Method())

	case chess.Draw:
		g.Status = Finished
		g.Result = Draw
		g.EndReason = mapEndReason(g.Chess.Method())
	}
}

func (g *Game) Finish(playerID string, reason EndReason) bool {
	if g.Status != Playing || !g.IsFull() {
		return false
	}

	player := g.GetPlayer(playerID)
	if player == nil {
		return false
	}

	white := g.White()
	black := g.Black()

	if white == nil || black == nil {
		return false
	}

	if player.ID == white.ID {
		g.Result = BlackWins
	} else if player.ID == black.ID {
		g.Result = WhiteWins
	} else {
		return false
	}

	g.Status = Finished
	g.EndReason = reason
	g.drawOfferPlayerID = ""

	return true
}

func (g *Game) CheckTimeout() bool {
	if g.Status != Playing || g.Clock == nil || !g.IsFull() {
		return false
	}

	g.Clock.updateElapsed()

	if g.Clock.TimeLeft[g.Clock.Active] > 0 {
		return false
	}

	var playerID string

	if g.Clock.Active == White {
		playerID = g.White().ID
	} else {
		playerID = g.Black().ID
	}

	return g.Finish(playerID, Timeout)
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