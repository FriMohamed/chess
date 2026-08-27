package main

import (
	"log"
	"net/http"

	"chess-backend/game"
	"chess-backend/server"
)

func main() {
	manager := game.NewManager()

	srv := server.New(manager)

	log.Println("server listening on :8080")

	err := http.ListenAndServe(":8080", srv.Handler())
	if err != nil {
		log.Fatal(err)
	}
}