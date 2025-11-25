package domain_test

import (
	"fmt"
	"testing"

	"flip7_strategy/internal/domain"
)

func TestDetermineWinners(t *testing.T) {
	tests := []struct {
		name           string
		scores         []int
		expectedWinner []string // Names of expected winners
	}{
		{
			name:           "No one reached 200",
			scores:         []int{100, 150, 199},
			expectedWinner: []string{},
		},
		{
			name:           "One player reached 200",
			scores:         []int{200, 150, 100},
			expectedWinner: []string{"P1"},
		},
		{
			name:           "Multiple players reached 200, distinct winner",
			scores:         []int{210, 205, 100},
			expectedWinner: []string{"P1"},
		},
		{
			name:           "Multiple players reached 200, tie",
			scores:         []int{210, 210, 100},
			expectedWinner: []string{"P1", "P2"},
		},
		{
			name:           "Multiple players reached 200, tie but lower score ignored",
			scores:         []int{220, 220, 210},
			expectedWinner: []string{"P1", "P2"},
		},
		{
			name:           "Three-way tie at exactly 200",
			scores:         []int{200, 200, 200},
			expectedWinner: []string{"P1", "P2", "P3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			players := make([]*domain.Player, len(tt.scores))
			for i, score := range tt.scores {
				name := fmt.Sprintf("P%d", i+1)
				p := domain.NewPlayer(name, nil)
				p.TotalScore = score
				players[i] = p
			}

			game := domain.NewGame(players)
			winners := game.DetermineWinners()

			if len(winners) != len(tt.expectedWinner) {
				t.Errorf("Expected %d winners, got %d", len(tt.expectedWinner), len(winners))
			}

			// Check names
			winnerNames := make(map[string]bool)
			for _, w := range winners {
				winnerNames[w.Name] = true
			}

			for _, name := range tt.expectedWinner {
				if !winnerNames[name] {
					t.Errorf("Expected winner %s not found", name)
				}
			}
		})
	}
}
