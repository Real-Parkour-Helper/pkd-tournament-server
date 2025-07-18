package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"tournament-manager/internal/tournament"
	"tournament-manager/internal/util"
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

	pb, err := util.ParseTime(body.PersonalBest)
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
