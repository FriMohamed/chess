package game

import "time"

type ActiveColor int

const (
	White ActiveColor = 0
	Black ActiveColor = 1
)

type GameClock struct {
	TimeLeft   [2]time.Duration
	Active     ActiveColor
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


func (c *GameClock) updateElapsed() {
	now := time.Now()

	elapsed := now.Sub(c.LastUpdate)

	c.TimeLeft[c.Active] -= elapsed
	c.LastUpdate = now

	if c.TimeLeft[c.Active] < 0 {
		c.TimeLeft[c.Active] = 0
	}
}

func (c *GameClock) switchClock() {
	if c.Active == White {
		c.Active = Black
	} else {
		c.Active = White
	}
}


func (c *GameClock) updateClock() {
	c.updateElapsed()
	c.switchClock()
}