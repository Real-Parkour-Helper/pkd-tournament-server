package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"time"
	"tournament-manager/internal/tournament"
)

func CreateTournament(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		slog.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var body struct {
		Name   string `json:"name"`
		Time   string `json:"time"`
		Format string `json:"format"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		slog.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	ts, err := time.Parse("2006-01-02 15:04 MST", body.Time)
	if err != nil {
		slog.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	tsUint := uint64(ts.Unix())

	if err := validateCreateTournament(body.Name, tsUint, body.Format); err != nil {
		slog.Error(err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := tournament.CreateTournament(body.Name, tsUint, body.Format); err != nil {
		slog.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func validateCreateTournament(name string, ts uint64, format string) error {
	var err1, err2, err3 error
	if name == "" {
		err1 = fmt.Errorf("tournament name cannot be empty")
	}

	if ts <= uint64(time.Now().Unix()) {
		err2 = fmt.Errorf("tournament time cannot be in the past")
	}

	if _, ok := tournament.AvailableFormats[format]; !ok {
		err3 = fmt.Errorf("unknown tournament format: %v", format)
	}

	if err1 == nil && err2 == nil && err3 == nil {
		return nil
	}

	err := errors.New("invalid tournament data")
	if err1 != nil {
		err = fmt.Errorf("%v: %v", err, err1)
	}

	if err2 != nil {
		err = fmt.Errorf("%v: %v", err, err2)
	}

	if err3 != nil {
		err = fmt.Errorf("%v: %v", err, err3)
	}
	return err
}
