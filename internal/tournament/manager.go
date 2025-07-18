package tournament

import (
	"context"
	"fmt"
	"log/slog"
	"slices"
	"sync"
	"tournament-manager/internal/database"
	"tournament-manager/internal/tournament/formats"
	"tournament-manager/internal/util"
)

type TournamentManager struct {
	activeTournaments map[string]*formats.SoloSingleElimState
	mu                sync.RWMutex
}

var Manager *TournamentManager

func init() {
	Manager = &TournamentManager{
		activeTournaments: make(map[string]*formats.SoloSingleElimState),
	}
}

func (tm *TournamentManager) StartTournament(tournamentID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if _, exists := tm.activeTournaments[tournamentID]; exists {
		return fmt.Errorf("tournament %s is already active", tournamentID)
	}

	tournament, err := tm.getTournamentFromDB(tournamentID)
	if err != nil {
		return fmt.Errorf("failed to get tournament from database: %w", err)
	}

	players, err := tm.getPlayersForTournament(tournamentID)
	if err != nil {
		return fmt.Errorf("failed to get players for tournament: %w", err)
	}

	if len(players) < 2 {
		return fmt.Errorf("tournament needs at least 2 players, got %d", len(players))
	}

	switch tournament.Format {
	case "solo_single_elim":
		state := formats.NewSoloSingleElimState(tournamentID, players)
		if state == nil {
			return fmt.Errorf("failed to create tournament state")
		}
		tm.activeTournaments[tournamentID] = state
		slog.Info("Tournament started", "tournament_id", tournamentID, "format", tournament.Format, "players", len(players))
	default:
		return fmt.Errorf("unsupported tournament format: %s", tournament.Format)
	}

	return nil
}

func (tm *TournamentManager) HandleGameResult(tournamentID, gameID string, players []string, timesStr []string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	state, exists := tm.activeTournaments[tournamentID]
	if !exists {
		return fmt.Errorf("tournament %s is not active", tournamentID)
	}

	times := make([]uint64, len(timesStr))
	for i, timeStr := range timesStr {
		time, err := util.ParseTime(timeStr)
		if err != nil {
			return fmt.Errorf("failed to parse time for player %s: %w", players[i], err)
		}
		times[i] = time
	}

	type result struct {
		ign  string
		time uint64
		pos  int
	}
	results := make([]result, len(players))
	for i, player := range players {
		results[i] = result{ign: player, time: times[i]}
	}

	slices.SortFunc(results, func(a, b result) int {
		if a.time < b.time {
			return -1
		}
		if a.time > b.time {
			return 1
		}
		return 0
	})

	for i, res := range results {
		results[i].pos = i + 1
		sql := `
			INSERT INTO GameResult (game_id, tournament_id, player_id, position, time)
			VALUES ($1, $2, $3, $4, $5)
		`

		if i > 0 && res.time == results[i-1].time {
			results[i].pos = results[i-1].pos
		}

		playerID, err := getPlayerUUIDByIGN(tournamentID, res.ign)
		if err != nil {
			err = fmt.Errorf("failed to get player UUID for IGN %s: %w", res.ign, err)
			slog.Error(err.Error())
			return err
		}

		if _, err := database.DB.Exec(context.TODO(), sql, gameID, tournamentID, playerID, results[i].pos, res.time); err != nil {
			err = fmt.Errorf("failed to save game result for game %v: %v", res, err)
			slog.Error(err.Error())
			return err
		}
	}

	err := state.HandleGameResult(gameID, players, times)
	if err != nil {
		return fmt.Errorf("failed to handle game result: %w", err)
	}

	if state.IsComplete {
		slog.Info("Tournament completed", "tournament_id", tournamentID, "winner", state.Winner)

		if err := tm.saveTournamentResults(tournamentID, state); err != nil {
			slog.Warn("Failed to save tournament results", "tournament_id", tournamentID, "error", err)
		}

		delete(tm.activeTournaments, tournamentID)
	}

	return nil
}

func (tm *TournamentManager) GetTournamentState(tournamentID string) (*formats.SoloSingleElimState, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	state, exists := tm.activeTournaments[tournamentID]
	if !exists {
		return nil, fmt.Errorf("tournament %s is not active", tournamentID)
	}

	return state, nil
}

func (tm *TournamentManager) GetNextMatches(tournamentID string) ([]formats.Match, error) {
	state, err := tm.GetTournamentState(tournamentID)
	if err != nil {
		return nil, err
	}

	return state.GetNextMatches(), nil
}

func (tm *TournamentManager) GetTournamentStatus(tournamentID string) (map[string]interface{}, error) {
	state, err := tm.GetTournamentState(tournamentID)
	if err != nil {
		return nil, err
	}

	return state.GetTournamentStatus(), nil
}

func (tm *TournamentManager) GetBracket(tournamentID string) (string, error) {
	state, err := tm.GetTournamentState(tournamentID)
	if err != nil {
		return "", err
	}

	return state.GetBracketVisualization(), nil
}

func (tm *TournamentManager) IsActive(tournamentID string) bool {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	_, exists := tm.activeTournaments[tournamentID]
	return exists
}

func (tm *TournamentManager) StopTournament(tournamentID string) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if _, exists := tm.activeTournaments[tournamentID]; !exists {
		return fmt.Errorf("tournament %s is not active", tournamentID)
	}

	delete(tm.activeTournaments, tournamentID)
	slog.Info("Tournament stopped", "tournament_id", tournamentID)
	return nil
}

func (tm *TournamentManager) ListActiveTournaments() []string {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	tournaments := make([]string, 0, len(tm.activeTournaments))
	for id := range tm.activeTournaments {
		tournaments = append(tournaments, id)
	}
	return tournaments
}

func (tm *TournamentManager) getTournamentFromDB(tournamentID string) (*Tournament, error) {
	query := "SELECT id, name, date, format FROM Tournament WHERE id = $1"
	row := database.DB.QueryRow(context.Background(), query, tournamentID)

	var tournament Tournament
	err := row.Scan(&tournament.ID, &tournament.Name, &tournament.Date, &tournament.Format)
	if err != nil {
		return nil, fmt.Errorf("failed to scan tournament: %w", err)
	}

	return &tournament, nil
}

func (tm *TournamentManager) getPlayersForTournament(tournamentID string) ([]string, error) {
	slog.Debug("Getting players for tournament", "tournament_id", tournamentID)

	query := "SELECT ign FROM Player WHERE tournament_id = $1"
	rows, err := database.DB.Query(context.Background(), query, tournamentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query players: %w", err)
	}
	defer rows.Close()

	var players []string
	for rows.Next() {
		var ign string
		if err := rows.Scan(&ign); err != nil {
			return nil, fmt.Errorf("failed to scan player: %w", err)
		}
		players = append(players, ign)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating players: %w", err)
	}

	slog.Debug("Retrieved players from database", "tournament_id", tournamentID, "count", len(players), "players", players)
	return players, nil
}

func getPlayerUUIDByIGN(tournamentID, ign string) (string, error) {
	query := "SELECT id FROM Player WHERE tournament_id = $1 AND ign = $2"
	row := database.DB.QueryRow(context.Background(), query, tournamentID, ign)

	var playerID string
	err := row.Scan(&playerID)
	if err != nil {
		return "", fmt.Errorf("failed to get player UUID for IGN %s: %w", ign, err)
	}

	return playerID, nil
}

func (tm *TournamentManager) saveTournamentResults(tournamentID string, state *formats.SoloSingleElimState) error {
	slog.Debug("Saving tournament results", "tournament_id", tournamentID, "winner", state.Winner)

	// TODO: Implement database operations to save:
	// - Match results
	// - Final winner
	// - Tournament completion status

	return nil
}
