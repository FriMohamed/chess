package game

import (
	"encoding/json"
	"sync"

	"github.com/gorilla/websocket"
)

type MessageType string

const (
	MessageGameStarted MessageType = "game_started"
	MessageMove        MessageType = "move"
	MessageError       MessageType = "error"
)

type Client struct {
	Player *Player
	Conn   *websocket.Conn

	writeMu sync.Mutex
}

type Message struct {
	Type MessageType     `json:"type"`
	Data json.RawMessage `json:"data,omitempty"`
}

type MoveCommand struct {
	From string `json:"from"`
	To   string `json:"to"`
}

func (c *Client) Listen(room *GameRoom) {
	defer func() {
		c.Conn.Close()
		// room.RemoveClient(c)
	}()

	for {
		_, data, err := c.Conn.ReadMessage()
		if err != nil {
			return
		}

		var message Message

		if err := json.Unmarshal(data, &message); err != nil {
			c.Send(Message{
				Type: MessageError,
			})
			continue
		}

		room.HandleMessage(c, message)
	}
}

func (c *Client) Send(message Message) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	return c.Conn.WriteJSON(message)
}

// func (c *Client) SendError(message string) error {
// 	return c.Conn.WriteJSON(Message{
// 		Type: MessageError,
// 		Data: json.RawMessage(
// 			fmt.Sprintf(`{"message":%q}`, message),
// 		),
// 	})
// }
