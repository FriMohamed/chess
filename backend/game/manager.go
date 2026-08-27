package game

import (
	"sync"
)

type Manager struct {
	mu    sync.Mutex
	rooms map[string]*GameRoom
}

func NewManager() *Manager {
	return &Manager{
		rooms: make(map[string]*GameRoom),
	}
}

func (m *Manager) QuickGame(player *Player) *GameRoom {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, r := range m.rooms {
		if r.Game.AddPlayer(player) {
			r.startPlayerTimer(player)
			return r
		}
	}

	game := NewGame(player)

	room := NewGameRoom(game, func() {
		m.RemoveRoom(game.ID)
	})

	m.rooms[game.ID] = room
	room.startPlayerTimer(player)

	return room
}

func (m *Manager) GetRoom(id string) *GameRoom {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.rooms[id]
}

func (m *Manager) RemoveRoom(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.rooms, id)
}
