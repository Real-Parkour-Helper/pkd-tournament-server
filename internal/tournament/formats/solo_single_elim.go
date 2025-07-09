package formats

import (
	"fmt"
	"log/slog"
	"math"
)

type PlayerStatus int

const (
	StatusActive PlayerStatus = iota
	StatusEliminated
	StatusBye
	StatusWinner
)

type Match struct {
	ID       string
	Round    int
	Player1  string
	Player2  string
	Winner   string
	Finished bool
}

type SoloSingleElimState struct {
	TournamentID string
	Players      []string
	PlayerStatus map[string]PlayerStatus
	Matches      []Match
	CurrentRound int
	TotalRounds  int
	IsComplete   bool
	Winner       string
	NextMatchID  int
	RoundWinners [][]string
}

func NewSoloSingleElimState(tournamentID string, players []string) *SoloSingleElimState {
	if len(players) < 2 {
		slog.Warn("Not enough players for single elimination", "count", len(players))
		return nil
	}

	totalRounds := int(math.Ceil(math.Log2(float64(len(players)))))

	playerStatus := make(map[string]PlayerStatus)
	for _, player := range players {
		playerStatus[player] = StatusActive
	}

	state := &SoloSingleElimState{
		TournamentID: tournamentID,
		Players:      players,
		PlayerStatus: playerStatus,
		Matches:      []Match{},
		CurrentRound: 1,
		TotalRounds:  totalRounds,
		IsComplete:   false,
		Winner:       "",
		NextMatchID:  1,
		RoundWinners: make([][]string, totalRounds),
	}

	state.generateRoundMatches()

	return state
}

func (s *SoloSingleElimState) generateRoundMatches() {
	var activePlayers []string

	if s.CurrentRound == 1 {
		activePlayers = make([]string, len(s.Players))
		copy(activePlayers, s.Players)
	} else {
		activePlayers = s.RoundWinners[s.CurrentRound-2]

		for player, status := range s.PlayerStatus {
			if status == StatusBye {
				activePlayers = append(activePlayers, player)
				s.PlayerStatus[player] = StatusActive
			}
		}
	}

	slog.Debug("Generating matches for round", "round", s.CurrentRound, "active_players", len(activePlayers))

	if len(activePlayers)%2 == 1 {
		byePlayer := activePlayers[len(activePlayers)-1]
		s.PlayerStatus[byePlayer] = StatusBye
		activePlayers = activePlayers[:len(activePlayers)-1]
		slog.Debug("Player gets bye", "player", byePlayer, "round", s.CurrentRound)
	}

	for i := 0; i < len(activePlayers); i += 2 {
		match := Match{
			ID:       fmt.Sprintf("match_%d", s.NextMatchID),
			Round:    s.CurrentRound,
			Player1:  activePlayers[i],
			Player2:  activePlayers[i+1],
			Winner:   "",
			Finished: false,
		}
		s.Matches = append(s.Matches, match)
		s.NextMatchID++

		slog.Debug("Created match", "match_id", match.ID, "player1", match.Player1, "player2", match.Player2)
	}
}

func (s *SoloSingleElimState) getActivePlayers() []string {
	if s.CurrentRound == 1 {
		var active []string
		for _, player := range s.Players {
			if s.PlayerStatus[player] == StatusActive {
				active = append(active, player)
			}
		}
		return active
	} else {
		if s.CurrentRound-2 < len(s.RoundWinners) {
			return s.RoundWinners[s.CurrentRound-2]
		}
		return []string{}
	}
}

func (s *SoloSingleElimState) getPlayersWithBye() []string {
	var byePlayers []string
	for player, status := range s.PlayerStatus {
		if status == StatusBye {
			byePlayers = append(byePlayers, player)
		}
	}
	return byePlayers
}

func (s *SoloSingleElimState) HandleGameResult(gameID string, players []string, times []uint64) error {
	if len(players) != len(times) {
		return fmt.Errorf("players and times arrays must have the same length")
	}

	matchIndex := -1
	for i, match := range s.Matches {
		if match.ID == gameID {
			matchIndex = i
			break
		}
	}

	if matchIndex == -1 {
		return fmt.Errorf("match not found: %s", gameID)
	}

	match := &s.Matches[matchIndex]

	if match.Finished {
		return fmt.Errorf("match already finished: %s", gameID)
	}

	winnerIndex := 0
	for i, time := range times {
		if time < times[winnerIndex] {
			winnerIndex = i
		}
	}

	winner := players[winnerIndex]
	slog.Info("setting winner to ", "winner", winner)
	match.Winner = winner
	match.Finished = true

	for _, player := range players {
		if player != winner {
			s.PlayerStatus[player] = StatusEliminated
		}
	}

	slog.Info("Match result processed", "match_id", gameID, "winner", winner)

	if s.isRoundComplete() {
		s.advanceToNextRound()
	}

	return nil
}

func (s *SoloSingleElimState) isRoundComplete() bool {
	for _, match := range s.Matches {
		if match.Round == s.CurrentRound && !match.Finished {
			return false
		}
	}
	return true
}

func (s *SoloSingleElimState) advanceToNextRound() {
	var roundWinners []string
	for _, match := range s.Matches {
		if match.Round == s.CurrentRound && match.Finished {
			roundWinners = append(roundWinners, match.Winner)
		}
	}

	for player, status := range s.PlayerStatus {
		if status == StatusBye {
			roundWinners = append(roundWinners, player)
		}
	}

	s.RoundWinners[s.CurrentRound-1] = roundWinners

	if len(roundWinners) == 1 {
		s.Winner = roundWinners[0]
		s.PlayerStatus[s.Winner] = StatusWinner
		s.IsComplete = true
		slog.Info("Tournament complete", "winner", s.Winner)
		return
	}

	s.CurrentRound++
	if s.CurrentRound > s.TotalRounds {
		slog.Warn("Tournament exceeded expected rounds", "current", s.CurrentRound, "expected", s.TotalRounds)
	}

	slog.Info("Advancing to next round", "round", s.CurrentRound, "remaining_players", len(roundWinners))

	s.generateRoundMatches()
}

func (s *SoloSingleElimState) GetNextMatches() []Match {
	if s.IsComplete {
		return []Match{}
	}

	var nextMatches []Match
	for _, match := range s.Matches {
		if match.Round == s.CurrentRound && !match.Finished {
			nextMatches = append(nextMatches, match)
		}
	}

	return nextMatches
}

func (s *SoloSingleElimState) GetTournamentStatus() map[string]interface{} {
	activePlayers := s.getActivePlayers()
	eliminatedPlayers := []string{}

	for player, status := range s.PlayerStatus {
		if status == StatusEliminated {
			eliminatedPlayers = append(eliminatedPlayers, player)
		}
	}

	return map[string]interface{}{
		"tournament_id":      s.TournamentID,
		"current_round":      s.CurrentRound,
		"total_rounds":       s.TotalRounds,
		"is_complete":        s.IsComplete,
		"winner":             s.Winner,
		"active_players":     activePlayers,
		"eliminated_players": eliminatedPlayers,
		"players_with_bye":   s.getPlayersWithBye(),
		"next_matches":       s.GetNextMatches(),
	}
}

func (s *SoloSingleElimState) GetMatchHistory() []Match {
	var history []Match
	for _, match := range s.Matches {
		if match.Finished {
			history = append(history, match)
		}
	}
	return history
}

func (s *SoloSingleElimState) GetBracketVisualization() string {
	result := fmt.Sprintf("Tournament: %s\n", s.TournamentID)
	result += fmt.Sprintf("Current Round: %d/%d\n", s.CurrentRound, s.TotalRounds)
	result += fmt.Sprintf("Status: %s\n", func() string {
		if s.IsComplete {
			return fmt.Sprintf("COMPLETE - Winner: %s", s.Winner)
		}
		return "IN PROGRESS"
	}())
	result += "\n"

	for round := 1; round <= s.CurrentRound; round++ {
		result += fmt.Sprintf("=== Round %d ===\n", round)

		roundMatches := []Match{}
		for _, match := range s.Matches {
			if match.Round == round {
				roundMatches = append(roundMatches, match)
			}
		}

		for _, match := range roundMatches {
			status := "PENDING"
			if match.Finished {
				status = fmt.Sprintf("WINNER: %s", match.Winner)
			}
			result += fmt.Sprintf("  %s vs %s - %s\n", match.Player1, match.Player2, status)
		}

		byePlayers := []string{}
		for player, status := range s.PlayerStatus {
			if status == StatusBye {
				byePlayers = append(byePlayers, player)
			}
		}
		for _, player := range byePlayers {
			result += fmt.Sprintf("  %s - BYE\n", player)
		}

		result += "\n"
	}

	return result
}
