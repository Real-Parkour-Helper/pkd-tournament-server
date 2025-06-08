package main

import (
	"log/slog"
	"os"
	"tournament-manager/internal/server"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
	}))
	slog.SetDefault(logger)

	server.StartServer()
}
