package game_test

import (
	"testing"

	"chess-backend/game"

	"github.com/corentings/chess"
)

func TestGameAddPlayer(t *testing.T) {
	white := &game.Player{
		ID:       "white",
		Nickname: "White",
	}

	black := &game.Player{
		ID:       "black",
		Nickname: "Black",
	}

	g := game.NewGame(white)

	if g.White != white {
		t.Fatal("expected first player to be white")
	}

	if !g.AddPlayer(black) {
		t.Fatal("expected second player to be added")
	}

	if g.Black != black {
		t.Fatal("expected second player to be black")
	}

	if g.Status != game.Waiting {
		t.Fatalf("expected waiting status, got %v", g.Status)
	}
}

func TestGameDoesNotAcceptThirdPlayer(t *testing.T) {
	white := &game.Player{ID: "white"}
	black := &game.Player{ID: "black"}
	third := &game.Player{ID: "third"}

	g := game.NewGame(white)

	if !g.AddPlayer(black) {
		t.Fatal("expected black to be added")
	}

	if g.AddPlayer(third) {
		t.Fatal("expected third player to be rejected")
	}
}

func TestGameMoveBeforeStart(t *testing.T) {
	white := &game.Player{ID: "white"}

	g := game.NewGame(white)

	err := g.Move(white, "e2", "e4")

	if err != game.ErrGameNotStarted {
		t.Fatalf("expected ErrGameNotStarted, got %v", err)
	}
}

func TestGameMoveWrongTurn(t *testing.T) {
	white := &game.Player{ID: "white"}
	black := &game.Player{ID: "black"}

	g := game.NewGame(white)
	g.AddPlayer(black)
	g.Status = game.Playing

	err := g.Move(black, "e7", "e5")

	if err != game.ErrNotYourTurn {
		t.Fatalf("expected ErrNotYourTurn, got %v", err)
	}
}

func TestGameValidMove(t *testing.T) {
	white := &game.Player{ID: "white"}
	black := &game.Player{ID: "black"}

	g := game.NewGame(white)
	g.AddPlayer(black)
	g.Status = game.Playing

	err := g.Move(white, "e2", "e4")

	if err != nil {
		t.Fatalf("expected valid move, got %v", err)
	}

	if g.Chess.Position().Turn() != chess.Black {
		t.Fatal("expected turn to change to black")
	}
}

func TestGameInvalidMove(t *testing.T) {
	white := &game.Player{ID: "white"}
	black := &game.Player{ID: "black"}

	g := game.NewGame(white)
	g.AddPlayer(black)
	g.Status = game.Playing

	err := g.Move(white, "e2", "e5")

	if err != game.ErrInvalidMove {
		t.Fatalf("expected ErrInvalidMove, got %v", err)
	}
}