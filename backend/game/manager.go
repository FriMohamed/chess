package game

import (
	"fmt"
	"log"
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

			log.Printf(
				"[ROOM] player joined game=%s player=%s type=quick",
				room.Game.ID,
				player.ID,
			)

			return room
		}
	}

	game := NewGame(player)
	room := NewQuickRoom(game)

	m.rooms[game.ID] = room

	room.StartConnectionTimer(player.ID)

	log.Printf(
		"[ROOM] quick game created game=%s player=%s",
		game.ID,
		player.ID,
	)

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

	log.Printf(
		"[ROOM] private game created game=%s code=%s player=%s",
		game.ID,
		code,
		player.ID,
	)

	return room
}

func (m *Manager) JoinPrivateGame(code string, player *Player, logger *log.Logger) (*Room, bool) {
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

		logger.Printf(
			"[ROOM] player joined private game=%s code=%s player=%s",
			room.Game.ID,
			code,
			player.ID,
		)

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

func (m *Manager) RemovePlayer(gameID string, playerID string) bool {
	m.mu.RLock()

	room, exists := m.rooms[gameID]

	m.mu.RUnlock()

	if !exists {
		return false
	}

	removed := room.QuitGame(playerID)

    // Delete room if no players remain
    if removed && room.IsEmpty() {
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

	log.Printf(
		"[MANAGER] room cleanup scheduled game=%s delay=%s",
		gameID,
		cleanupDelay,
	)

	time.AfterFunc(cleanupDelay, func() {
		log.Printf(
			"[MANAGER] room cleanup started game=%s",
			gameID,
		)

		m.RemoveRoom(gameID)

		log.Printf(
			"[MANAGER] room cleanup completed game=%s",
			gameID,
		)
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
