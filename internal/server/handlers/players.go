package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"tournament-manager/internal/tournament"
)

func Signup(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		slog.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var body struct {
		Ign          string `json:"ign"`
		DiscordName  string `json:"discord_name"`
		PersonalBest string `json:"personal_best"`
		TournamentID string `json:"tournament_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		slog.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	pb, err := parsePersonalBest(body.PersonalBest)
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := tournament.Signup(body.Ign, body.DiscordName, pb, body.TournamentID); err != nil {
		slog.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func parsePersonalBest(pb string) (int64, error) {
	pb = strings.TrimSpace(pb)
	minutes, rest, found := strings.Cut(pb, ":")
	if !found {
		err := errors.New("invalid personal best format: failed to parse minutes")
		slog.Warn(err.Error())
		return 0, err
	}

	seconds, rest, found := strings.Cut(rest, ".")
	if !found {
		err := errors.New("invalid personal best format: failed to parse seconds")
		slog.Warn(err.Error())
		return 0, err
	}

	minutesInt, err := strconv.ParseInt(minutes, 10, 64)
	if err != nil {
		slog.Warn(err.Error())
		return 0, err
	}

	secondsInt, err := strconv.ParseInt(seconds, 10, 64)
	if err != nil {
		slog.Warn(err.Error())
		return 0, err
	}

	if len(rest) == 0 {
		err := errors.New("invalid personal best format: failed to parse milliseconds")
		slog.Warn(err.Error())
		return 0, err
	}

	millisecondsInt, err := strconv.ParseInt(rest[0:1], 10, 64)
	if err != nil {
		slog.Warn(err.Error())
		return 0, err
	}

	return minutesInt*60*1000 + secondsInt*1000 + millisecondsInt, nil
}
