package main

import (
	"log"
	"practice-7/config"
	"practice-7/internal/app"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load: %v", err)
	}
	app.Run(cfg)
}
