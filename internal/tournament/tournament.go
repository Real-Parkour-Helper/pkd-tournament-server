package tournament

import (
	"context"
	"log/slog"
	"tournament-manager/internal/database"
)

type Tournament struct {
	ID           int64
	Name         string
	Time         uint64
	Participants []Player
	Format       Format
}

type Format struct {
	Name    string
	Handler func([]Player, []uint64)
}

var AvailableFormats = map[string]Format{
	"double_elim": {
		Name:    "Double Elimination",
		Handler: func(p []Player, u []uint64) {},
	},
}

func CreateTournament(name string, time uint64, format string) error {
	slog.Debug("inserting values", "name", name, "time", time, "format", format)
	insertQuery := "INSERT INTO Tournament (name, date, format) VALUES ($1, $2, $3)"

	if _, err := database.DB.Exec(context.Background(), insertQuery, name, time, format); err != nil {
		slog.Warn(err.Error())
		return err
	}

	return nil
}
