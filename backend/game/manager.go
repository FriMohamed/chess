package game

import "sync"

type Manager struct {
	mu    sync.RWMutex
	rooms map[string]*Room
}

func NewManager() *Manager {
	return &Manager{
		rooms: make(map[string]*Room),
	}
}

func (m *Manager) QuickGame(player *Player) *Room {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, room := range m.rooms {
		if room.AddPlayer(player) {
			room.StartConnectionTimer(player.ID)
			return room
		}
	}

	game := NewGame(player)

	room := NewRoom(game)

	m.rooms[game.ID] = room

	room.StartConnectionTimer(player.ID)

	return room
}

func (m *Manager) OpenGames() []*Game {
	m.mu.RLock()
	defer m.mu.RUnlock()

	games := make([]*Game, 0)

	for _, room := range m.rooms {
		// if room.IsOpen() {
			games = append(games, room.Game)
		// }
	}

	return games
}

func (m *Manager) GetRoom(gameID string) *Room {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.rooms[gameID]
}

func (m *Manager) RemoveRoom(gameID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.rooms[gameID]; !exists {
		return false
	}

	delete(m.rooms, gameID)
	return true
}
