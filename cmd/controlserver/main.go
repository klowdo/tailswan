package main

import (
	"embed"
	"log"

	"github.com/klowdo/tailswan/internal/config"
	"github.com/klowdo/tailswan/internal/server"
)

//go:embed web
var webFS embed.FS

func main() {
	cfg := config.Load()

	srv, err := server.New(cfg, webFS)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	defer srv.Close()

	if err := srv.Start(); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
