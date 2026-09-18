package game

import (
	"errors"
	"testing"
	"time"

	"github.com/corentings/chess"
)

// ------------------------------------------------------------
// Test helpers
// ------------------------------------------------------------

func newTestPlayers() (*Player, *Player) {
	return &Player{ID: "white-player"}, &Player{ID: "black-player"}
}

func newPlayingGame(t *testing.T) (*Game, *Player, *Player) {
	t.Helper()

	player1 := &Player{ID: "white-player"}
	player2 := &Player{ID: "black-player"}

	g := NewGame(player1)

	if !g.AddPlayer(player2) {
		t.Fatal("expected second player to be added")
	}

	if !g.Start() {
		t.Fatal("expected game to start")
	}

	// Start() intentionally chooses colors randomly.
	// For unit tests, make the result deterministic afterward.
	g.WhiteIndex = 0

	return g, g.White(), g.Black()
}

func TestGame_Move_AfterTimeExpired(t *testing.T) {
	g, white, _ := newPlayingGame(t)

	// Force the active player's clock to exactly zero.
	g.Clock.TimeLeft[White] = 0

	err := g.Move(white, "e2e4")
	if !errors.Is(err, ErrTimeout) {
		t.Fatalf("expected ErrTimeout, got %v", err)
	}

	if g.Status != Finished {
		t.Fatalf("expected game to be finished, got %s", g.Status)
	}

	if g.Result != BlackWins {
		t.Fatalf("expected black to win, got %s", g.Result)
	}

	if g.EndReason != Timeout {
		t.Fatalf("expected timeout end reason, got %s", g.EndReason)
	}

	// The move must NOT have been played.
	if g.Chess.Position().Turn() != chess.White {
		t.Fatal("expected turn to remain white after rejected move")
	}
}

func newPlayingRoom(t *testing.T) (*Room, *Player, *Player) {
	t.Helper()

	player1 := &Player{ID: "white-player"}
	player2 := &Player{ID: "black-player"}

	g := NewGame(player1)

	if !g.AddPlayer(player2) {
		t.Fatal("expected second player to be added")
	}

	room := NewQuickRoom(g)

	if !room.PlayerConnected(player1.ID) {
		t.Fatal("expected player1 to connect")
	}

	if !room.PlayerConnected(player2.ID) {
		t.Fatal("expected player2 to connect")
	}

	if !room.TryStart(func(bool) {}) {
		t.Fatal("expected room to start")
	}

	// TryStart calls Game.Start(), which randomly assigns colors.
	// Make the test deterministic afterward.
	g.WhiteIndex = 0

	return room, g.White(), g.Black()
}

func playMove(
	t *testing.T,
	room *Room,
	player *Player,
	move string,
) {
	t.Helper()

	if err := room.Move(player.ID, move); err != nil {
		t.Fatalf(
			"move %s by %s failed: %v",
			move,
			player.ID,
			err,
		)
	}
}

func assertGameState(
	t *testing.T,
	g *Game,
	status GameStatus,
	result GameResult,
	reason EndReason,
) {
	t.Helper()

	if g.Status != status {
		t.Fatalf(
			"expected status %q, got %q",
			status,
			g.Status,
		)
	}

	if g.Result != result {
		t.Fatalf(
			"expected result %q, got %q",
			result,
			g.Result,
		)
	}

	if g.EndReason != reason {
		t.Fatalf(
			"expected end reason %q, got %q",
			reason,
			g.EndReason,
		)
	}
}

// ------------------------------------------------------------
// Game creation / players
// ------------------------------------------------------------

func TestGame_NewGame(t *testing.T) {
	player := &Player{ID: "player-1"}

	g := NewGame(player)

	if g == nil {
		t.Fatal("expected game")
	}

	if g.ID == "" {
		t.Fatal("expected game ID")
	}

	if g.Status != Waiting {
		t.Fatalf("expected Waiting, got %s", g.Status)
	}

	if g.Result != NoResult {
		t.Fatalf("expected NoResult, got %s", g.Result)
	}

	if g.EndReason != NoEndReason {
		t.Fatalf("expected NoEndReason, got %s", g.EndReason)
	}

	if g.Players[0] != player {
		t.Fatal("expected first player")
	}

	if g.Players[1] != nil {
		t.Fatal("expected second player to be nil")
	}

	if g.Chess == nil {
		t.Fatal("expected chess game")
	}

	if g.Clock != nil {
		t.Fatal("expected clock to be nil before game starts")
	}
}

func TestGame_AddPlayer(t *testing.T) {
	white, black := newTestPlayers()

	g := NewGame(white)

	if !g.AddPlayer(black) {
		t.Fatal("expected player to be added")
	}

	if g.Players[1] != black {
		t.Fatal("expected player in second slot")
	}

	if g.IsFull() != true {
		t.Fatal("expected game to be full")
	}
}

func TestGame_AddPlayer_Nil(t *testing.T) {
	g := NewGame(&Player{ID: "player-1"})

	if g.AddPlayer(nil) {
		t.Fatal("expected nil player to be rejected")
	}
}

func TestGame_AddPlayer_Duplicate(t *testing.T) {
	player := &Player{ID: "player-1"}

	g := NewGame(player)

	if g.AddPlayer(player) {
		t.Fatal("expected duplicate player to be rejected")
	}
}

func TestGame_AddPlayer_WhenFull(t *testing.T) {
	white, black := newTestPlayers()

	g := NewGame(white)

	if !g.AddPlayer(black) {
		t.Fatal("expected black player to be added")
	}

	third := &Player{ID: "player-3"}

	if g.AddPlayer(third) {
		t.Fatal("expected third player to be rejected")
	}
}

func TestGame_GetPlayer(t *testing.T) {
	white, black := newTestPlayers()

	g := NewGame(white)
	g.AddPlayer(black)

	if g.GetPlayer(white.ID) != white {
		t.Fatal("expected white player")
	}

	if g.GetPlayer(black.ID) != black {
		t.Fatal("expected black player")
	}

	if g.GetPlayer("missing") != nil {
		t.Fatal("expected missing player to be nil")
	}
}

func TestGame_RemovePlayer(t *testing.T) {
	white, black := newTestPlayers()

	g := NewGame(white)
	g.AddPlayer(black)

	if !g.RemovePlayer(black.ID) {
		t.Fatal("expected player to be removed")
	}

	if g.GetPlayer(black.ID) != nil {
		t.Fatal("expected player to be gone")
	}

	if g.IsFull() {
		t.Fatal("expected game not to be full")
	}
}

func TestGame_RemovePlayer_Missing(t *testing.T) {
	g := NewGame(&Player{ID: "player-1"})

	if g.RemovePlayer("missing") {
		t.Fatal("expected missing player removal to fail")
	}
}

// ------------------------------------------------------------
// Game start / colors
// ------------------------------------------------------------

func TestGame_Start_RequiresTwoPlayers(t *testing.T) {
	g := NewGame(&Player{ID: "player-1"})

	if g.Start() {
		t.Fatal("expected start to fail with one player")
	}

	if g.Status != Waiting {
		t.Fatalf("expected Waiting, got %s", g.Status)
	}

	if g.Clock != nil {
		t.Fatal("expected no clock")
	}
}

func TestGame_Start(t *testing.T) {
	g, white, black := newPlayingGame(t)

	if g.Status != Playing {
		t.Fatalf("expected Playing, got %s", g.Status)
	}

	if g.Clock == nil {
		t.Fatal("expected clock")
	}

	if g.White() != white {
		t.Fatal("expected white player")
	}

	if g.Black() != black {
		t.Fatal("expected black player")
	}
}

func TestGame_Start_Twice(t *testing.T) {
	g, _, _ := newPlayingGame(t)

	clock := g.Clock

	if g.Start() {
		t.Fatal("expected second start to fail")
	}

	if g.Clock != clock {
		t.Fatal("expected original clock to remain")
	}
}

// ------------------------------------------------------------
// Normal moves
// ------------------------------------------------------------

func TestGame_Move_Success(t *testing.T) {
	g, white, black := newPlayingGame(t)

	if err := g.Move(white, "e2e4"); err != nil {
		t.Fatalf("expected move to succeed, got %v", err)
	}

	if g.Chess.Position().Turn() != chess.Black {
		t.Fatal("expected black turn")
	}

	if g.Clock.Active != Black {
		t.Fatal("expected black clock to be active")
	}

	if err := g.Move(black, "e7e5"); err != nil {
		t.Fatalf("expected move to succeed, got %v", err)
	}

	if g.Chess.Position().Turn() != chess.White {
		t.Fatal("expected white turn")
	}

	if g.Clock.Active != White {
		t.Fatal("expected white clock to be active")
	}

	assertGameState(
		t,
		g,
		Playing,
		NoResult,
		NoEndReason,
	)
}

func TestGame_Move_WrongTurn(t *testing.T) {
	g, _, black := newPlayingGame(t)

	err := g.Move(black, "e7e5")

	if !errors.Is(err, ErrNotYourTurn) {
		t.Fatalf(
			"expected ErrNotYourTurn, got %v",
			err,
		)
	}

	if g.Chess.Position().Turn() != chess.White {
		t.Fatal("expected turn to remain white")
	}

	if g.Clock.Active != White {
		t.Fatal("expected white clock to remain active")
	}
}

func TestGame_Move_InvalidMove(t *testing.T) {
	g, white, _ := newPlayingGame(t)

	err := g.Move(white, "e2e5")

	if !errors.Is(err, ErrInvalidMove) {
		t.Fatalf(
			"expected ErrInvalidMove, got %v",
			err,
		)
	}

	if g.Chess.Position().Turn() != chess.White {
		t.Fatal("expected turn to remain white")
	}

	if g.Clock.Active != White {
		t.Fatal("expected clock to remain white")
	}
}

func TestGame_Move_NilPlayer(t *testing.T) {
	g, _, _ := newPlayingGame(t)

	err := g.Move(nil, "e2e4")

	if !errors.Is(err, ErrPlayerNotFound) {
		t.Fatalf(
			"expected ErrPlayerNotFound, got %v",
			err,
		)
	}
}

func TestGame_Move_UnknownPlayer(t *testing.T) {
	g, _, _ := newPlayingGame(t)

	unknown := &Player{ID: "unknown"}

	err := g.Move(unknown, "e2e4")

	if !errors.Is(err, ErrPlayerNotFound) {
		t.Fatalf(
			"expected ErrPlayerNotFound, got %v",
			err,
		)
	}
}

func TestGame_Move_BeforeStart(t *testing.T) {
	white, black := newTestPlayers()

	g := NewGame(white)
	g.AddPlayer(black)

	err := g.Move(white, "e2e4")

	if !errors.Is(err, ErrGameNotStarted) {
		t.Fatalf(
			"expected ErrGameNotStarted, got %v",
			err,
		)
	}
}

func TestGame_Move_AfterFinished(t *testing.T) {
	g, white, black := newPlayingGame(t)

	if err := g.Resign(white); err != nil {
		t.Fatalf("expected resignation to succeed, got %v", err)
	}

	err := g.Move(black, "e7e5")

	if !errors.Is(err, ErrGameNotStarted) {
		t.Fatalf(
			"expected ErrGameNotStarted, got %v",
			err,
		)
	}
}

// ------------------------------------------------------------
// Checkmate / automatic chess outcome through YOUR Move()
// ------------------------------------------------------------

func TestGame_Move_Checkmate(t *testing.T) {
	g, white, black := newPlayingGame(t)

	// Fool's mate.
	moves := []struct {
		player *Player
		move   string
	}{
		{white, "f2f3"},
		{black, "e7e6"},
		{white, "g2g4"},
		{black, "d8h4"},
	}

	for _, m := range moves {
		if err := g.Move(m.player, m.move); err != nil {
			t.Fatalf(
				"move %s failed: %v",
				m.move,
				err,
			)
		}
	}

	assertGameState(
		t,
		g,
		Finished,
		BlackWins,
		Checkmate,
	)
}

func TestGame_Move_AfterCheckmate(t *testing.T) {
	g, white, black := newPlayingGame(t)

	g.Move(white, "f2f3")
	g.Move(black, "e7e6")
	g.Move(white, "g2g4")
	g.Move(black, "d8h4")

	err := g.Move(white, "a2a3")

	if !errors.Is(err, ErrGameNotStarted) {
		t.Fatalf(
			"expected ErrGameNotStarted, got %v",
			err,
		)
	}
}

// ------------------------------------------------------------
// Resignation
// ------------------------------------------------------------

func TestGame_Resign_White(t *testing.T) {
	g, white, _ := newPlayingGame(t)

	if err := g.Resign(white); err != nil {
		t.Fatalf("expected resignation to succeed, got %v", err)
	}

	assertGameState(
		t,
		g,
		Finished,
		BlackWins,
		Resignation,
	)
}

func TestGame_Resign_Black(t *testing.T) {
	g, _, black := newPlayingGame(t)

	if err := g.Resign(black); err != nil {
		t.Fatalf("expected resignation to succeed, got %v", err)
	}

	assertGameState(
		t,
		g,
		Finished,
		WhiteWins,
		Resignation,
	)
}

func TestGame_Resign_UnknownPlayer(t *testing.T) {
	g, _, _ := newPlayingGame(t)

	unknown := &Player{ID: "unknown"}

	err := g.Resign(unknown)

	if !errors.Is(err, ErrPlayerNotFound) {
		t.Fatalf(
			"expected ErrPlayerNotFound, got %v",
			err,
		)
	}

	assertGameState(
		t,
		g,
		Playing,
		NoResult,
		NoEndReason,
	)
}

func TestGame_Resign_AfterFinished(t *testing.T) {
	g, white, _ := newPlayingGame(t)

	g.Resign(white)

	err := g.Resign(white)

	if !errors.Is(err, ErrGameNotStarted) {
		t.Fatalf(
			"expected ErrGameNotStarted, got %v",
			err,
		)
	}
}

// ------------------------------------------------------------
// Draw offers
// ------------------------------------------------------------

func TestGame_DrawOffer(t *testing.T) {
	g, white, black := newPlayingGame(t)

	if err := g.OfferDraw(white); err != nil {
		t.Fatalf("expected draw offer to succeed, got %v", err)
	}

	if err := g.OfferDraw(black); !errors.Is(err, ErrDrawAlreadyOffered) {
		t.Fatalf(
			"expected ErrDrawAlreadyOffered, got %v",
			err,
		)
	}
}

func TestGame_DrawOffer_UnknownPlayer(t *testing.T) {
	g, _, _ := newPlayingGame(t)

	unknown := &Player{ID: "unknown"}

	err := g.OfferDraw(unknown)

	if !errors.Is(err, ErrPlayerNotFound) {
		t.Fatalf(
			"expected ErrPlayerNotFound, got %v",
			err,
		)
	}
}

func TestGame_DrawResponse_NoOffer(t *testing.T) {
	g, _, black := newPlayingGame(t)

	_, err := g.RespondDraw(black, true)

	if !errors.Is(err, ErrNoDrawOffer) {
		t.Fatalf(
			"expected ErrNoDrawOffer, got %v",
			err,
		)
	}
}

func TestGame_DrawResponse_ByOfferer(t *testing.T) {
	g, white, _ := newPlayingGame(t)

	if err := g.OfferDraw(white); err != nil {
		t.Fatalf("offer failed: %v", err)
	}

	_, err := g.RespondDraw(white, true)

	if !errors.Is(err, ErrInvalidDrawResponse) {
		t.Fatalf(
			"expected ErrInvalidDrawResponse, got %v",
			err,
		)
	}
}

func TestGame_DrawAccepted(t *testing.T) {
	g, white, black := newPlayingGame(t)

	if err := g.OfferDraw(white); err != nil {
		t.Fatalf("offer failed: %v", err)
	}

	offerer, err := g.RespondDraw(black, true)

	if err != nil {
		t.Fatalf("expected acceptance to succeed, got %v", err)
	}

	if offerer != white.ID {
		t.Fatalf(
			"expected offerer %s, got %s",
			white.ID,
			offerer,
		)
	}

	assertGameState(
		t,
		g,
		Finished,
		Draw,
		DrawAgreement,
	)
}

func TestGame_DrawDeclined(t *testing.T) {
	g, white, black := newPlayingGame(t)

	if err := g.OfferDraw(white); err != nil {
		t.Fatalf("offer failed: %v", err)
	}

	offerer, err := g.RespondDraw(black, false)

	if err != nil {
		t.Fatalf("expected decline to succeed, got %v", err)
	}

	if offerer != white.ID {
		t.Fatalf(
			"expected offerer %s, got %s",
			white.ID,
			offerer,
		)
	}

	assertGameState(
		t,
		g,
		Playing,
		NoResult,
		NoEndReason,
	)

	// The old offer must be cleared.
	if _, err := g.RespondDraw(black, false); !errors.Is(err, ErrNoDrawOffer) {
		t.Fatalf(
			"expected ErrNoDrawOffer after decline, got %v",
			err,
		)
	}
}

func TestGame_Move_ClearsDrawOffer(t *testing.T) {
	g, white, black := newPlayingGame(t)

	if err := g.OfferDraw(white); err != nil {
		t.Fatalf("offer failed: %v", err)
	}

	if err := g.Move(white, "e2e4"); err != nil {
		t.Fatalf("move failed: %v", err)
	}

	if err := g.OfferDraw(black); err != nil {
		t.Fatalf(
			"expected black to be able to offer after move, got %v",
			err,
		)
	}
}

// ------------------------------------------------------------
// Finish()
// ------------------------------------------------------------

func TestGame_Finish_WhiteDisconnect(t *testing.T) {
	g, white, _ := newPlayingGame(t)

	if !g.Finish(white.ID, EndReasonDisconnect) {
		t.Fatal("expected Finish to succeed")
	}

	assertGameState(
		t,
		g,
		Finished,
		BlackWins,
		EndReasonDisconnect,
	)
}

func TestGame_Finish_BlackDisconnect(t *testing.T) {
	g, _, black := newPlayingGame(t)

	if !g.Finish(black.ID, EndReasonDisconnect) {
		t.Fatal("expected Finish to succeed")
	}

	assertGameState(
		t,
		g,
		Finished,
		WhiteWins,
		EndReasonDisconnect,
	)
}

func TestGame_Finish_UnknownPlayer(t *testing.T) {
	g, _, _ := newPlayingGame(t)

	if g.Finish("unknown", EndReasonDisconnect) {
		t.Fatal("expected Finish to fail")
	}

	assertGameState(
		t,
		g,
		Playing,
		NoResult,
		NoEndReason,
	)
}

func TestGame_Finish_AfterFinished(t *testing.T) {
	g, white, _ := newPlayingGame(t)

	g.Resign(white)

	if g.Finish(white.ID, EndReasonDisconnect) {
		t.Fatal("expected Finish to fail")
	}

	assertGameState(
		t,
		g,
		Finished,
		BlackWins,
		Resignation,
	)
}

// ------------------------------------------------------------
// Timeout
// ------------------------------------------------------------

func TestGame_CheckTimeout_NoTimeout(t *testing.T) {
	g, _, _ := newPlayingGame(t)

	if g.CheckTimeout() {
		t.Fatal("expected no timeout")
	}

	if g.Status != Playing {
		t.Fatalf("expected Playing, got %s", g.Status)
	}
}

func TestGame_CheckTimeout_White(t *testing.T) {
	g, white, _ := newPlayingGame(t)

	g.Clock.TimeLeft[White] = time.Millisecond
	g.Clock.LastUpdate = time.Now().Add(-time.Second)

	if !g.CheckTimeout() {
		t.Fatal("expected timeout")
	}

	assertGameState(
		t,
		g,
		Finished,
		BlackWins,
		Timeout,
	)

	if g.GetPlayer(white.ID) == nil {
		t.Fatal("expected player to remain in game")
	}
}

func TestGame_CheckTimeout_Black(t *testing.T) {
	g, _, black := newPlayingGame(t)

	g.Clock.Active = Black
	g.Clock.TimeLeft[Black] = time.Millisecond
	g.Clock.LastUpdate = time.Now().Add(-time.Second)

	if !g.CheckTimeout() {
		t.Fatal("expected timeout")
	}

	assertGameState(
		t,
		g,
		Finished,
		WhiteWins,
		Timeout,
	)

	if g.GetPlayer(black.ID) == nil {
		t.Fatal("expected player to remain in game")
	}
}

func TestGame_Move_TimeoutBeforeMove(t *testing.T) {
	g, white, _ := newPlayingGame(t)

	g.Clock.TimeLeft[White] = time.Millisecond
	g.Clock.LastUpdate = time.Now().Add(-time.Second)

	err := g.Move(white, "e2e4")

	if !errors.Is(err, ErrTimeout) {
		t.Fatalf(
			"expected ErrTimeout, got %v",
			err,
		)
	}

	assertGameState(
		t,
		g,
		Finished,
		BlackWins,
		Timeout,
	)

	// The move must NOT happen.
	if g.Chess.Position().Turn() != chess.White {
		t.Fatal("expected position to remain on white turn")
	}
}

// ------------------------------------------------------------
// Room behavior
// ------------------------------------------------------------

func TestRoom_AddPlayer_OnlyWhileWaiting(t *testing.T) {
	white := &Player{ID: "white"}
	black := &Player{ID: "black"}

	g := NewGame(white)
	room := NewQuickRoom(g)

	if !room.AddPlayer(black) {
		t.Fatal("expected player to be added")
	}

	if !room.TryStart(func(bool) {}) {
		// Both players need to be connected.
		room.PlayerConnected(white.ID)
		room.PlayerConnected(black.ID)

		if !room.TryStart(func(bool) {}) {
			t.Fatal("expected game to start")
		}
	}

	third := &Player{ID: "third"}

	if room.AddPlayer(third) {
		t.Fatal("expected player to be rejected after game started")
	}
}

func TestRoom_TryStart_RequiresBothConnections(t *testing.T) {
	white, black := newTestPlayers()

	g := NewGame(white)
	g.AddPlayer(black)

	room := NewQuickRoom(g)

	room.PlayerConnected(white.ID)

	if room.TryStart(func(bool) {}) {
		t.Fatal("expected TryStart to fail while black is disconnected")
	}

	if g.Status != Waiting {
		t.Fatalf("expected Waiting, got %s", g.Status)
	}

	room.PlayerConnected(black.ID)

	if !room.TryStart(func(bool) {}) {
		t.Fatal("expected TryStart to succeed")
	}

	if g.Status != Playing {
		t.Fatalf("expected Playing, got %s", g.Status)
	}
}

func TestRoom_TryStart_RegistersCallback(t *testing.T) {
	room, _, _ := newPlayingRoom(t)

	if room.onFinished == nil {
		t.Fatal("expected finish callback")
	}
}

func TestRoom_PlayerIDs(t *testing.T) {
	white, black := newTestPlayers()

	g := NewGame(white)
	g.AddPlayer(black)

	room := NewQuickRoom(g)

	ids := room.PlayerIDs()

	if len(ids) != 2 {
		t.Fatalf("expected 2 player IDs, got %d", len(ids))
	}

	foundWhite := false
	foundBlack := false

	for _, id := range ids {
		if id == white.ID {
			foundWhite = true
		}
		if id == black.ID {
			foundBlack = true
		}
	}

	if !foundWhite || !foundBlack {
		t.Fatal("expected both player IDs")
	}
}

func TestRoom_PlayerAndOpponent(t *testing.T) {
	room, white, black := newPlayingRoom(t)

	if room.Player(white.ID) != white {
		t.Fatal("expected white player")
	}

	if room.Player(black.ID) != black {
		t.Fatal("expected black player")
	}

	if room.Opponent(white.ID) != black {
		t.Fatal("expected black as white's opponent")
	}

	if room.Opponent(black.ID) != white {
		t.Fatal("expected white as black's opponent")
	}

	if room.Player("missing") != nil {
		t.Fatal("expected missing player to be nil")
	}

	if room.Opponent("missing") != nil {
		t.Fatal("expected missing opponent to be nil")
	}
}

func TestRoom_Move(t *testing.T) {
	room, white, _ := newPlayingRoom(t)

	if err := room.Move(white.ID, "e2e4"); err != nil {
		t.Fatalf("expected move to succeed, got %v", err)
	}
}

func TestRoom_Move_UnknownPlayer(t *testing.T) {
	room, _, _ := newPlayingRoom(t)

	err := room.Move("missing", "e2e4")

	if !errors.Is(err, ErrPlayerNotFound) {
		t.Fatalf(
			"expected ErrPlayerNotFound, got %v",
			err,
		)
	}
}

func TestRoom_Resign(t *testing.T) {
	room, white, _ := newPlayingRoom(t)

	if err := room.Resign(white.ID); err != nil {
		t.Fatalf("expected resignation to succeed, got %v", err)
	}

	assertGameState(
		t,
		room.Game,
		Finished,
		BlackWins,
		Resignation,
	)
}

func TestRoom_DrawLifecycle(t *testing.T) {
	room, white, black := newPlayingRoom(t)

	if err := room.OfferDraw(white.ID); err != nil {
		t.Fatalf("offer failed: %v", err)
	}

	offerer, err := room.RespondDraw(black.ID, true)

	if err != nil {
		t.Fatalf("response failed: %v", err)
	}

	if offerer != white.ID {
		t.Fatalf("expected offerer %s, got %s", white.ID, offerer)
	}

	assertGameState(
		t,
		room.Game,
		Finished,
		Draw,
		DrawAgreement,
	)
}

// ------------------------------------------------------------
// Room finish callback
// ------------------------------------------------------------

func TestRoom_NotifyFinished_NoCallback(t *testing.T) {
	room, _, _ := newPlayingRoom(t)

	// Should simply do nothing.
	room.onFinished = nil

	room.NotifyFinished(true)
}

func TestRoom_NotifyFinished_BroadcastValue(t *testing.T) {
	room, _, _ := newPlayingRoom(t)

	called := make(chan bool, 1)

	room.onFinished = func(broadcast bool) {
		called <- broadcast
	}

	room.NotifyFinished(false)

	select {
	case value := <-called:
		if value {
			t.Fatal("expected broadcast=false")
		}
	case <-time.After(time.Second):
		t.Fatal("finish callback was not called")
	}
}

// ------------------------------------------------------------
// Room connection state
// ------------------------------------------------------------

func TestRoom_PlayerConnected_UnknownPlayer(t *testing.T) {
	room, _, _ := newPlayingRoom(t)

	if room.PlayerConnected("unknown") {
		t.Fatal("expected unknown player connection to fail")
	}
}

func TestRoom_PlayerConnected_CancelsTimer(t *testing.T) {
	white, black := newTestPlayers()

	g := NewGame(white)
	g.AddPlayer(black)

	room := NewQuickRoom(g)

	room.StartConnectionTimer(white.ID)

	if !room.PlayerConnected(white.ID) {
		t.Fatal("expected connection")
	}

	room.mu.Lock()
	_, exists := room.timers[white.ID]
	room.mu.Unlock()

	if exists {
		t.Fatal("expected connection timer to be removed")
	}
}

func TestRoom_PlayerDisconnected_CreatesTimer(t *testing.T) {
	room, white, _ := newPlayingRoom(t)

	room.PlayerDisconnected(white.ID)

	room.mu.Lock()
	_, exists := room.timers[white.ID]
	connected := room.connected[white.ID]
	room.mu.Unlock()

	if !exists {
		t.Fatal("expected reconnect timer")
	}

	if connected {
		t.Fatal("expected player to be disconnected")
	}

	// Stop the timer so this unit test doesn't leave a 30-second timer.
	room.mu.Lock()
	if timer := room.timers[white.ID]; timer != nil {
		timer.Stop()
	}
	delete(room.timers, white.ID)
	room.mu.Unlock()
}

func TestRoom_PlayerDisconnected_FinishedGame_NoTimer(t *testing.T) {
	room, white, _ := newPlayingRoom(t)

	if err := room.Resign(white.ID); err != nil {
		t.Fatalf("resign failed: %v", err)
	}

	room.PlayerDisconnected(white.ID)

	room.mu.Lock()
	_, exists := room.timers[white.ID]
	room.mu.Unlock()

	if exists {
		t.Fatal("expected no reconnect timer for finished game")
	}
}

// ------------------------------------------------------------
// Room waiting-player timeout
// ------------------------------------------------------------

func TestRoom_ConnectionTimeout_WaitingGame_RemovesPlayer(t *testing.T) {
	white, black := newTestPlayers()

	g := NewGame(white)
	g.AddPlayer(black)

	room := NewQuickRoom(g)

	timer := time.NewTimer(time.Hour)
	defer timer.Stop()

	room.mu.Lock()
	room.timers[white.ID] = timer
	room.mu.Unlock()

	room.connectionTimeout(white.ID, timer)

	if room.Game.GetPlayer(white.ID) != nil {
		t.Fatal("expected disconnected waiting player to be removed")
	}

	if room.Game.GetPlayer(black.ID) == nil {
		t.Fatal("expected other player to remain")
	}

	if g.Status != Waiting {
		t.Fatalf("expected Waiting, got %s", g.Status)
	}
}

// ------------------------------------------------------------
// Room quit
// ------------------------------------------------------------

func TestRoom_QuitGame_Waiting(t *testing.T) {
	white, black := newTestPlayers()

	g := NewGame(white)
	g.AddPlayer(black)

	room := NewQuickRoom(g)

	if !room.QuitGame(white.ID) {
		t.Fatal("expected waiting player to be removed")
	}

	if g.GetPlayer(white.ID) != nil {
		t.Fatal("expected player to be removed")
	}

	if g.GetPlayer(black.ID) == nil {
		t.Fatal("expected black to remain")
	}
}

func TestRoom_QuitGame_Playing(t *testing.T) {
	room, white, _ := newPlayingRoom(t)

	if !room.QuitGame(white.ID) {
		t.Fatal("expected quit to finish game")
	}

	assertGameState(
		t,
		room.Game,
		Finished,
		BlackWins,
		EndReasonQuit,
	)
}

func TestRoom_QuitGame_UnknownPlayer(t *testing.T) {
	room, _, _ := newPlayingRoom(t)

	if room.QuitGame("unknown") {
		t.Fatal("expected unknown player quit to fail")
	}
}

func TestRoom_QuitGame_Finished(t *testing.T) {
	room, white, _ := newPlayingRoom(t)

	room.Resign(white.ID)

	if room.QuitGame(white.ID) {
		t.Fatal("expected quit on finished game to fail")
	}
}

// ------------------------------------------------------------
// Snapshot
// ------------------------------------------------------------

func TestGame_Snapshot_BeforeStart(t *testing.T) {
	player := &Player{ID: "player-1"}

	g := NewGame(player)

	snapshot := g.Snapshot()

	if snapshot.ID != g.ID {
		t.Fatal("expected game ID")
	}

	if snapshot.Status != Waiting {
		t.Fatalf("expected Waiting, got %s", snapshot.Status)
	}

	if snapshot.Result != NoResult {
		t.Fatalf("expected NoResult, got %s", snapshot.Result)
	}

	if snapshot.EndReason != NoEndReason {
		t.Fatalf("expected NoEndReason, got %s", snapshot.EndReason)
	}

	if snapshot.White != nil {
		t.Fatal("expected no white player before game is full")
	}

	if snapshot.Black != nil {
		t.Fatal("expected no black player before game is full")
	}
}

func TestGame_Snapshot_Playing(t *testing.T) {
	g, white, black := newPlayingGame(t)

	snapshot := g.Snapshot()

	if snapshot.White != white {
		t.Fatal("expected white player in snapshot")
	}

	if snapshot.Black != black {
		t.Fatal("expected black player in snapshot")
	}

	if snapshot.Status != Playing {
		t.Fatalf("expected Playing, got %s", snapshot.Status)
	}

	if snapshot.WhiteTime <= 0 {
		t.Fatal("expected white time")
	}

	if snapshot.BlackTime <= 0 {
		t.Fatal("expected black time")
	}

	if snapshot.Active != White {
		t.Fatal("expected white to be active")
	}
}

// ------------------------------------------------------------
// GameClock
// ------------------------------------------------------------

func TestGameClock_InitialState(t *testing.T) {
	clock := newGameClock()

	if clock.Active != White {
		t.Fatalf("expected White active, got %v", clock.Active)
	}

	if clock.TimeLeft[White] != 10*time.Minute {
		t.Fatalf(
			"expected white 10 minutes, got %v",
			clock.TimeLeft[White],
		)
	}

	if clock.TimeLeft[Black] != 10*time.Minute {
		t.Fatalf(
			"expected black 10 minutes, got %v",
			clock.TimeLeft[Black],
		)
	}

	if clock.LastUpdate.IsZero() {
		t.Fatal("expected LastUpdate")
	}
}

func TestGameClock_UpdateElapsed_OnlyActivePlayer(t *testing.T) {
	clock := newGameClock()

	clock.LastUpdate = time.Now().Add(-time.Second)

	whiteBefore := clock.TimeLeft[White]
	blackBefore := clock.TimeLeft[Black]

	clock.updateElapsed()

	if clock.TimeLeft[White] >= whiteBefore {
		t.Fatal("expected white time to decrease")
	}

	if clock.TimeLeft[Black] != blackBefore {
		t.Fatal("expected black time to remain unchanged")
	}
}

func TestGameClock_UpdateElapsed_Black(t *testing.T) {
	clock := newGameClock()

	clock.Active = Black
	clock.LastUpdate = time.Now().Add(-time.Second)

	whiteBefore := clock.TimeLeft[White]
	blackBefore := clock.TimeLeft[Black]

	clock.updateElapsed()

	if clock.TimeLeft[Black] >= blackBefore {
		t.Fatal("expected black time to decrease")
	}

	if clock.TimeLeft[White] != whiteBefore {
		t.Fatal("expected white time to remain unchanged")
	}
}

func TestGameClock_NeverBelowZero(t *testing.T) {
	clock := newGameClock()

	clock.TimeLeft[White] = time.Second
	clock.LastUpdate = time.Now().Add(-time.Minute)

	clock.updateElapsed()

	if clock.TimeLeft[White] != 0 {
		t.Fatalf(
			"expected white clock to stop at zero, got %v",
			clock.TimeLeft[White],
		)
	}
}

func TestGameClock_SwitchClock(t *testing.T) {
	clock := newGameClock()

	clock.switchClock()

	if clock.Active != Black {
		t.Fatal("expected Black")
	}

	clock.switchClock()

	if clock.Active != White {
		t.Fatal("expected White")
	}
}

func TestGameClock_UpdateClock(t *testing.T) {
	clock := newGameClock()

	clock.LastUpdate = time.Now().Add(-time.Second)

	whiteBefore := clock.TimeLeft[White]

	clock.updateClock()

	if clock.TimeLeft[White] >= whiteBefore {
		t.Fatal("expected white time to decrease")
	}

	if clock.Active != Black {
		t.Fatal("expected active clock to switch to black")
	}
}
