package server

import (
	"bytes"
	"encoding/json"
	"github.com/gorilla/websocket"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"chess-backend/game"
)

func newTestServer(t *testing.T) (*Server, *httptest.Server) {
	t.Helper()

	manager := game.NewManager()
	server := New(manager)

	httpServer := httptest.NewServer(server.Handler())

	t.Cleanup(func() {
		httpServer.Close()
	})

	return server, httpServer
}

func createQuickGame(
	t *testing.T,
	httpServer *httptest.Server,
	nickname string,
) quickGameResponse {
	t.Helper()

	body := bytes.NewBufferString(
		`{"nickname":"` + nickname + `"}`,
	)

	resp, err := http.Post(
		httpServer.URL+"/games/quick",
		"application/json",
		body,
	)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf(
			"expected status 200, got %d",
			resp.StatusCode,
		)
	}

	var result quickGameResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatal(err)
	}

	if result.GameID == "" {
		t.Fatal("expected game ID")
	}

	if result.SessionID == "" {
		t.Fatal("expected session ID")
	}

	return result
}

func wsURL(
	httpURL string,
	gameID string,
	sessionID string,
) string {
	u, err := url.Parse(httpURL)
	if err != nil {
		panic(err)
	}

	u.Scheme = "ws"
	u.Path = "/games/" + gameID + "/ws"

	query := u.Query()
	query.Set("sessionId", sessionID)
	u.RawQuery = query.Encode()

	return u.String()
}

func connectWS(
	t *testing.T,
	httpServer *httptest.Server,
	gameID string,
	sessionID string,
) *websocket.Conn {
	t.Helper()

	conn, resp, err := websocket.DefaultDialer.Dial(
		wsURL(httpServer.URL, gameID, sessionID),
		nil,
	)

	if err != nil {
		if resp != nil {
			resp.Body.Close()
		}

		t.Fatal(err)
	}

	return conn
}

func TestGameWebSocket_FirstPlayerConnection(t *testing.T) {
	server, httpServer := newTestServer(t)

	player := createQuickGame(t, httpServer, "Player1")

	conn := connectWS(
		t,
		httpServer,
		player.GameID,
		player.SessionID,
	)
	defer conn.Close()

	// The first player should receive the current waiting state.
	var message Message

	if err := conn.ReadJSON(&message); err != nil {
		t.Fatal(err)
	}

	if message.Type != MessageGameState {
		t.Fatalf(
			"expected %q, got %q",
			MessageGameState,
			message.Type,
		)
	}

	// The connection must be registered on the server.
	server.mu.RLock()
	client := server.clients[server.sessions[player.SessionID].PlayerID]
	server.mu.RUnlock()

	if client == nil {
		t.Fatal("expected player connection to be registered")
	}
}

func TestGameWebSocket_SecondPlayerStartsGame(t *testing.T) {
	server, httpServer := newTestServer(t)

	player1 := createQuickGame(t, httpServer, "Player1")
	player2 := createQuickGame(t, httpServer, "Player2")

	conn1 := connectWS(
		t,
		httpServer,
		player1.GameID,
		player1.SessionID,
	)
	defer conn1.Close()

	// First player receives waiting state.
	var message Message

	if err := conn1.ReadJSON(&message); err != nil {
		t.Fatal(err)
	}

	if message.Type != MessageGameState {
		t.Fatalf(
			"player 1: expected %q, got %q",
			MessageGameState,
			message.Type,
		)
	}

	conn2 := connectWS(
		t,
		httpServer,
		player2.GameID,
		player2.SessionID,
	)
	defer conn2.Close()

	// Both players should receive game_started.
	if err := conn1.ReadJSON(&message); err != nil {
		t.Fatal(err)
	}

	if message.Type != MessageGameStarted {
		t.Fatalf(
			"player 1: expected %q, got %q",
			MessageGameStarted,
			message.Type,
		)
	}

	if err := conn2.ReadJSON(&message); err != nil {
		t.Fatal(err)
	}

	if message.Type != MessageGameStarted {
		t.Fatalf(
			"player 2: expected %q, got %q",
			MessageGameStarted,
			message.Type,
		)
	}

	// Verify the actual game state.
	room := server.manager.GetRoom(player1.GameID)

	if room == nil {
		t.Fatal("expected room to exist")
	}

	snapshot := room.Snapshot()

	if snapshot.Status != game.Playing {
		t.Fatalf(
			"expected game status %q, got %q",
			game.Playing,
			snapshot.Status,
		)
	}

	if snapshot.White == nil {
		t.Fatal("expected white player")
	}

	if snapshot.Black == nil {
		t.Fatal("expected black player")
	}
}

func TestGameWebSocket_InvalidSession(t *testing.T) {
	_, httpServer := newTestServer(t)

	player := createQuickGame(t, httpServer, "Player1")

	conn, resp, err := websocket.DefaultDialer.Dial(
		wsURL(
			httpServer.URL,
			player.GameID,
			"invalid-session",
		),
		nil,
	)

	if err == nil {
		conn.Close()
		t.Fatal("expected websocket connection to be rejected")
	}

	if resp == nil {
		t.Fatal("expected HTTP response")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf(
			"expected status 403, got %d",
			resp.StatusCode,
		)
	}
}

func TestGameWebSocket_WrongGame(t *testing.T) {
	server, httpServer := newTestServer(t)

	player := createQuickGame(t, httpServer, "Player1")

	// The session is valid, but it belongs to another game.
	wrongGame := "00000000-0000-0000-0000-000000000000"

	conn, resp, err := websocket.DefaultDialer.Dial(
		wsURL(
			httpServer.URL,
			wrongGame,
			player.SessionID,
		),
		nil,
	)

	if err == nil {
		conn.Close()
		t.Fatal("expected websocket connection to be rejected")
	}

	if resp == nil {
		t.Fatal("expected HTTP response")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf(
			"expected status 403, got %d",
			resp.StatusCode,
		)
	}

	// The original room must still exist.
	if server.manager.GetRoom(player.GameID) == nil {
		t.Fatal("expected original game to remain")
	}
}

func TestGameWebSocket_GameNotFound(t *testing.T) {
	_, httpServer := newTestServer(t)

	conn, resp, err := websocket.DefaultDialer.Dial(
		wsURL(
			httpServer.URL,
			"00000000-0000-0000-0000-000000000000",
			"invalid-session",
		),
		nil,
	)

	if err == nil {
		conn.Close()
		t.Fatal("expected websocket connection to be rejected")
	}

	if resp == nil {
		t.Fatal("expected HTTP response")
	}
	defer resp.Body.Close()

	// Session validation happens before room lookup, so this is 403.
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf(
			"expected status 403, got %d",
			resp.StatusCode,
		)
	}
}

func TestGameWebSocket_PlayerNotFound(t *testing.T) {
	server, httpServer := newTestServer(t)

	player := createQuickGame(t, httpServer, "Player1")

	// Create a valid session for the same game,
	// but for a player that is not actually in the room.
	session := server.createSession(
		player.GameID,
		"nonexistent-player",
	)

	conn, resp, err := websocket.DefaultDialer.Dial(
		wsURL(
			httpServer.URL,
			player.GameID,
			session.ID,
		),
		nil,
	)

	if err == nil {
		conn.Close()
		t.Fatal("expected websocket connection to be rejected")
	}

	if resp == nil {
		t.Fatal("expected HTTP response")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf(
			"expected status 403, got %d",
			resp.StatusCode,
		)
	}

	// The real player must still be able to connect.
	conn = connectWS(
		t,
		httpServer,
		player.GameID,
		player.SessionID,
	)
	defer conn.Close()
}

func TestGameWebSocket_ValidSession_GameNotFound(t *testing.T) {
	server, httpServer := newTestServer(t)

	nonexistentGameID := "00000000-0000-0000-0000-000000000000"

	// Create a valid session, but point it at a game
	// that does not exist in the manager.
	session := server.createSession(
		nonexistentGameID,
		"player-123",
	)

	conn, resp, err := websocket.DefaultDialer.Dial(
		wsURL(
			httpServer.URL,
			nonexistentGameID,
			session.ID,
		),
		nil,
	)

	if err == nil {
		conn.Close()
		t.Fatal("expected websocket connection to be rejected")
	}

	if resp == nil {
		t.Fatal("expected HTTP response")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf(
			"expected status 404, got %d",
			resp.StatusCode,
		)
	}

	if server.manager.GetRoom(nonexistentGameID) != nil {
		t.Fatal("expected game to not exist")
	}
}

func waitForCondition(
	t *testing.T,
	timeout time.Duration,
	condition func() bool,
) {
	t.Helper()

	deadline := time.Now().Add(timeout)

	for time.Now().Before(deadline) {
		if condition() {
			return
		}

		time.Sleep(10 * time.Millisecond)
	}

	t.Fatal("condition was not satisfied before timeout")
}

func TestGameWebSocket_Reconnect_ReplacesOldConnection(t *testing.T) {
	server, httpServer := newTestServer(t)

	player := createQuickGame(t, httpServer, "Player1")

	conn1 := connectWS(
		t,
		httpServer,
		player.GameID,
		player.SessionID,
	)
	defer conn1.Close()

	// Consume the initial waiting state.
	var message Message

	if err := conn1.ReadJSON(&message); err != nil {
		t.Fatal(err)
	}

	if message.Type != MessageGameState {
		t.Fatalf(
			"expected %q, got %q",
			MessageGameState,
			message.Type,
		)
	}

	// Capture the server-side client representing conn1.
	session := server.getSession(player.SessionID)
	if session == nil {
		t.Fatal("expected session to exist")
	}

	waitForCondition(t, time.Second, func() bool {
		server.mu.RLock()
		defer server.mu.RUnlock()

		return server.clients[session.PlayerID] != nil
	})

	server.mu.RLock()
	oldServerClient := server.clients[session.PlayerID]
	server.mu.RUnlock()

	if oldServerClient == nil {
		t.Fatal("expected first server-side client")
	}

	// Connect again using the same session.
	conn2 := connectWS(
		t,
		httpServer,
		player.GameID,
		player.SessionID,
	)
	defer conn2.Close()

	// The server must replace the old client with a new server-side client.
	waitForCondition(t, time.Second, func() bool {
		server.mu.RLock()
		defer server.mu.RUnlock()

		current := server.clients[session.PlayerID]

		return current != nil && current != oldServerClient
	})

	server.mu.RLock()
	newServerClient := server.clients[session.PlayerID]
	server.mu.RUnlock()

	if newServerClient == oldServerClient {
		t.Fatal("expected reconnect to replace the old server-side client")
	}

	// Give the old read goroutine time to finish.
	// Its cleanup must NOT remove the new client.
	time.Sleep(50 * time.Millisecond)

	server.mu.RLock()
	current := server.clients[session.PlayerID]
	server.mu.RUnlock()

	if current != newServerClient {
		t.Fatal(
			"old connection cleanup removed or replaced the new active connection",
		)
	}

	// The new client should still be registered.
	if !server.isCurrentClient(newServerClient) {
		t.Fatal("expected new connection to remain the current client")
	}
}

func TestGameWebSocket_Reconnect_DuringGame(t *testing.T) {
	server, httpServer := newTestServer(t)

	player1 := createQuickGame(t, httpServer, "Player1")
	player2 := createQuickGame(t, httpServer, "Player2")

	conn1 := connectWS(
		t,
		httpServer,
		player1.GameID,
		player1.SessionID,
	)
	defer conn1.Close()

	var message Message

	// Player 1 is waiting.
	if err := conn1.ReadJSON(&message); err != nil {
		t.Fatal(err)
	}

	conn2 := connectWS(
		t,
		httpServer,
		player2.GameID,
		player2.SessionID,
	)
	defer conn2.Close()

	// Both players receive game_started.
	if err := conn1.ReadJSON(&message); err != nil {
		t.Fatal(err)
	}

	if message.Type != MessageGameStarted {
		t.Fatalf(
			"player 1: expected %q, got %q",
			MessageGameStarted,
			message.Type,
		)
	}

	if err := conn2.ReadJSON(&message); err != nil {
		t.Fatal(err)
	}

	if message.Type != MessageGameStarted {
		t.Fatalf(
			"player 2: expected %q, got %q",
			MessageGameStarted,
			message.Type,
		)
	}

	room := server.manager.GetRoom(player1.GameID)

	if room == nil {
		t.Fatal("expected room to exist")
	}

	if snapshot := room.Snapshot(); snapshot.Status != game.Playing {
		t.Fatalf(
			"expected game to be playing, got %q",
			snapshot.Status,
		)
	}

	// Capture the original server-side client.
	session := server.getSession(player1.SessionID)
	if session == nil {
		t.Fatal("expected session")
	}

	waitForCondition(t, time.Second, func() bool {
		server.mu.RLock()
		defer server.mu.RUnlock()

		return server.clients[session.PlayerID] != nil
	})

	server.mu.RLock()
	oldServerClient := server.clients[session.PlayerID]
	server.mu.RUnlock()

	// Reconnect Player 1.
	conn3 := connectWS(
		t,
		httpServer,
		player1.GameID,
		player1.SessionID,
	)
	defer conn3.Close()

	// Wait until the replacement is registered.
	waitForCondition(t, time.Second, func() bool {
		server.mu.RLock()
		defer server.mu.RUnlock()

		current := server.clients[session.PlayerID]

		return current != nil && current != oldServerClient
	})

	server.mu.RLock()
	newServerClient := server.clients[session.PlayerID]
	server.mu.RUnlock()

	if newServerClient == oldServerClient {
		t.Fatal("expected reconnect to replace old client")
	}

	// The old connection's cleanup must not disconnect Player 1.
	waitForCondition(t, time.Second, func() bool {
		server.mu.RLock()
		defer server.mu.RUnlock()

		return server.clients[session.PlayerID] == newServerClient
	})

	// Most importantly: the game must still be playing.
	snapshot := room.Snapshot()

	if snapshot.Status != game.Playing {
		t.Fatalf(
			"expected game to remain playing after reconnect, got %q",
			snapshot.Status,
		)
	}

	player2Session := server.getSession(player2.SessionID)

	if player2Session == nil {
		t.Fatal("expected player 2 session")
	}

	server.mu.RLock()
	player2Client := server.clients[player2Session.PlayerID]
	server.mu.RUnlock()

	if player2Client == nil {
		t.Fatal("expected player 2 to remain connected")
	}

	if player2Client == nil {
		t.Fatal("expected player 2 to remain connected")
	}
}

func TestGameWebSocket_ValidMove(t *testing.T) {
	server, httpServer := newTestServer(t)

	player1 := createQuickGame(t, httpServer, "Player1")
	player2 := createQuickGame(t, httpServer, "Player2")

	conn1 := connectWS(
		t,
		httpServer,
		player1.GameID,
		player1.SessionID,
	)
	defer conn1.Close()

	var message Message

	// Player 1 receives the initial waiting state.
	if err := conn1.ReadJSON(&message); err != nil {
		t.Fatal(err)
	}

	conn2 := connectWS(
		t,
		httpServer,
		player2.GameID,
		player2.SessionID,
	)
	defer conn2.Close()

	// Game started → both players receive game_started.
	if err := conn1.ReadJSON(&message); err != nil {
		t.Fatal(err)
	}

	if message.Type != MessageGameStarted {
		t.Fatalf(
			"player 1: expected %q, got %q",
			MessageGameStarted,
			message.Type,
		)
	}

	if err := conn2.ReadJSON(&message); err != nil {
		t.Fatal(err)
	}

	if message.Type != MessageGameStarted {
		t.Fatalf(
			"player 2: expected %q, got %q",
			MessageGameStarted,
			message.Type,
		)
	}

	// Read the GameState from the server so we know
	// which player received White.
	var state1 GameState

	if err := json.Unmarshal(message.Data, &state1); err != nil {
		t.Fatal(err)
	}

	if state1.Status != game.Playing {
		t.Fatalf(
			"expected playing status, got %q",
			state1.Status,
		)
	}

	var whitePlayerID string

	if state1.White != nil {
		whitePlayerID = state1.White.ID
	}

	if whitePlayerID == "" {
		t.Fatal("expected white player")
	}

	// Determine which connection belongs to White.
	whiteConn := conn1

	if whitePlayerID == player2SessionPlayerID(server, player2.SessionID) {
		whiteConn = conn2
	}

	// Play e2 -> e4.
	moveData, err := json.Marshal(MoveCommand{
		From: "e2",
		To:   "e4",
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := whiteConn.WriteJSON(Message{
		Type: MessageMove,
		Data: moveData,
	}); err != nil {
		t.Fatal(err)
	}

	// Both players should receive the updated state.
	for i, conn := range []*websocket.Conn{conn1, conn2} {
		if err := conn.ReadJSON(&message); err != nil {
			t.Fatalf(
				"player %d failed to receive move state: %v",
				i+1,
				err,
			)
		}

		if message.Type != MessageGameState {
			t.Fatalf(
				"player %d: expected %q, got %q",
				i+1,
				MessageGameState,
				message.Type,
			)
		}

		var state GameState

		if err := json.Unmarshal(message.Data, &state); err != nil {
			t.Fatal(err)
		}

		if state.Status != game.Playing {
			t.Fatalf(
				"player %d: expected playing status, got %q",
				i+1,
				state.Status,
			)
		}

		if state.FEN == "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1" {
			t.Fatalf(
				"player %d: expected FEN to change after move",
				i+1,
			)
		}

		if state.Active == state1.Active {
			t.Fatalf(
				"player %d: expected active color to switch",
				i+1,
			)
		}
	}
}

func player2SessionPlayerID(
	server *Server,
	sessionID string,
) string {
	session := server.getSession(sessionID)

	if session == nil {
		return ""
	}

	return session.PlayerID
}

func TestGameWebSocket_WrongTurn(t *testing.T) {
	server, httpServer := newTestServer(t)

	player1 := createQuickGame(t, httpServer, "Player1")
	player2 := createQuickGame(t, httpServer, "Player2")

	conn1 := connectWS(
		t,
		httpServer,
		player1.GameID,
		player1.SessionID,
	)
	defer conn1.Close()

	var message Message

	// Initial waiting state for player 1.
	if err := conn1.ReadJSON(&message); err != nil {
		t.Fatal(err)
	}

	conn2 := connectWS(
		t,
		httpServer,
		player2.GameID,
		player2.SessionID,
	)
	defer conn2.Close()

	// Game started for player 1.
	if err := conn1.ReadJSON(&message); err != nil {
		t.Fatal(err)
	}

	if message.Type != MessageGameStarted {
		t.Fatalf(
			"player 1: expected %q, got %q",
			MessageGameStarted,
			message.Type,
		)
	}

	// Game started for player 2.
	if err := conn2.ReadJSON(&message); err != nil {
		t.Fatal(err)
	}

	if message.Type != MessageGameStarted {
		t.Fatalf(
			"player 2: expected %q, got %q",
			MessageGameStarted,
			message.Type,
		)
	}

	room := server.manager.GetRoom(player1.GameID)

	if room == nil {
		t.Fatal("expected room")
	}

	before := room.Snapshot()

	// Find the player who is NOT supposed to move.
	var wrongConn *websocket.Conn

	player1Session := server.getSession(player1.SessionID)
	player2Session := server.getSession(player2.SessionID)

	if player1Session == nil || player2Session == nil {
		t.Fatal("expected both sessions")
	}

	if before.Active == game.White {
		if before.White.ID == player1Session.PlayerID {
			wrongConn = conn2
		} else {
			wrongConn = conn1
		}
	} else {
		if before.Black.ID == player1Session.PlayerID {
			wrongConn = conn2
		} else {
			wrongConn = conn1
		}
	}

	// Attempt e2 -> e4 from the wrong player.
	moveData, err := json.Marshal(MoveCommand{
		From: "e2",
		To:   "e4",
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := wrongConn.WriteJSON(Message{
		Type: MessageMove,
		Data: moveData,
	}); err != nil {
		t.Fatal(err)
	}

	// The wrong player should receive an error.
	var errorMessage Message

	if wrongConn == conn1 {
		if err := conn1.ReadJSON(&errorMessage); err != nil {
			t.Fatal(err)
		}
	} else {
		if err := conn2.ReadJSON(&errorMessage); err != nil {
			t.Fatal(err)
		}
	}

	if errorMessage.Type != MessageError {
		t.Fatalf(
			"expected %q, got %q",
			MessageError,
			errorMessage.Type,
		)
	}

	var errorData ErrorData

	if err := json.Unmarshal(errorMessage.Data, &errorData); err != nil {
		t.Fatal(err)
	}

	if errorData.Code != ErrorNotYourTurn {
		t.Fatalf(
			"expected error code %q, got %q",
			ErrorNotYourTurn,
			errorData.Code,
		)
	}

	// The game state must not have changed.
	after := room.Snapshot()

	if after.FEN != before.FEN {
		t.Fatal("expected FEN to remain unchanged")
	}

	if after.Active != before.Active {
		t.Fatal("expected active color to remain unchanged")
	}

	if after.Status != game.Playing {
		t.Fatalf(
			"expected game to remain playing, got %q",
			after.Status,
		)
	}
}

func TestGameWebSocket_InvalidMove(t *testing.T) {
	server, httpServer := newTestServer(t)

	player1 := createQuickGame(t, httpServer, "Player1")
	player2 := createQuickGame(t, httpServer, "Player2")

	conn1 := connectWS(
		t,
		httpServer,
		player1.GameID,
		player1.SessionID,
	)
	defer conn1.Close()

	var message Message

	// Initial waiting state for player 1.
	if err := conn1.ReadJSON(&message); err != nil {
		t.Fatal(err)
	}

	conn2 := connectWS(
		t,
		httpServer,
		player2.GameID,
		player2.SessionID,
	)
	defer conn2.Close()

	// Game started for player 1.
	if err := conn1.ReadJSON(&message); err != nil {
		t.Fatal(err)
	}

	if message.Type != MessageGameStarted {
		t.Fatalf(
			"player 1: expected %q, got %q",
			MessageGameStarted,
			message.Type,
		)
	}

	// Game started for player 2.
	if err := conn2.ReadJSON(&message); err != nil {
		t.Fatal(err)
	}

	if message.Type != MessageGameStarted {
		t.Fatalf(
			"player 2: expected %q, got %q",
			MessageGameStarted,
			message.Type,
		)
	}

	room := server.manager.GetRoom(player1.GameID)

	if room == nil {
		t.Fatal("expected room")
	}

	before := room.Snapshot()

	player1Session := server.getSession(player1.SessionID)
	player2Session := server.getSession(player2.SessionID)

	if player1Session == nil || player2Session == nil {
		t.Fatal("expected both sessions")
	}

	// Determine the player whose turn it is.
	var activeConn *websocket.Conn

	if before.Active == game.White {
		if before.White == nil {
			t.Fatal("expected white player")
		}

		if before.White.ID == player1Session.PlayerID {
			activeConn = conn1
		} else if before.White.ID == player2Session.PlayerID {
			activeConn = conn2
		} else {
			t.Fatal("active white player does not match either session")
		}
	} else {
		if before.Black == nil {
			t.Fatal("expected black player")
		}

		if before.Black.ID == player1Session.PlayerID {
			activeConn = conn1
		} else if before.Black.ID == player2Session.PlayerID {
			activeConn = conn2
		} else {
			t.Fatal("active black player does not match either session")
		}
	}

	// Send an illegal pawn move for the active side.
	var from, to string

	if before.Active == game.White {
		// White pawn cannot move three squares.
		from = "e2"
		to = "e5"
	} else {
		// Black pawn cannot move three squares.
		from = "e7"
		to = "e4"
	}

	moveData, err := json.Marshal(MoveCommand{
		From: from,
		To:   to,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := activeConn.WriteJSON(Message{
		Type: MessageMove,
		Data: moveData,
	}); err != nil {
		t.Fatal(err)
	}

	// The active player should receive an invalid-move error.
	var errorMessage Message

	if err := activeConn.ReadJSON(&errorMessage); err != nil {
		t.Fatal(err)
	}

	if errorMessage.Type != MessageError {
		t.Fatalf(
			"expected %q, got %q",
			MessageError,
			errorMessage.Type,
		)
	}

	var errorData ErrorData

	if err := json.Unmarshal(errorMessage.Data, &errorData); err != nil {
		t.Fatal(err)
	}

	if errorData.Code != ErrorInvalidMove {
		t.Fatalf(
			"expected error code %q, got %q",
			ErrorInvalidMove,
			errorData.Code,
		)
	}

	// The illegal move must not change the board or turn.
	after := room.Snapshot()

	if after.FEN != before.FEN {
		t.Fatalf("expected FEN to remain unchanged")
	}

	if after.Active != before.Active {
		t.Fatalf("expected active color to remain unchanged")
	}

	if after.Status != game.Playing {
		t.Fatalf(
			"expected game to remain playing, got %q",
			after.Status,
		)
	}
}

func TestGameWebSocket_MalformedMessage(t *testing.T) {
	server, httpServer := newTestServer(t)

	player1 := createQuickGame(t, httpServer, "Player1")
	player2 := createQuickGame(t, httpServer, "Player2")

	conn1 := connectWS(
		t,
		httpServer,
		player1.GameID,
		player1.SessionID,
	)
	defer conn1.Close()

	var message Message

	// Initial waiting state.
	if err := conn1.ReadJSON(&message); err != nil {
		t.Fatal(err)
	}

	conn2 := connectWS(
		t,
		httpServer,
		player2.GameID,
		player2.SessionID,
	)
	defer conn2.Close()

	// Game started for player 1.
	if err := conn1.ReadJSON(&message); err != nil {
		t.Fatal(err)
	}

	if message.Type != MessageGameStarted {
		t.Fatalf(
			"player 1: expected %q, got %q",
			MessageGameStarted,
			message.Type,
		)
	}

	// Game started for player 2.
	if err := conn2.ReadJSON(&message); err != nil {
		t.Fatal(err)
	}

	if message.Type != MessageGameStarted {
		t.Fatalf(
			"player 2: expected %q, got %q",
			MessageGameStarted,
			message.Type,
		)
	}

	room := server.manager.GetRoom(player1.GameID)

	if room == nil {
		t.Fatal("expected room")
	}

	before := room.Snapshot()

	player1Session := server.getSession(player1.SessionID)
	player2Session := server.getSession(player2.SessionID)

	if player1Session == nil || player2Session == nil {
		t.Fatal("expected both sessions")
	}

	// Find the player whose turn it is.
	var activeConn *websocket.Conn

	if before.Active == game.White {
		if before.White == nil {
			t.Fatal("expected white player")
		}

		if before.White.ID == player1Session.PlayerID {
			activeConn = conn1
		} else if before.White.ID == player2Session.PlayerID {
			activeConn = conn2
		} else {
			t.Fatal("active white player does not match either session")
		}
	} else {
		if before.Black == nil {
			t.Fatal("expected black player")
		}

		if before.Black.ID == player1Session.PlayerID {
			activeConn = conn1
		} else if before.Black.ID == player2Session.PlayerID {
			activeConn = conn2
		} else {
			t.Fatal("active black player does not match either session")
		}
	}

	// Send deliberately malformed JSON.
	if err := activeConn.WriteMessage(
		websocket.TextMessage,
		[]byte(`{"type": "move", "data":`),
	); err != nil {
		t.Fatal(err)
	}

	// The server should reject it with MessageError.
	var errorMessage Message

	if err := activeConn.ReadJSON(&errorMessage); err != nil {
		t.Fatal(err)
	}

	if errorMessage.Type != MessageError {
		t.Fatalf(
			"expected %q, got %q",
			MessageError,
			errorMessage.Type,
		)
	}

	var errorData ErrorData

	if err := json.Unmarshal(errorMessage.Data, &errorData); err != nil {
		t.Fatal(err)
	}

	if errorData.Code != ErrorInvalidMessage {
		t.Fatalf(
			"expected error code %q, got %q",
			ErrorInvalidMessage,
			errorData.Code,
		)
	}

	// The malformed message must not change the game.
	afterMalformed := room.Snapshot()

	if afterMalformed.FEN != before.FEN {
		t.Fatal("expected FEN to remain unchanged")
	}

	if afterMalformed.Active != before.Active {
		t.Fatal("expected active color to remain unchanged")
	}

	if afterMalformed.Status != game.Playing {
		t.Fatalf(
			"expected game to remain playing, got %q",
			afterMalformed.Status,
		)
	}

	// ------------------------------------------------------------
	// Prove the WebSocket is still usable.
	// ------------------------------------------------------------

	var from, to string

	if before.Active == game.White {
		from = "e2"
		to = "e4"
	} else {
		from = "e7"
		to = "e5"
	}

	moveData, err := json.Marshal(MoveCommand{
		From: from,
		To:   to,
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := activeConn.WriteJSON(Message{
		Type: MessageMove,
		Data: moveData,
	}); err != nil {
		t.Fatal(err)
	}

	// The active player should receive the resulting game state.
	var stateMessage Message

	if err := activeConn.ReadJSON(&stateMessage); err != nil {
		t.Fatal(err)
	}

	if stateMessage.Type != MessageGameState {
		t.Fatalf(
			"expected %q after valid move, got %q",
			MessageGameState,
			stateMessage.Type,
		)
	}

	// The move must actually have changed the game.
	afterMove := room.Snapshot()

	if afterMove.FEN == before.FEN {
		t.Fatal("expected FEN to change after valid move")
	}

	if afterMove.Active == before.Active {
		t.Fatal("expected active color to change after valid move")
	}

	if afterMove.Status != game.Playing {
		t.Fatalf(
			"expected game to remain playing, got %q",
			afterMove.Status,
		)
	}
}

func TestGameWebSocket_Resign(t *testing.T) {
	server, httpServer := newTestServer(t)

	player1 := createQuickGame(t, httpServer, "Player1")
	player2 := createQuickGame(t, httpServer, "Player2")

	conn1 := connectWS(
		t,
		httpServer,
		player1.GameID,
		player1.SessionID,
	)
	defer conn1.Close()

	var message Message

	// Initial waiting state.
	if err := conn1.ReadJSON(&message); err != nil {
		t.Fatal(err)
	}

	conn2 := connectWS(
		t,
		httpServer,
		player2.GameID,
		player2.SessionID,
	)
	defer conn2.Close()

	// Game started for player 1.
	if err := conn1.ReadJSON(&message); err != nil {
		t.Fatal(err)
	}

	if message.Type != MessageGameStarted {
		t.Fatalf(
			"player 1: expected %q, got %q",
			MessageGameStarted,
			message.Type,
		)
	}

	// Game started for player 2.
	if err := conn2.ReadJSON(&message); err != nil {
		t.Fatal(err)
	}

	if message.Type != MessageGameStarted {
		t.Fatalf(
			"player 2: expected %q, got %q",
			MessageGameStarted,
			message.Type,
		)
	}

	room := server.manager.GetRoom(player1.GameID)

	if room == nil {
		t.Fatal("expected room")
	}

	before := room.Snapshot()

	if before.Status != game.Playing {
		t.Fatalf(
			"expected game to be playing, got %q",
			before.Status,
		)
	}

	player1Session := server.getSession(player1.SessionID)
	player2Session := server.getSession(player2.SessionID)

	if player1Session == nil || player2Session == nil {
		t.Fatal("expected both sessions")
	}

	// Determine the expected result when player 1 resigns.
	var expectedResult game.GameResult

	if before.White.ID == player1Session.PlayerID {
		expectedResult = game.BlackWins
	} else if before.Black.ID == player1Session.PlayerID {
		expectedResult = game.WhiteWins
	} else {
		t.Fatal("player 1 is neither white nor black")
	}

	// Player 1 resigns.
	if err := conn1.WriteJSON(Message{
		Type: MessageResign,
	}); err != nil {
		t.Fatal(err)
	}

	// Both players must receive the final state.
	var state1 Message
	var state2 Message

	if err := conn1.ReadJSON(&state1); err != nil {
		t.Fatal(err)
	}

	if err := conn2.ReadJSON(&state2); err != nil {
		t.Fatal(err)
	}

	if state1.Type != MessageGameState {
		t.Fatalf(
			"player 1: expected %q, got %q",
			MessageGameState,
			state1.Type,
		)
	}

	if state2.Type != MessageGameState {
		t.Fatalf(
			"player 2: expected %q, got %q",
			MessageGameState,
			state2.Type,
		)
	}

	var finalState1 GameState
	var finalState2 GameState

	if err := json.Unmarshal(state1.Data, &finalState1); err != nil {
		t.Fatal(err)
	}

	if err := json.Unmarshal(state2.Data, &finalState2); err != nil {
		t.Fatal(err)
	}

	// Both clients must receive a finished game.
	if finalState1.Status != game.Finished {
		t.Fatalf(
			"player 1: expected status %q, got %q",
			game.Finished,
			finalState1.Status,
		)
	}

	if finalState2.Status != game.Finished {
		t.Fatalf(
			"player 2: expected status %q, got %q",
			game.Finished,
			finalState2.Status,
		)
	}

	// Both clients must receive the same result.
	if game.GameResult(finalState1.Result) != expectedResult {
		t.Fatalf(
			"player 1: expected result %q, got %q",
			expectedResult,
			finalState1.Result,
		)
	}

	if game.GameResult(finalState2.Result) != expectedResult {
		t.Fatalf(
			"player 2: expected result %q, got %q",
			expectedResult,
			finalState2.Result,
		)
	}

	// Both clients must receive the resignation end reason.
	if game.EndReason(finalState1.EndReason) != game.Resignation {
		t.Fatalf(
			"player 1: expected end reason %q, got %q",
			game.Resignation,
			finalState1.EndReason,
		)
	}

	if game.EndReason(finalState2.EndReason) != game.Resignation {
		t.Fatalf(
			"player 2: expected end reason %q, got %q",
			game.Resignation,
			finalState2.EndReason,
		)
	}

	// Server-side state must match the broadcast state.
	after := room.Snapshot()

	if after.Status != game.Finished {
		t.Fatalf(
			"expected room status %q, got %q",
			game.Finished,
			after.Status,
		)
	}

	if game.GameResult(after.Result) != expectedResult {
		t.Fatalf(
			"expected room result %q, got %q",
			expectedResult,
			after.Result,
		)
	}

	if game.EndReason(after.EndReason) != game.Resignation {
		t.Fatalf(
			"expected room end reason %q, got %q",
			game.Resignation,
			after.EndReason,
		)
	}
}

func TestGameWebSocket_DrawLifecycle(t *testing.T) {
	server, httpServer := newTestServer(t)

	player1 := createQuickGame(t, httpServer, "Player1")
	player2 := createQuickGame(t, httpServer, "Player2")

	conn1 := connectWS(t, httpServer, player1.GameID, player1.SessionID)
	defer conn1.Close()

	var message Message

	// First player connects.
	if err := conn1.ReadJSON(&message); err != nil {
		t.Fatal(err)
	}

	conn2 := connectWS(t, httpServer, player2.GameID, player2.SessionID)
	defer conn2.Close()

	// Both players receive game_started.
	if err := conn1.ReadJSON(&message); err != nil {
		t.Fatal(err)
	}

	if message.Type != MessageGameStarted {
		t.Fatalf(
			"expected game_started for player 1, got %s",
			message.Type,
		)
	}

	if err := conn2.ReadJSON(&message); err != nil {
		t.Fatal(err)
	}

	if message.Type != MessageGameStarted {
		t.Fatalf(
			"expected game_started for player 2, got %s",
			message.Type,
		)
	}

	room := server.manager.GetRoom(player1.GameID)
	if room == nil {
		t.Fatal("expected room")
	}

	before := room.Snapshot()

	if before.Status != game.Playing {
		t.Fatalf(
			"expected game to be playing, got %s",
			before.Status,
		)
	}

	// ---------------------------------------------------------
	// 1. Player 1 offers a draw.
	// ---------------------------------------------------------

	if err := conn1.WriteJSON(Message{
		Type: MessageOfferDraw,
	}); err != nil {
		t.Fatal(err)
	}

	var drawOffered Message

	if err := conn2.ReadJSON(&drawOffered); err != nil {
		t.Fatal(err)
	}

	if drawOffered.Type != MessageDrawOffered {
		t.Fatalf(
			"expected draw_offered, got %s",
			drawOffered.Type,
		)
	}

	var offer DrawOfferData

	if err := json.Unmarshal(drawOffered.Data, &offer); err != nil {
		t.Fatal(err)
	}

	player1Session := server.getSession(player1.SessionID)
	if player1Session == nil {
		t.Fatal("expected player 1 session")
	}

	if offer.PlayerID != player1Session.PlayerID {
		t.Fatalf(
			"expected draw offer from %s, got %s",
			player1Session.PlayerID,
			offer.PlayerID,
		)
	}

	// The game must still be playing.
	afterOffer := room.Snapshot()

	if afterOffer.Status != game.Playing {
		t.Fatalf(
			"expected game to remain playing after draw offer, got %s",
			afterOffer.Status,
		)
	}

	if game.GameResult(afterOffer.Result) != game.NoResult {
		t.Fatalf(
			"expected no result after draw offer, got %s",
			afterOffer.Result,
		)
	}

	if game.EndReason(afterOffer.EndReason) != game.NoEndReason {
		t.Fatalf(
			"expected no end reason after draw offer, got %s",
			afterOffer.EndReason,
		)
	}

	// ---------------------------------------------------------
	// 2. Player 2 declines the draw.
	// ---------------------------------------------------------

	if err := conn2.WriteJSON(Message{
		Type: MessageRespondDraw,
		Data: mustJSON(DrawResponseCommand{
			Accepted: false,
		}),
	}); err != nil {
		t.Fatal(err)
	}

	var drawDeclined Message

	if err := conn1.ReadJSON(&drawDeclined); err != nil {
		t.Fatal(err)
	}

	if drawDeclined.Type != MessageDrawDeclined {
		t.Fatalf(
			"expected draw_declined, got %s",
			drawDeclined.Type,
		)
	}

	afterDecline := room.Snapshot()

	if afterDecline.Status != game.Playing {
		t.Fatalf(
			"expected game to remain playing after draw decline, got %s",
			afterDecline.Status,
		)
	}

	if game.GameResult(afterDecline.Result) != game.NoResult {
		t.Fatalf(
			"expected no result after draw decline, got %s",
			afterDecline.Result,
		)
	}

	if game.EndReason(afterDecline.EndReason) != game.NoEndReason {
		t.Fatalf(
			"expected no end reason after draw decline, got %s",
			afterDecline.EndReason,
		)
	}

	// ---------------------------------------------------------
	// 3. Player 1 can offer another draw.
	// ---------------------------------------------------------

	if err := conn1.WriteJSON(Message{
		Type: MessageOfferDraw,
	}); err != nil {
		t.Fatal(err)
	}

	var secondOffer Message

	if err := conn2.ReadJSON(&secondOffer); err != nil {
		t.Fatal(err)
	}

	if secondOffer.Type != MessageDrawOffered {
		t.Fatalf(
			"expected second draw_offered, got %s",
			secondOffer.Type,
		)
	}

	// ---------------------------------------------------------
	// 4. Player 2 accepts.
	// ---------------------------------------------------------

	if err := conn2.WriteJSON(Message{
		Type: MessageRespondDraw,
		Data: mustJSON(DrawResponseCommand{
			Accepted: true,
		}),
	}); err != nil {
		t.Fatal(err)
	}

	var final1 Message
	var final2 Message

	if err := conn1.ReadJSON(&final1); err != nil {
		t.Fatal(err)
	}

	if err := conn2.ReadJSON(&final2); err != nil {
		t.Fatal(err)
	}

	if final1.Type != MessageGameState {
		t.Fatalf(
			"expected game_state for player 1, got %s",
			final1.Type,
		)
	}

	if final2.Type != MessageGameState {
		t.Fatalf(
			"expected game_state for player 2, got %s",
			final2.Type,
		)
	}

	var state1 GameState
	var state2 GameState

	if err := json.Unmarshal(final1.Data, &state1); err != nil {
		t.Fatal(err)
	}

	if err := json.Unmarshal(final2.Data, &state2); err != nil {
		t.Fatal(err)
	}

	// ---------------------------------------------------------
	// 5. Both clients must see the finished draw.
	// ---------------------------------------------------------

	for name, state := range map[string]GameState{
		"player1": state1,
		"player2": state2,
	} {
		if state.Status != game.Finished {
			t.Fatalf(
				"%s: expected finished status, got %s",
				name,
				state.Status,
			)
		}

		if game.GameResult(state.Result) != game.Draw {
			t.Fatalf(
				"%s: expected draw result, got %s",
				name,
				state.Result,
			)
		}

		if game.EndReason(state.EndReason) != game.DrawAgreement {
			t.Fatalf(
				"%s: expected draw_agreement, got %s",
				name,
				state.EndReason,
			)
		}
	}

	// ---------------------------------------------------------
	// 6. Server-side state must match.
	// ---------------------------------------------------------

	after := room.Snapshot()

	if after.Status != game.Finished {
		t.Fatalf(
			"expected server game to be finished, got %s",
			after.Status,
		)
	}

	if game.GameResult(after.Result) != game.Draw {
		t.Fatalf(
			"expected server result to be draw, got %s",
			after.Result,
		)
	}

	if game.EndReason(after.EndReason) != game.DrawAgreement {
		t.Fatalf(
			"expected server end reason to be draw_agreement, got %s",
			after.EndReason,
		)
	}
}

func mustJSON(v any) json.RawMessage {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}

	return data
}

func TestGameWebSocket_QuitDuringGame(t *testing.T) {
	server, httpServer := newTestServer(t)

	player1 := createQuickGame(t, httpServer, "Player1")
	player2 := createQuickGame(t, httpServer, "Player2")

	conn1 := connectWS(t, httpServer, player1.GameID, player1.SessionID)
	defer conn1.Close()

	var message Message

	if err := conn1.ReadJSON(&message); err != nil {
		t.Fatal(err)
	}

	conn2 := connectWS(t, httpServer, player2.GameID, player2.SessionID)
	defer conn2.Close()

	if err := conn1.ReadJSON(&message); err != nil {
		t.Fatal(err)
	}

	if message.Type != MessageGameStarted {
		t.Fatalf("expected game_started for player 1, got %s", message.Type)
	}

	if err := conn2.ReadJSON(&message); err != nil {
		t.Fatal(err)
	}

	if message.Type != MessageGameStarted {
		t.Fatalf("expected game_started for player 2, got %s", message.Type)
	}

	room := server.manager.GetRoom(player1.GameID)
	if room == nil {
		t.Fatal("expected room")
	}

	before := room.Snapshot()

	if before.Status != game.Playing {
		t.Fatalf("expected game to be playing, got %s", before.Status)
	}

	player1Session := server.getSession(player1.SessionID)
	if player1Session == nil {
		t.Fatal("expected player 1 session")
	}

	var expectedResult game.GameResult

	switch {
	case before.White != nil && before.White.ID == player1Session.PlayerID:
		expectedResult = game.BlackWins

	case before.Black != nil && before.Black.ID == player1Session.PlayerID:
		expectedResult = game.WhiteWins

	default:
		t.Fatal("player 1 is neither white nor black")
	}

	req, err := http.NewRequest(
		http.MethodDelete,
		httpServer.URL+
			"/games?sessionId="+
			url.QueryEscape(player1.SessionID),
		nil,
	)
	if err != nil {
		t.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", resp.StatusCode)
	}

	if server.getSession(player1.SessionID) != nil {
		t.Fatal("expected player 1 session to be deleted")
	}

	var finalMessage1 Message
	var finalMessage2 Message

	if err := conn1.ReadJSON(&finalMessage1); err != nil {
		t.Fatal(err)
	}

	if err := conn2.ReadJSON(&finalMessage2); err != nil {
		t.Fatal(err)
	}

	if finalMessage1.Type != MessageGameState {
		t.Fatalf("expected game_state for player 1, got %s", finalMessage1.Type)
	}

	if finalMessage2.Type != MessageGameState {
		t.Fatalf("expected game_state for player 2, got %s", finalMessage2.Type)
	}

	var state1 GameState
	var state2 GameState

	if err := json.Unmarshal(finalMessage1.Data, &state1); err != nil {
		t.Fatal(err)
	}

	if err := json.Unmarshal(finalMessage2.Data, &state2); err != nil {
		t.Fatal(err)
	}

	for name, state := range map[string]GameState{
		"player1": state1,
		"player2": state2,
	} {
		if state.Status != game.Finished {
			t.Fatalf("%s: expected finished, got %s", name, state.Status)
		}

		if game.GameResult(state.Result) != expectedResult {
			t.Fatalf(
				"%s: expected result %s, got %s",
				name,
				expectedResult,
				state.Result,
			)
		}

		if game.EndReason(state.EndReason) != game.EndReasonQuit {
			t.Fatalf(
				"%s: expected end reason %s, got %s",
				name,
				game.EndReasonQuit,
				state.EndReason,
			)
		}
	}

	final := room.Snapshot()

	if final.Status != game.Finished {
		t.Fatalf("expected server game to be finished, got %s", final.Status)
	}

	if game.GameResult(final.Result) != expectedResult {
		t.Fatalf(
			"expected server result %s, got %s",
			expectedResult,
			final.Result,
		)
	}

	if game.EndReason(final.EndReason) != game.EndReasonQuit {
		t.Fatalf(
			"expected server end reason %s, got %s",
			game.EndReasonQuit,
			final.EndReason,
		)
	}
}

func TestGameWebSocket_DisconnectTimeout(t *testing.T) {
	server, httpServer := newTestServer(t)

	player1 := createQuickGame(t, httpServer, "Player1")
	player2 := createQuickGame(t, httpServer, "Player2")

	conn1 := connectWS(t, httpServer, player1.GameID, player1.SessionID)

	var message Message

	if err := conn1.ReadJSON(&message); err != nil {
		t.Fatal(err)
	}

	conn2 := connectWS(t, httpServer, player2.GameID, player2.SessionID)
	defer conn2.Close()

	if err := conn1.ReadJSON(&message); err != nil {
		t.Fatal(err)
	}

	if message.Type != MessageGameStarted {
		t.Fatalf("expected game_started for player 1, got %s", message.Type)
	}

	if err := conn2.ReadJSON(&message); err != nil {
		t.Fatal(err)
	}

	if message.Type != MessageGameStarted {
		t.Fatalf("expected game_started for player 2, got %s", message.Type)
	}

	room := server.manager.GetRoom(player1.GameID)
	if room == nil {
		t.Fatal("expected room")
	}

	before := room.Snapshot()

	if before.Status != game.Playing {
		t.Fatalf("expected game to be playing, got %s", before.Status)
	}

	player1Session := server.getSession(player1.SessionID)
	if player1Session == nil {
		t.Fatal("expected player 1 session")
	}

	var expectedResult game.GameResult

	switch {
	case before.White != nil && before.White.ID == player1Session.PlayerID:
		expectedResult = game.BlackWins

	case before.Black != nil && before.Black.ID == player1Session.PlayerID:
		expectedResult = game.WhiteWins

	default:
		t.Fatal("player 1 is neither white nor black")
	}

	if err := conn1.Close(); err != nil {
		t.Fatal(err)
	}

	time.Sleep(500 * time.Millisecond)

	duringGracePeriod := room.Snapshot()

	if duringGracePeriod.Status != game.Playing {
		t.Fatalf(
			"expected game to remain playing during reconnect window, got %s",
			duringGracePeriod.Status,
		)
	}

	if game.EndReason(duringGracePeriod.EndReason) != game.NoEndReason {
		t.Fatalf(
			"expected no end reason during reconnect window, got %s",
			duringGracePeriod.EndReason,
		)
	}

	finalReceived := make(chan Message, 1)
	readErr := make(chan error, 1)

	go func() {
		var final Message

		if err := conn2.ReadJSON(&final); err != nil {
			readErr <- err
			return
		}

		finalReceived <- final
	}()

	select {
	case final := <-finalReceived:
		if final.Type != MessageGameState {
			t.Fatalf(
				"expected game_state after disconnect timeout, got %s",
				final.Type,
			)
		}

		var state GameState

		if err := json.Unmarshal(final.Data, &state); err != nil {
			t.Fatal(err)
		}

		if state.Status != game.Finished {
			t.Fatalf(
				"expected finished state, got %s",
				state.Status,
			)
		}

		if game.GameResult(state.Result) != expectedResult {
			t.Fatalf(
				"expected result %s, got %s",
				expectedResult,
				state.Result,
			)
		}

		if game.EndReason(state.EndReason) != game.EndReasonDisconnect {
			t.Fatalf(
				"expected end reason %s, got %s",
				game.EndReasonDisconnect,
				state.EndReason,
			)
		}

	case err := <-readErr:
		t.Fatalf(
			"opponent connection closed before final state: %v",
			err,
		)

	case <-time.After(35 * time.Second):
		t.Fatal("timed out waiting for disconnect timeout")
	}

	final := room.Snapshot()

	if final.Status != game.Finished {
		t.Fatalf(
			"expected server game to be finished, got %s",
			final.Status,
		)
	}

	if game.GameResult(final.Result) != expectedResult {
		t.Fatalf(
			"expected server result %s, got %s",
			expectedResult,
			final.Result,
		)
	}

	if game.EndReason(final.EndReason) != game.EndReasonDisconnect {
		t.Fatalf(
			"expected server end reason %s, got %s",
			game.EndReasonDisconnect,
			final.EndReason,
		)
	}
}


func readGameStateMessage(
	t *testing.T,
	conn *websocket.Conn,
) GameState {
	t.Helper()

	var message Message

	if err := conn.ReadJSON(&message); err != nil {
		t.Fatal(err)
	}

	if message.Type != MessageGameState {
		t.Fatalf(
			"expected %q, got %q",
			MessageGameState,
			message.Type,
		)
	}

	var state GameState

	if err := json.Unmarshal(message.Data, &state); err != nil {
		t.Fatal(err)
	}

	return state
}

func TestGameWebSocket_MoveAfterTimeout(t *testing.T) {
	server, httpServer := newTestServer(t)

	player1 := createQuickGame(t, httpServer, "Player1")
	player2 := createQuickGame(t, httpServer, "Player2")

	conn1 := connectWS(
		t,
		httpServer,
		player1.GameID,
		player1.SessionID,
	)
	defer conn1.Close()

	var message Message

	// Player 1 receives the initial waiting state.
	if err := conn1.ReadJSON(&message); err != nil {
		t.Fatal(err)
	}

	if message.Type != MessageGameState {
		t.Fatalf(
			"expected %q, got %q",
			MessageGameState,
			message.Type,
		)
	}

	conn2 := connectWS(
		t,
		httpServer,
		player2.GameID,
		player2.SessionID,
	)
	defer conn2.Close()

	// Both players receive game_started.
	if err := conn1.ReadJSON(&message); err != nil {
		t.Fatal(err)
	}

	if message.Type != MessageGameStarted {
		t.Fatalf(
			"player 1: expected %q, got %q",
			MessageGameStarted,
			message.Type,
		)
	}

	if err := conn2.ReadJSON(&message); err != nil {
		t.Fatal(err)
	}

	if message.Type != MessageGameStarted {
		t.Fatalf(
			"player 2: expected %q, got %q",
			MessageGameStarted,
			message.Type,
		)
	}

	room := server.manager.GetRoom(player1.GameID)

	if room == nil {
		t.Fatal("expected room")
	}

	snapshot := room.Snapshot()

	if snapshot.Status != game.Playing {
		t.Fatalf(
			"expected game to be playing, got %q",
			snapshot.Status,
		)
	}

	// Force the active player's clock to zero.
	activeColor := room.Game.Clock.Active
	room.Game.Clock.TimeLeft[activeColor] = 0

	// Determine which player owns the active color.
	var activePlayerID string
	var expectedResult game.GameResult

	player1Session := server.getSession(player1.SessionID)
	player2Session := server.getSession(player2.SessionID)

	if player1Session == nil || player2Session == nil {
		t.Fatal("expected both sessions")
	}

	if activeColor == game.White {
		if snapshot.White == nil {
			t.Fatal("expected white player")
		}

		activePlayerID = snapshot.White.ID
		expectedResult = game.BlackWins
	} else {
		if snapshot.Black == nil {
			t.Fatal("expected black player")
		}

		activePlayerID = snapshot.Black.ID
		expectedResult = game.WhiteWins
	}

	var activeConn *websocket.Conn
	var opponentConn *websocket.Conn

	switch activePlayerID {
	case player1Session.PlayerID:
		activeConn = conn1
		opponentConn = conn2

	case player2Session.PlayerID:
		activeConn = conn2
		opponentConn = conn1

	default:
		t.Fatal("active player does not match either session")
	}

	// Send a move after the active player's clock has reached zero.
	moveData, err := json.Marshal(MoveCommand{
		From: "e2",
		To:   "e4",
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := activeConn.WriteJSON(Message{
		Type: MessageMove,
		Data: moveData,
	}); err != nil {
		t.Fatal(err)
	}

	// The opponent must receive the final game state.
	_ = opponentConn.SetReadDeadline(time.Now().Add(2 * time.Second))

	finalState := readGameStateMessage(t, opponentConn)

	if finalState.Status != game.Finished {
		t.Fatalf(
			"expected finished game, got %q",
			finalState.Status,
		)
	}

	if game.GameResult(finalState.Result) != expectedResult {
		t.Fatalf(
			"expected result %q, got %q",
			expectedResult,
			finalState.Result,
		)
	}

	if game.EndReason(finalState.EndReason) != game.Timeout {
		t.Fatalf(
			"expected end reason %q, got %q",
			game.Timeout,
			finalState.EndReason,
		)
	}

	// The move must not have been played.
	finalSnapshot := room.Snapshot()

	if finalSnapshot.Status != game.Finished {
		t.Fatalf(
			"expected room to be finished, got %q",
			finalSnapshot.Status,
		)
	}

	if game.GameResult(finalSnapshot.Result) != expectedResult {
		t.Fatalf(
			"expected room result %q, got %q",
			expectedResult,
			finalSnapshot.Result,
		)
	}

	if game.EndReason(finalSnapshot.EndReason) != game.Timeout {
		t.Fatalf(
			"expected room end reason %q, got %q",
			game.Timeout,
			finalSnapshot.EndReason,
		)
	}
}

func TestGameWebSocket_ClockWatcherTimeout(t *testing.T) {
	server, httpServer := newTestServer(t)

	player1 := createQuickGame(t, httpServer, "Player1")
	player2 := createQuickGame(t, httpServer, "Player2")

	conn1 := connectWS(
		t,
		httpServer,
		player1.GameID,
		player1.SessionID,
	)
	defer conn1.Close()

	var message Message

	// Player 1 receives the initial waiting state.
	if err := conn1.ReadJSON(&message); err != nil {
		t.Fatal(err)
	}

	if message.Type != MessageGameState {
		t.Fatalf(
			"expected %q, got %q",
			MessageGameState,
			message.Type,
		)
	}

	conn2 := connectWS(
		t,
		httpServer,
		player2.GameID,
		player2.SessionID,
	)
	defer conn2.Close()

	// Both players receive game_started.
	if err := conn1.ReadJSON(&message); err != nil {
		t.Fatal(err)
	}

	if message.Type != MessageGameStarted {
		t.Fatalf(
			"player 1: expected %q, got %q",
			MessageGameStarted,
			message.Type,
		)
	}

	if err := conn2.ReadJSON(&message); err != nil {
		t.Fatal(err)
	}

	if message.Type != MessageGameStarted {
		t.Fatalf(
			"player 2: expected %q, got %q",
			MessageGameStarted,
			message.Type,
		)
	}

	room := server.manager.GetRoom(player1.GameID)

	if room == nil {
		t.Fatal("expected room")
	}

	snapshot := room.Snapshot()

	if snapshot.Status != game.Playing {
		t.Fatalf(
			"expected game to be playing, got %q",
			snapshot.Status,
		)
	}

	// Force the active player's clock to zero.
	activeColor := room.Game.Clock.Active
	room.Game.Clock.TimeLeft[activeColor] = 0

	// Determine which player owns the active color.
	var activePlayerID string
	var expectedResult game.GameResult

	player1Session := server.getSession(player1.SessionID)
	player2Session := server.getSession(player2.SessionID)

	if player1Session == nil || player2Session == nil {
		t.Fatal("expected both sessions")
	}

	if activeColor == game.White {
		if snapshot.White == nil {
			t.Fatal("expected white player")
		}

		activePlayerID = snapshot.White.ID
		expectedResult = game.BlackWins
	} else {
		if snapshot.Black == nil {
			t.Fatal("expected black player")
		}

		activePlayerID = snapshot.Black.ID
		expectedResult = game.WhiteWins
	}

	var opponentConn *websocket.Conn

	switch activePlayerID {
	case player1Session.PlayerID:
		opponentConn = conn2

	case player2Session.PlayerID:
		opponentConn = conn1

	default:
		t.Fatal("active player does not match either session")
	}

	// Start the real room clock watcher.
	//
	// We are intentionally not calling CheckTimeout() directly.
	// This exercises:
	//
	// ticker -> CheckTimeout -> Finish -> NotifyFinished(true)
	//        -> server callback -> broadcastState
	room.StartClockWatcher()

	// The watcher runs asynchronously.
	_ = opponentConn.SetReadDeadline(time.Now().Add(2 * time.Second))

	finalState := readGameStateMessage(t, opponentConn)

	if finalState.Status != game.Finished {
		t.Fatalf(
			"expected finished game, got %q",
			finalState.Status,
		)
	}

	if game.GameResult(finalState.Result) != expectedResult {
		t.Fatalf(
			"expected result %q, got %q",
			expectedResult,
			finalState.Result,
		)
	}

	if game.EndReason(finalState.EndReason) != game.Timeout {
		t.Fatalf(
			"expected end reason %q, got %q",
			game.Timeout,
			finalState.EndReason,
		)
	}

	// Verify the server-side state too.
	finalSnapshot := room.Snapshot()

	if finalSnapshot.Status != game.Finished {
		t.Fatalf(
			"expected room to be finished, got %q",
			finalSnapshot.Status,
		)
	}

	if game.GameResult(finalSnapshot.Result) != expectedResult {
		t.Fatalf(
			"expected room result %q, got %q",
			expectedResult,
			finalSnapshot.Result,
		)
	}

	if game.EndReason(finalSnapshot.EndReason) != game.Timeout {
		t.Fatalf(
			"expected room end reason %q, got %q",
			game.Timeout,
			finalSnapshot.EndReason,
		)
	}
}