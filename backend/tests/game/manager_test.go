package game_test

import (
	"testing"

	"chess-backend/game"
)

func TestManagerQuickGame(t *testing.T) {
	manager := game.NewManager()

	player1 := &game.Player{ID: "player-1"}
	player2 := &game.Player{ID: "player-2"}

	room1 := manager.QuickGame(player1)
	room2 := manager.QuickGame(player2)

	if room1 != room2 {
		t.Fatal("expected both players to be placed in the same room")
	}

	if room1.Game.White != player1 {
		t.Fatal("expected player1 to be white")
	}

	if room1.Game.Black != player2 {
		t.Fatal("expected player2 to be black")
	}
}

func TestManagerQuickGameCreatesNewRoomWhenFull(t *testing.T) {
	manager := game.NewManager()

	player1 := &game.Player{ID: "player-1"}
	player2 := &game.Player{ID: "player-2"}
	player3 := &game.Player{ID: "player-3"}

	room1 := manager.QuickGame(player1)
	manager.QuickGame(player2)

	room2 := manager.QuickGame(player3)

	if room1 == room2 {
		t.Fatal("expected third player to get a new room")
	}
}