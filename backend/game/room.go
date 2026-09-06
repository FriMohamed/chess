package game

import (
	"sync"
	"time"
)

const connectionTimeout = 30 * time.Second

type Room struct {
	mu sync.Mutex

	Game *Game

	connected map[string]bool
	timers    map[string]*time.Timer
}

func NewRoom(game *Game) *Room {
	return &Room{
		Game:      game,
		connected: make(map[string]bool),
		timers:    make(map[string]*time.Timer),
	}
}

func (r *Room) AddPlayer(player *Player) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.Game.AddPlayer(player)
}

func (r *Room) Player(playerID string) *Player {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.Game.GetPlayer(playerID)
}

func (r *Room) Start() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.Game.Start()
	return true
}

func (r *Room) Move(playerID, notation string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	player := r.Game.GetPlayer(playerID)
	if player == nil {
		return ErrPlayerNotFound
	}

	return r.Game.Move(player, notation)
}

func (r *Room) Players() [2]*Player {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.Game.Players
}

func (r *Room) IsEmpty() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.Game.IsEmpty()
}

func (r *Room) IsFull() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.Game.IsFull()
}

func (r *Room) GameID() string {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.Game.ID
}

func (r *Room) StartConnectionTimer(playerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if timer, ok := r.timers[playerID]; ok {
		timer.Stop()
	}

	r.timers[playerID] = time.AfterFunc(
		connectionTimeout,
		func() {
			r.connectionTimeout(playerID)
		},
	)
}

func (r *Room) PlayerConnected(playerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.connected[playerID] = true

	if timer, ok := r.timers[playerID]; ok {
		timer.Stop()
		delete(r.timers, playerID)
	}
	println("Player", playerID, "re/connected. Stopping connection timer.")
}

func (r *Room) PlayerDisconnected(playerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.connected, playerID)
	println("Player", playerID, "disconnected.")
	r.timers[playerID] = time.AfterFunc(
		connectionTimeout,
		func() {
			r.connectionTimeout(playerID)
		},
	)
}

func (r *Room) connectionTimeout(playerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.timers, playerID)
	println("Player", playerID, "connection timeout. Removing from game.")
	r.Game.RemovePlayer(playerID)
}
