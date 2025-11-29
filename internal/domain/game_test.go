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

func TestRoundRobinDealerRotation(t *testing.T) {
	p1 := domain.NewPlayer("P1", nil)
	p2 := domain.NewPlayer("P2", nil)
	p3 := domain.NewPlayer("P3", nil)
	players := []*domain.Player{p1, p2, p3}
	game := domain.NewGame(players)

	if game.DealerIndex != 0 {
		t.Errorf("Expected initial DealerIndex 0, got %d", game.DealerIndex)
	}
	if game.DealerIndex >= len(game.Players) {
		t.Fatalf("DealerIndex %d is out of bounds for %d players", game.DealerIndex, len(game.Players))
	}

	game.CurrentRound = domain.NewRound(game.Players, game.Players[game.DealerIndex], domain.NewDeck())
	if game.CurrentRound.Dealer.ID != p1.ID {
		t.Errorf("Round 1: Expected Dealer P1, got %s", game.CurrentRound.Dealer.Name)
	}
	if len(game.CurrentRound.ActivePlayers) != 3 {
		t.Errorf("Round 1: Expected 3 active players, got %d", len(game.CurrentRound.ActivePlayers))
	}
	if game.CurrentRound.ActivePlayers[0].ID != p1.ID {
		t.Errorf("Round 1: Expected Start Player P1, got %s", game.CurrentRound.ActivePlayers[0].Name)
	}

	game.DealerIndex = (game.DealerIndex + 1) % len(game.Players)

	game.CurrentRound = domain.NewRound(game.Players, game.Players[game.DealerIndex], domain.NewDeck())
	if game.CurrentRound.Dealer.ID != p2.ID {
		t.Errorf("Round 2: Expected Dealer P2, got %s", game.CurrentRound.Dealer.Name)
	}
	if len(game.CurrentRound.ActivePlayers) != 3 {
		t.Errorf("Round 2: Expected 3 active players, got %d", len(game.CurrentRound.ActivePlayers))
	}
	if game.CurrentRound.ActivePlayers[0].ID != p2.ID {
		t.Errorf("Round 2: Expected Start Player P2, got %s", game.CurrentRound.ActivePlayers[0].Name)
	}
	if game.CurrentRound.ActivePlayers[1].ID != p3.ID {
		t.Errorf("Round 2: Expected Second Player P3, got %s", game.CurrentRound.ActivePlayers[1].Name)
	}
	if game.CurrentRound.ActivePlayers[2].ID != p1.ID {
		t.Errorf("Round 2: Expected Third Player P1, got %s", game.CurrentRound.ActivePlayers[2].Name)
	}

	game.DealerIndex = (game.DealerIndex + 1) % len(game.Players)

	game.CurrentRound = domain.NewRound(game.Players, game.Players[game.DealerIndex], domain.NewDeck())
	if game.CurrentRound.Dealer.ID != p3.ID {
		t.Errorf("Round 3: Expected Dealer P3, got %s", game.CurrentRound.Dealer.Name)
	}
	if len(game.CurrentRound.ActivePlayers) != 3 {
		t.Errorf("Round 3: Expected 3 active players, got %d", len(game.CurrentRound.ActivePlayers))
	}
	if game.CurrentRound.ActivePlayers[0].ID != p3.ID {
		t.Errorf("Round 3: Expected Start Player P3, got %s", game.CurrentRound.ActivePlayers[0].Name)
	}

	game.DealerIndex = (game.DealerIndex + 1) % len(game.Players)
	if game.DealerIndex != 0 {
		t.Errorf("Expected DealerIndex to wrap to 0, got %d", game.DealerIndex)
	}
}
