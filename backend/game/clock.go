package game

import "time"

const (
	White = 0
	Black = 1
)

type GameClock struct {
	TimeLeft   [2]time.Duration
	Active     int
	LastUpdate time.Time
}

func newGameClock() *GameClock {
	return &GameClock{
		TimeLeft: [2]time.Duration{
			10 * time.Minute,
			10 * time.Minute,
		},
		Active:     White,
		LastUpdate: time.Now(),
	}
}

func (c *GameClock) updateClock() {
	now := time.Now()

	elapsed := now.Sub(c.LastUpdate)

	c.TimeLeft[c.Active] -= elapsed
	c.LastUpdate = now
	c.switchClock()
}

func (c *GameClock) switchClock() {
	if c.Active == White {
		c.Active = Black
	} else {
		c.Active = White
	}
}