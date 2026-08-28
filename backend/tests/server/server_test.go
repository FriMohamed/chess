package server_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"chess-backend/game"
	"chess-backend/server"

	"github.com/gorilla/websocket"
)

type quickGameResponse struct {
	GameID   string `json:"gameId"`
	PlayerID string `json:"playerId"`
}

type wsMessage struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

type testGameState struct {
	GameID string       `json:"gameId"`
	White  *game.Player `json:"white"`
	Black  *game.Player `json:"black"`
	FEN    string       `json:"fen"`
}

func newTestServer() *httptest.Server {
	manager := game.NewManager()
	srv := server.New(manager)

	return httptest.NewServer(srv.Handler())
}

func quickGame(
	t *testing.T,
	baseURL string,
	nickname string,
) quickGameResponse {
	t.Helper()

	request := struct {
		Nickname string `json:"nickname"`
	}{
		Nickname: nickname,
	}

	body, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("failed to encode quick game request: %v", err)
	}

	resp, err := http.Post(
		baseURL+"/games/quick",
		"application/json",
		strings.NewReader(string(body)),
	)
	if err != nil {
		t.Fatalf("quick game request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var result quickGameResponse

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if result.GameID == "" {
		t.Fatal("expected game ID")
	}

	if result.PlayerID == "" {
		t.Fatal("expected player ID")
	}

	return result
}

func createPlayers(
	t *testing.T,
	baseURL string,
) (quickGameResponse, quickGameResponse) {
	t.Helper()

	player1 := quickGame(t, baseURL, "Alice")
	player2 := quickGame(t, baseURL, "Bob")

	if player1.GameID != player2.GameID {
		t.Fatal("expected both players to get the same game")
	}

	return player1, player2
}

func connectWebSocket(
	t *testing.T,
	baseURL string,
	gameID string,
	playerID string,
) *websocket.Conn {
	t.Helper()

	wsURL := "ws" + strings.TrimPrefix(baseURL, "http") +
		"/games/" + gameID + "/ws?playerId=" + playerID

	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("WebSocket connection failed: %v", err)
	}

	return conn
}

func connectPlayers(
	t *testing.T,
	baseURL string,
	player1 quickGameResponse,
	player2 quickGameResponse,
) (*websocket.Conn, *websocket.Conn) {
	t.Helper()

	conn1 := connectWebSocket(
		t,
		baseURL,
		player1.GameID,
		player1.PlayerID,
	)

	conn2 := connectWebSocket(
		t,
		baseURL,
		player2.GameID,
		player2.PlayerID,
	)

	return conn1, conn2
}

func readMessage(
	t *testing.T,
	conn *websocket.Conn,
) wsMessage {
	t.Helper()

	_ = conn.SetReadDeadline(time.Now().Add(2 * time.Second))

	_, data, err := conn.ReadMessage()
	if err != nil {
		t.Fatalf("failed to read WebSocket message: %v", err)
	}

	var message wsMessage

	if err := json.Unmarshal(data, &message); err != nil {
		t.Fatalf("failed to decode WebSocket message: %v", err)
	}

	return message
}

func startGame(
	t *testing.T,
	conn1 *websocket.Conn,
	conn2 *websocket.Conn,
) {
	t.Helper()

	message1 := readMessage(t, conn1)
	message2 := readMessage(t, conn2)

	if message1.Type != "game_started" {
		t.Fatalf(
			"player 1: expected game_started, got %s",
			message1.Type,
		)
	}

	if message2.Type != "game_started" {
		t.Fatalf(
			"player 2: expected game_started, got %s",
			message2.Type,
		)
	}
}

func sendMove(
	t *testing.T,
	conn *websocket.Conn,
	from string,
	to string,
) {
	t.Helper()

	message := wsMessage{
		Type: "move",
		Data: json.RawMessage(
			`{"from":"` + from + `","to":"` + to + `"}`,
		),
	}

	if err := conn.WriteJSON(message); err != nil {
		t.Fatalf("failed to send move: %v", err)
	}
}

func assertError(
	t *testing.T,
	conn *websocket.Conn,
	expectedCode string,
) {
	t.Helper()

	message := readMessage(t, conn)

	if message.Type != "error" {
		t.Fatalf(
			"expected error message, got %s",
			message.Type,
		)
	}

	var data struct {
		Code string `json:"code"`
	}

	if err := json.Unmarshal(message.Data, &data); err != nil {
		t.Fatalf("failed to decode error: %v", err)
	}

	if data.Code != expectedCode {
		t.Fatalf(
			"expected error code %q, got %q",
			expectedCode,
			data.Code,
		)
	}
}

func TestQuickGame(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	player := quickGame(t, ts.URL, "Alice")

	if player.GameID == "" {
		t.Fatal("expected game ID")
	}

	if player.PlayerID == "" {
		t.Fatal("expected player ID")
	}
}

func TestWebSocketGameStart(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	player1, player2 := createPlayers(t, ts.URL)

	conn1, conn2 := connectPlayers(
		t,
		ts.URL,
		player1,
		player2,
	)
	defer conn1.Close()
	defer conn2.Close()

	message1 := readMessage(t, conn1)
	message2 := readMessage(t, conn2)

	if message1.Type != "game_started" {
		t.Fatalf(
			"player 1: expected game_started, got %s",
			message1.Type,
		)
	}

	if message2.Type != "game_started" {
		t.Fatalf(
			"player 2: expected game_started, got %s",
			message2.Type,
		)
	}

	var state testGameState

	if err := json.Unmarshal(message1.Data, &state); err != nil {
		t.Fatalf("failed to decode game state: %v", err)
	}

	if state.GameID != player1.GameID {
		t.Fatalf(
			"expected game ID %s, got %s",
			player1.GameID,
			state.GameID,
		)
	}

	if state.White == nil {
		t.Fatal("expected white player")
	}

	if state.Black == nil {
		t.Fatal("expected black player")
	}

	if state.FEN == "" {
		t.Fatal("expected FEN")
	}
}

func TestWebSocketMove(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	player1, player2 := createPlayers(t, ts.URL)

	conn1, conn2 := connectPlayers(
		t,
		ts.URL,
		player1,
		player2,
	)
	defer conn1.Close()
	defer conn2.Close()

	startGame(t, conn1, conn2)

	sendMove(t, conn1, "e2", "e4")

	message1 := readMessage(t, conn1)
	message2 := readMessage(t, conn2)

	if message1.Type != "game_state" {
		t.Fatalf(
			"player 1: expected game_state, got %s",
			message1.Type,
		)
	}

	if message2.Type != "game_state" {
		t.Fatalf(
			"player 2: expected game_state, got %s",
			message2.Type,
		)
	}

	var state testGameState

	if err := json.Unmarshal(message1.Data, &state); err != nil {
		t.Fatalf("failed to decode game state: %v", err)
	}

	if state.FEN == "" {
		t.Fatal("expected FEN after move")
	}
}

func TestWebSocketInvalidMessage(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	player := quickGame(t, ts.URL, "Alice")

	conn := connectWebSocket(
		t,
		ts.URL,
		player.GameID,
		player.PlayerID,
	)
	defer conn.Close()

	if err := conn.WriteMessage(
		websocket.TextMessage,
		[]byte(`not valid json`),
	); err != nil {
		t.Fatalf("failed to send invalid message: %v", err)
	}

	assertError(t, conn, "invalid_message")
}

func TestWebSocketUnknownMessageType(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	player1, player2 := createPlayers(t, ts.URL)

	conn1, conn2 := connectPlayers(
		t,
		ts.URL,
		player1,
		player2,
	)
	defer conn1.Close()
	defer conn2.Close()

	startGame(t, conn1, conn2)

	message := wsMessage{
		Type: "something_unknown",
	}

	if err := conn1.WriteJSON(message); err != nil {
		t.Fatalf("failed to send message: %v", err)
	}

	assertError(t, conn1, "invalid_message")
}

func TestWebSocketInvalidMoveCommand(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	player1, player2 := createPlayers(t, ts.URL)

	conn1, conn2 := connectPlayers(
		t,
		ts.URL,
		player1,
		player2,
	)
	defer conn1.Close()
	defer conn2.Close()

	startGame(t, conn1, conn2)

	message := wsMessage{
		Type: "move",
		Data: json.RawMessage(`{"from":123,"to":true}`),
	}

	if err := conn1.WriteJSON(message); err != nil {
		t.Fatalf("failed to send move: %v", err)
	}

	assertError(t, conn1, "invalid_message")
}

func TestWebSocketNotYourTurn(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	player1, player2 := createPlayers(t, ts.URL)

	conn1, conn2 := connectPlayers(
		t,
		ts.URL,
		player1,
		player2,
	)
	defer conn1.Close()
	defer conn2.Close()

	startGame(t, conn1, conn2)

	sendMove(t, conn2, "e7", "e5")

	assertError(t, conn2, "not_your_turn")
}

func TestWebSocketInvalidMove(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	player1, player2 := createPlayers(t, ts.URL)

	conn1, conn2 := connectPlayers(
		t,
		ts.URL,
		player1,
		player2,
	)
	defer conn1.Close()
	defer conn2.Close()

	startGame(t, conn1, conn2)

	sendMove(t, conn1, "e2", "e5")

	assertError(t, conn1, "invalid_move")
}
