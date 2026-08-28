package game

import "sync"

type Room struct {
	mu   sync.Mutex
	Game *Game
}

func NewRoom(game *Game) *Room {
	return &Room{
		Game: game,
	}
}

func (r *Room) AddPlayer(player *Player) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.Game.AddPlayer(player)
}

func (r *Room) RemovePlayer(playerID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.Game.RemovePlayer(playerID)
}

func (r *Room) Player(playerID string) *Player {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.Game.GetPlayer(playerID)
}

func (r *Room) Start() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.Game.IsFull() {
		return false
	}

	if r.Game.Status != Waiting {
		return false
	}

	r.Game.Status = Playing
	return true
}

func (r *Room) Move(playerID, from, to string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	player := r.Game.GetPlayer(playerID)
	if player == nil {
		return ErrPlayerNotFound
	}

	return r.Game.Move(player, from, to)
}

func (r *Room) Players() []*Player {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	players := make([]*Player, 0, 2)

	if r.Game.White != nil {
		players = append(players, r.Game.White)
	}

	if r.Game.Black != nil {
		players = append(players, r.Game.Black)
	}

	return players
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
	return r.Game.ID
}