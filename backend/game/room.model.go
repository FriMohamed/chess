package game

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type GameRoom struct {
	mu      sync.Mutex
	Game    *Game
	Clients [2]*Client
	timers  map[string]*time.Timer
	onEmpty func()
}

func NewGameRoom(game *Game, onEmpty func()) *GameRoom {
	return &GameRoom{
		Game:    game,
		Clients: [2]*Client{},
		timers:  make(map[string]*time.Timer),
		onEmpty: onEmpty,
	}
}

func (r *GameRoom) AddClient(client *Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i := range r.Clients {
		if r.Clients[i] == nil {
			r.Clients[i] = client
			break
		}
	}

	if timer, ok := r.timers[client.Player.ID]; ok {
		timer.Stop()
		delete(r.timers, client.Player.ID)
	}

	if r.Clients[0] != nil && r.Clients[1] != nil {
		r.startGame()
	}

	go client.Listen(r)
}

func (r *GameRoom) startPlayerTimer(player *Player) {
	timer := time.AfterFunc(30*time.Second, func() {
		r.playerTimeout(player.ID)
	})

	r.timers[player.ID] = timer
}

func (r *GameRoom) playerTimeout(playerID string) {
	r.mu.Lock()

	delete(r.timers, playerID)

	if r.hasClientForPlayer(playerID) {
		r.mu.Unlock()
		return
	}

	r.Game.RemovePlayer(playerID)

	empty := r.Game.isEmpty()

	r.mu.Unlock()

	if empty {
		r.onEmpty()
	}
}

func (r *GameRoom) hasClientForPlayer(playerID string) bool {
	for _, client := range r.Clients {
		if client == nil {
			continue
		}

		if client.Player.ID == playerID {
			return true
		}
	}

	return false
}

func (r *GameRoom) HandleMessage(client *Client, message Message) {
	switch message.Type {
	case MessageMove:
		r.handleMove(client, message)

	default:
		client.Send(Message{
			Type: MessageError,
		})
	}
}

func (r *GameRoom) handleMove(client *Client, message Message) {
	var command MoveCommand

	if err := json.Unmarshal(message.Data, &command); err != nil {
		client.Send(Message{
			Type: MessageError,
		})
		return
	}

	fmt.Printf(
		"player %s wants to move %s -> %s\n",
		client.Player.Nickname,
		command.From,
		command.To,
	)
}

func (r *GameRoom) startGame() {
	r.Game.Status = Playing

	r.Broadcast(Message{
		Type: MessageGameStarted,
		// Data:,
	})
}

func (r *GameRoom) Broadcast(message Message) {
	for _, client := range r.Clients {
		if client == nil {
			continue
		}

		if err := client.Send(message); err != nil {
			// For now, just ignore it.
			// We'll handle disconnects properly later.
		}
	}
}
