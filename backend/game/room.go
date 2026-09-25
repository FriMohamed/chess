package game

import (
	"sync"
	"time"
)

const (
	connectionTimeout  = 30 * time.Second
	privateWaitTimeout = 5 * time.Minute
	clockCheckInterval = 100 * time.Millisecond
)

type RoomType string

const (
	QuickRoom   RoomType = "quick"
	PrivateRoom RoomType = "private"
)

type Room struct {
	mu sync.Mutex

	Game *Game

	Type RoomType
	Code string

	connected map[string]bool
	timers    map[string]*time.Timer

	waitTimer *time.Timer

	onFinished func(broadcast bool)
}

func NewQuickRoom(game *Game) *Room {
	return &Room{
		Game:      game,
		Type:      QuickRoom,
		connected: make(map[string]bool),
		timers:    make(map[string]*time.Timer),
	}
}

func NewPrivateRoom(game *Game, code string) *Room {
	return &Room{
		Game:      game,
		Type:      PrivateRoom,
		Code:      code,
		connected: make(map[string]bool),
		timers:    make(map[string]*time.Timer),
	}
}

func (r *Room) StartWaitingTimer(onTimeout func()) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.waitTimer = time.AfterFunc(privateWaitTimeout, onTimeout)
}

func (r *Room) CancelWaitingTimer() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.waitTimer != nil {
		r.waitTimer.Stop()
		r.waitTimer = nil
	}
}

func (r *Room) AddPlayer(player *Player) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Game.Status != Waiting {
		return false
	}

	added := r.Game.AddPlayer(player)

	return added
}

func (r *Room) NotifyFinished(broadcast bool) {
	r.mu.Lock()
	callback := r.onFinished
	r.mu.Unlock()

	if callback == nil {
		return
	}

	callback(broadcast)
}

func (r *Room) PlayerIDs() []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	ids := make([]string, 0, 2)

	for _, player := range r.Game.Players {
		if player != nil {
			ids = append(ids, player.ID)
		}
	}

	return ids
}

// TryStart registers the finish callback and starts the game atomically.
// This prevents the game from becoming Playing before the callback exists.
func (r *Room) TryStart(onFinished func(broadcast bool)) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Game.Status != Waiting {
		return false
	}

	if !r.Game.IsFull() {
		return false
	}

	for _, player := range r.Game.Players {
		if player == nil || !r.connected[player.ID] {
			return false
		}
	}

	r.onFinished = onFinished

	started := r.Game.Start()

	return started
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

func (r *Room) Resign(playerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	player := r.Game.GetPlayer(playerID)
	if player == nil {
		return ErrPlayerNotFound
	}

	return r.Game.Resign(player)
}

func (r *Room) OfferDraw(playerID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	player := r.Game.GetPlayer(playerID)
	if player == nil {
		return ErrPlayerNotFound
	}

	return r.Game.OfferDraw(player)
}

func (r *Room) RespondDraw(playerID string, accepted bool) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	player := r.Game.GetPlayer(playerID)
	if player == nil {
		return ErrPlayerNotFound
	}

	return r.Game.RespondDraw(player, accepted)
}

func (r *Room) Snapshot() GameSnapshot {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.Game.Snapshot()
}

func (r *Room) Player(playerID string) *Player {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.Game.GetPlayer(playerID)
}

func (r *Room) Opponent(playerID string) *Player {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Game.White() != nil && r.Game.White().ID == playerID {
		return r.Game.Black()
	}

	if r.Game.Black() != nil && r.Game.Black().ID == playerID {
		return r.Game.White()
	}

	return nil
}

func (r *Room) StartConnectionTimer(playerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if timer, ok := r.timers[playerID]; ok {
		timer.Stop()
	}

	var timer *time.Timer

	timer = time.AfterFunc(
		connectionTimeout,
		func() {
			r.connectionTimeout(playerID, timer)
		},
	)

	r.timers[playerID] = timer
}

func (r *Room) PlayerConnected(playerID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Game.GetPlayer(playerID) == nil {
		return false
	}

	r.connected[playerID] = true

	if timer, ok := r.timers[playerID]; ok {
		timer.Stop()
		delete(r.timers, playerID)
	}

	return true
}

func (r *Room) PlayerDisconnected(playerID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.connected, playerID)

	// A finished game no longer needs a reconnect timer.
	if r.Game.Status == Finished {
		return
	}

	if timer, ok := r.timers[playerID]; ok {
		timer.Stop()
	}

	var timer *time.Timer

	timer = time.AfterFunc(
		connectionTimeout,
		func() {
			r.connectionTimeout(playerID, timer)
		},
	)

	r.timers[playerID] = timer
}

func (r *Room) connectionTimeout(playerID string, timer *time.Timer) {
	r.mu.Lock()

	// This timer is no longer the active timer for this player.
	if current, ok := r.timers[playerID]; !ok || current != timer {
		r.mu.Unlock()
		return
	}

	delete(r.timers, playerID)
	delete(r.connected, playerID)

	// Game hasn't started yet.
	if r.Game.Status == Waiting {
		r.Game.RemovePlayer(playerID)

		r.mu.Unlock()
		return
	}

	// Finished games do not need disconnect handling.
	if r.Game.Status == Finished {
		r.mu.Unlock()
		return
	}

	// Game is already playing.
	finished := r.Game.Finish(
		playerID,
		EndReasonDisconnect,
	)

	r.mu.Unlock()

	if finished {
		r.NotifyFinished(true)
	}
}

func (r *Room) QuitGame(playerID string) bool {
	r.mu.Lock()

	if timer, ok := r.timers[playerID]; ok {
		timer.Stop()
		delete(r.timers, playerID)
	}

	delete(r.connected, playerID)

	player := r.Game.GetPlayer(playerID)
	if player == nil {
		r.mu.Unlock()
		return false
	}

	// Game hasn't started.
	if r.Game.Status == Waiting {
		removed := r.Game.RemovePlayer(playerID)

		r.mu.Unlock()

		return removed
	}

	// Already finished.
	if r.Game.Status == Finished {
		r.mu.Unlock()
		return false
	}

	// Game is already playing.
	finished := r.Game.Finish(
		playerID,
		EndReasonQuit,
	)

	r.mu.Unlock()

	if finished {
		// HTTP quit has no WS handler to broadcast the final state.
		r.NotifyFinished(true)
	}

	return finished
}

func (r *Room) IsEmpty() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.Game.IsEmpty()
}

func (r *Room) StartClockWatcher() {
	go func() {

		ticker := time.NewTicker(clockCheckInterval)
		defer ticker.Stop()

		for range ticker.C {
			r.mu.Lock()

			if r.Game.Status != Playing {
				r.mu.Unlock()
				return
			}

			finished := r.Game.CheckTimeout()

			r.mu.Unlock()

			if finished {
				r.NotifyFinished(true)
				return
			}
		}
	}()
}
