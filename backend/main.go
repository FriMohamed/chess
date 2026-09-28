package main

import (
	"log"
	"net/http"

	"chess-backend/game"
	"chess-backend/server"

	"github.com/rs/cors"

)

func main() {
	manager := game.NewManager()

	srv := server.New(manager)

	handler := cors.AllowAll().Handler(srv.Handler())


	log.Println("server listening on :8080")

	err := http.ListenAndServe(":8080", handler)
	if err != nil {
		log.Fatal(err)
	}
}
