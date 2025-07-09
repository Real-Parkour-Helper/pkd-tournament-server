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
	"double_elim":      "Double Elimination",
}

func CreateTournament(name string, date uint64, format string) error {
	slog.Debug("inserting values", "name", name, "date", date, "format", format)

	// Validate format
	if _, exists := AvailableFormats[format]; !exists {
		return fmt.Errorf("unsupported format: %s", format)
	}

	insertQuery := "INSERT INTO Tournament (name, date, format) VALUES ($1, $2, $3)"

	if _, err := database.DB.Exec(context.Background(), insertQuery, name, date, format); err != nil {
		slog.Warn(err.Error())
		return err
	}

	return nil
}
