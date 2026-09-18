package game

type GameSnapshot struct {
	ID         string
	White      *Player
	Black      *Player
	FEN        string
	WhiteTime  int64
	BlackTime  int64
	Active     ActiveColor
	Status     GameStatus
	Result     GameResult
	EndReason  EndReason
	Check      bool
}