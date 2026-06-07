package main

import (
	"log"

	"algoogle.ai/apps/api/internal/server"
)

func main() {
	cfg := server.LoadConfig()

	app, err := server.New(cfg)
	if err != nil {
		log.Fatalf("failed to start app: %v", err)
	}

	log.Printf("AlgoSensei API listening on :%s", cfg.Port)
	if err := app.Listen(); err != nil {
		log.Fatal(err)
	}
}
