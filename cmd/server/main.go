package main

import (
	"log"
	"net/http"
	"os"

	"ai-japanese-learning/internal/app"
	"ai-japanese-learning/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("create app: %v", err)
	}
	defer application.Close()

	addr := cfg.ServerAddress
	if port := os.Getenv("PORT"); port != "" {
		addr = ":" + port
	}

	log.Printf("server listening on %s", addr)
	if err := http.ListenAndServe(addr, application.Router()); err != nil {
		log.Fatalf("listen: %v", err)
	}
}
