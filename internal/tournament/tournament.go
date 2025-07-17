package tournament

import (
	"context"
	"fmt"
	"log/slog"
	"tournament-manager/internal/database"
)

type Tournament struct {
	ID           string
	Name         string
	Date         uint64
	Format       string
	Participants []Player
}

var AvailableFormats = map[string]string{
	"solo_single_elim": "Solo Single Elimination",
}

func CreateTournament(name string, date uint64, format string) (string, error) {
	slog.Debug("inserting values", "name", name, "date", date, "format", format)

	if _, exists := AvailableFormats[format]; !exists {
		return "", fmt.Errorf("unsupported format: %s", format)
	}

	insertQuery := "INSERT INTO Tournament (name, date, format) VALUES ($1, $2, $3) RETURNING id"

	var id string
	err := database.DB.QueryRow(context.Background(), insertQuery, name, date, format).Scan(&id)
	if err != nil {
		slog.Warn(err.Error())
		return "", err
	}

	return id, nil
}
