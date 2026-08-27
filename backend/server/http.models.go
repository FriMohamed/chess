package server

type quickGameRequest struct {
	Nickname string `json:"nickname"`
}

type quickGameResponse struct {
	GameID   string `json:"gameId"`
	PlayerID string `json:"playerId"`
}