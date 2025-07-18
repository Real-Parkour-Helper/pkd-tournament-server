package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"tournament-manager/internal/tournament"

	"github.com/gorilla/mux"
)

type StartTournamentRequest struct {
	TournamentID string `json:"tournament_id"`
}

type GameResultRequest struct {
	GameID  string   `json:"game_id"`
	Players []string `json:"players"`
	Times   []string `json:"times"`
}

func StartTournament(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tournamentID := vars["id"]

	if tournamentID == "" {
		http.Error(w, "tournament ID is required", http.StatusBadRequest)
		return
	}

	err := tournament.Manager.StartTournament(tournamentID)
	if err != nil {
		slog.Warn("Failed to start tournament", "tournament_id", tournamentID, "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"message":       "Tournament started successfully",
		"tournament_id": tournamentID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func SubmitGameResult(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tournamentID := vars["id"]

	if tournamentID == "" {
		http.Error(w, "tournament ID is required", http.StatusBadRequest)
		return
	}

	var req GameResultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.GameID == "" {
		http.Error(w, "game_id is required", http.StatusBadRequest)
		return
	}

	if len(req.Players) == 0 {
		http.Error(w, "players array cannot be empty", http.StatusBadRequest)
		return
	}

	if len(req.Players) != len(req.Times) {
		http.Error(w, "players and times arrays must have the same length", http.StatusBadRequest)
		return
	}

	err := tournament.Manager.HandleGameResult(tournamentID, req.GameID, req.Players, req.Times)
	if err != nil {
		slog.Warn("Failed to handle game result", "tournament_id", tournamentID, "game_id", req.GameID, "error", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := map[string]interface{}{
		"message":       "Game result processed successfully",
		"tournament_id": tournamentID,
		"game_id":       req.GameID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func GetTournamentStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tournamentID := vars["id"]

	if tournamentID == "" {
		http.Error(w, "tournament ID is required", http.StatusBadRequest)
		return
	}

	status, err := tournament.Manager.GetTournamentStatus(tournamentID)
	if err != nil {
		slog.Warn("Failed to get tournament status", "tournament_id", tournamentID, "error", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

func GetNextMatches(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tournamentID := vars["id"]

	if tournamentID == "" {
		http.Error(w, "tournament ID is required", http.StatusBadRequest)
		return
	}

	matches, err := tournament.Manager.GetNextMatches(tournamentID)
	if err != nil {
		slog.Warn("Failed to get next matches", "tournament_id", tournamentID, "error", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"tournament_id": tournamentID,
		"matches":       matches,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func GetTournamentBracket(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tournamentID := vars["id"]

	if tournamentID == "" {
		http.Error(w, "tournament ID is required", http.StatusBadRequest)
		return
	}

	bracket, err := tournament.Manager.GetBracket(tournamentID)
	if err != nil {
		slog.Warn("Failed to get tournament bracket", "tournament_id", tournamentID, "error", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"tournament_id": tournamentID,
		"bracket":       bracket,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func StopTournament(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	tournamentID := vars["id"]

	if tournamentID == "" {
		http.Error(w, "tournament ID is required", http.StatusBadRequest)
		return
	}

	err := tournament.Manager.StopTournament(tournamentID)
	if err != nil {
		slog.Warn("Failed to stop tournament", "tournament_id", tournamentID, "error", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"message":       "Tournament stopped successfully",
		"tournament_id": tournamentID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func ListActiveTournaments(w http.ResponseWriter, r *http.Request) {
	tournaments := tournament.Manager.ListActiveTournaments()

	response := map[string]interface{}{
		"active_tournaments": tournaments,
		"count":              len(tournaments),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
