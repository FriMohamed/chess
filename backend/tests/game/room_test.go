package game_test

import (
	"testing"

	"chess-backend/game"
)

func TestRoomPlayers(t *testing.T) {
	white := &game.Player{ID: "white"}
	black := &game.Player{ID: "black"}

	room := game.NewRoom(game.NewGame(white))

	room.AddPlayer(black)

	players := room.Players()

	if len(players) != 2 {
		t.Fatalf("expected 2 players, got %d", len(players))
	}

	if players[0] != white {
		t.Fatal("expected white as first player")
	}

	if players[1] != black {
		t.Fatal("expected black as second player")
	}
}

func TestRoomPlayer(t *testing.T) {
	white := &game.Player{ID: "white"}

	room := game.NewRoom(game.NewGame(white))

	player := room.Player("white")

	if player != white {
		t.Fatal("expected to find white player")
	}

	if room.Player("unknown") != nil {
		t.Fatal("expected unknown player to return nil")
	}
}

func TestRoomStart(t *testing.T) {
	white := &game.Player{ID: "white"}
	black := &game.Player{ID: "black"}

	room := game.NewRoom(game.NewGame(white))

	if room.Start() {
		t.Fatal("game should not start with one player")
	}

	room.AddPlayer(black)

	if !room.Start() {
		t.Fatal("game should start with two players")
	}

	if room.Game.Status != game.Playing {
		t.Fatal("expected game to be playing")
	}

	if room.Start() {
		t.Fatal("game should not start twice")
	}
}

func TestRoomMove(t *testing.T) {
	white := &game.Player{ID: "white"}
	black := &game.Player{ID: "black"}

	room := game.NewRoom(game.NewGame(white))
	room.AddPlayer(black)
	room.Start()

	err := room.Move("white", "e2", "e4")

	if err != nil {
		t.Fatalf("expected move to succeed, got %v", err)
	}
}

func TestRoomMoveUnknownPlayer(t *testing.T) {
	white := &game.Player{ID: "white"}
	black := &game.Player{ID: "black"}

	room := game.NewRoom(game.NewGame(white))
	room.AddPlayer(black)
	room.Start()

	err := room.Move("unknown", "e2", "e4")

	if err != game.ErrPlayerNotFound {
		t.Fatalf("expected ErrPlayerNotFound, got %v", err)
	}
}