package game

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

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
	if player == nil {
		return nil
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, room := range m.rooms {
		if room.Type != QuickRoom {
			continue
		}

		if room.AddPlayer(player) {
			room.StartConnectionTimer(player.ID)
			return room
		}
	}

	game := NewGame(player)
	room := NewQuickRoom(game)

	m.rooms[game.ID] = room

	room.StartConnectionTimer(player.ID)

	return room
}

func (m *Manager) CreatePrivateGame(player *Player) *Room {
	m.mu.Lock()
	defer m.mu.Unlock()

	code := m.newPrivateCode()

	game := NewGame(player)

	room := NewPrivateRoom(game, code)

	m.rooms[game.ID] = room

	room.StartConnectionTimer(player.ID)
	room.StartWaitingTimer(func() {
		m.RemoveRoom(game.ID)
	})

	return room
}

func (m *Manager) JoinPrivateGame(code string, player *Player) (*Room, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, room := range m.rooms {
		if room.Type != PrivateRoom || room.Code != code {
			continue
		}

		if !room.AddPlayer(player) {
			return nil, false
		}

		room.CancelWaitingTimer()

		room.StartConnectionTimer(player.ID)

		return room, true
	}

	return nil, false
}

func (m *Manager) newPrivateCode() string {
	for {
		code := fmt.Sprintf("%06d", rand.Intn(1000000))

		found := false

		for _, room := range m.rooms {
			if room.Type == PrivateRoom && room.Code == code {
				found = true
				break
			}
		}

		if !found {
			return code
		}
	}
}

func (m *Manager) RemovePlayer(gameID string, playerID string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()

	room, exists := m.rooms[gameID]
	if !exists {
		return false
	}

	removed := room.QuitGame(playerID)


	if room.IsEmpty() {
		delete(m.rooms, gameID)
	}

	return removed
}

func (m *Manager) GetRoom(gameID string) *Room {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.rooms[gameID]
}

func (m *Manager) ScheduleRoomCleanup(gameID string) {
	const cleanupDelay = 1 * time.Minute

	time.AfterFunc(cleanupDelay, func() {
		m.RemoveRoom(gameID)
	})
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
