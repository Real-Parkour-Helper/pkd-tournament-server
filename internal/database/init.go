package database

import (
	"context"
	"log/slog"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func connect() error {
	slog.Info("connecting to postgres", "database_url", os.Getenv("DATABASE_URL"))

	pool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		slog.Warn("failed to connect to database", "error", err.Error())
		return err
	}

	DB = pool
	return nil
}

func initTables() error {
	slog.Info("creating tables")

	sql := `
	CREATE TABLE IF NOT EXISTS Tournament (
	    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
	    name VARCHAR(100) NOT NULL,
	    date INT NOT NULL,
	    format VARCHAR(50)
	);

	CREATE TABLE IF NOT EXISTS Player (
	    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
	    ign VARCHAR(100) NOT NULL,
		discord_name VARCHAR(100) NOT NULL,
		tournament_id UUID NOT NULL,
		personal_best INT,
		FOREIGN KEY (tournament_id) REFERENCES Tournament(id)
	);

	CREATE TABLE IF NOT EXISTS GameResult (
	    game_id VARCHAR(20) NOT NULL,
	    tournament_id UUID NOT NULL,
	    player_id UUID NOT NULL,
	    position INT,
	    time INT,
	    FOREIGN KEY (tournament_id) REFERENCES Tournament(id),
	    FOREIGN KEY (player_id) REFERENCES Player(id)
	);`

	if _, err := DB.Exec(context.Background(), sql); err != nil {
		slog.Warn("failed to create tables", "error", err.Error())
		return err
	}

	return nil
}

func Init() error {
	if err := connect(); err != nil {
		slog.Warn(err.Error())
		return err
	}

	if err := initTables(); err != nil {
		slog.Warn(err.Error())
		return err
	}
	return nil
}
