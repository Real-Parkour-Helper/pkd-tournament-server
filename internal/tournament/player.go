package tournament

import (
	"context"
	"log/slog"
	"tournament-manager/internal/database"
)

type Player struct {
	IGN          string
	DiscordName  string
	PersonalBest int64
}

func Signup(ign string, discord string, pb int64, tournament_id string) error {
	slog.Debug("inserting values", "ign", ign, "discord", discord, "pb", pb, "tournament_id", tournament_id)
	insertQuery := "INSERT INTO Player (ign, discord_name, personal_best, tournament_id) VALUES ($1, $2, $3, $4)"

	if _, err := database.DB.Exec(context.Background(), insertQuery, ign, discord, pb, tournament_id); err != nil {
		slog.Warn(err.Error())
		return err
	}

	return nil
}
