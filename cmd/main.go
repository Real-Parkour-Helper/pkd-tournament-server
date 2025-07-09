package main

import (
	"log/slog"
	"os"
	"tournament-manager/internal/database"
	"tournament-manager/internal/server"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		AddSource: true,
		Level:     slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	if err := database.Init(); err != nil {
		slog.Error(err.Error())
		return
	}

	server.StartServer()
}
