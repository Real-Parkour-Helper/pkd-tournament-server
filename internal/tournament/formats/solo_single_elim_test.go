package formats_test

import (
	"fmt"
	"log/slog"
	"testing"
	"tournament-manager/internal/tournament/formats"
)

func TestSingleElim(t *testing.T) {
	s := formats.NewSoloSingleElimState("id", []string{
		"senez",
		"kha0x",
		"i77_",
		"tauktes",
	})

	slog.Debug("current status: ", "status", s.GetTournamentStatus())
	m := s.GetNextMatches()
	if m[0].Player1 != "senez" || m[0].Player2 != "kha0x" {
		t.Errorf("unexpected matchup, expected %v vs %v, got %v vs %v", "senez", "kha0x", m[0].Player1, m[0].Player2)
	}

	if m[1].Player1 != "i77_" || m[1].Player2 != "tauktes" {
		t.Errorf("unexpected matchup, expected %v vs %v, got %v vs %v", "i77_", "tauktes", m[1].Player1, m[1].Player2)
	}

	fmt.Println(s.GetBracketVisualization())

	s.HandleGameResult("match_1", []string{"senez", "kha0x"}, []uint64{135000, 120000})
	if s.Matches[0].Winner != "kha0x" {
		t.Errorf("unexpected winner, expected %v, got %v", "kha0x", m[0].Winner)
	}

	fmt.Println(s.GetBracketVisualization())

	s.HandleGameResult("match_2", []string{"i77_", "tauktes"}, []uint64{120000, 135000})
	if s.Matches[1].Winner != "i77_" {
		t.Errorf("unexpected winner, expected %v, got %v", "tauktes", m[1].Winner)
	}

	fmt.Println(s.GetBracketVisualization())

	m = s.GetNextMatches()
	if len(m) != 1 {
		t.Errorf("unexpected number of matches, expected %v, got %v", 1, len(m))
	}

	if m[0].Player1 != "kha0x" || m[0].Player2 != "i77_" {
		t.Errorf("unexpected matchup, expected %v vs %v, got %v vs %v", "kha0x", "i77_", m[0].Player1, m[0].Player2)
	}

	s.HandleGameResult("match_3", []string{"kha0x", "i77_"}, []uint64{135000, 120000})
	if s.Matches[2].Winner != "i77_" {
		t.Errorf("unexpected winner, expected %v, got %v", "i77_", m[0].Winner)
	}

	fmt.Println(s.GetBracketVisualization())

	slog.Info("history: ", "history", s.GetMatchHistory())
}
