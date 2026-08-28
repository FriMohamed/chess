package server

import (
	"sync"

	"github.com/gorilla/websocket"

	"chess-backend/game"
)

type Client struct {
	Player *game.Player
	Conn   *websocket.Conn

	writeMu sync.Mutex
}

func NewClient(player *game.Player, conn *websocket.Conn) *Client {
	return &Client{
		Player: player,
		Conn:   conn,
	}
}

func (c *Client) Send(message Message) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	return c.Conn.WriteJSON(message)
}

func (c *Client) Close() error {
	return c.Conn.Close()
}
