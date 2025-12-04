package domain_test

import (
	"testing"

	"flip7_strategy/internal/domain"
)

// mockTargetSelector is a mock implementation of TargetSelector for testing.
type mockTargetSelector struct {
	targetToReturn *domain.Player
}

func (m *mockTargetSelector) SelectTarget(actionType domain.ActionType, candidates []*domain.Player, actor *domain.Player) *domain.Player {
	return m.targetToReturn
}

func TestSecondChanceHandler_HandleSecondChance(t *testing.T) {
	tests := []struct {
		name                    string
		playerHasSecondChance   bool
		otherPlayersHaveSecond  []bool
		expectedShouldDiscard   bool
		expectedAddToHand       bool
		expectedPassToPlayer    bool
	}{
		{
			name:                  "Player doesn't have Second Chance - add to hand",
			playerHasSecondChance: false,
			expectedAddToHand:     true,
		},
		{
			name:                  "Player has Second Chance, no other active players - discard",
			playerHasSecondChance: true,
			otherPlayersHaveSecond: []bool{},
			expectedShouldDiscard: true,
		},
		{
			name:                   "Player has Second Chance, all others have it - discard",
			playerHasSecondChance:  true,
			otherPlayersHaveSecond: []bool{true, true},
			expectedShouldDiscard:  true,
		},
		{
			name:                   "Player has Second Chance, one other doesn't - pass to player",
			playerHasSecondChance:  true,
			otherPlayersHaveSecond: []bool{false},
			expectedPassToPlayer:   true,
		},
		{
			name:                   "Player has Second Chance, some others don't - pass to player",
			playerHasSecondChance:  true,
			otherPlayersHaveSecond: []bool{true, false, true},
			expectedPassToPlayer:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := domain.NewSecondChanceHandler()

			// Create test player
			player := domain.NewPlayer("TestPlayer", nil)
			player.StartNewRound()

			// Add Second Chance to player if needed
			if tt.playerHasSecondChance {
				player.CurrentHand.ActionCards = append(player.CurrentHand.ActionCards,
					domain.Card{Type: domain.CardTypeAction, ActionType: domain.ActionSecondChance})
			}

			// Create other active players
			activePlayers := []*domain.Player{player}
			var targetPlayer *domain.Player
			for i, hasSecond := range tt.otherPlayersHaveSecond {
				otherPlayer := domain.NewPlayer("OtherPlayer"+string(rune('A'+i)), nil)
				otherPlayer.StartNewRound()
				if hasSecond {
					otherPlayer.CurrentHand.ActionCards = append(otherPlayer.CurrentHand.ActionCards,
						domain.Card{Type: domain.CardTypeAction, ActionType: domain.ActionSecondChance})
				}
				activePlayers = append(activePlayers, otherPlayer)
				if !hasSecond && targetPlayer == nil {
					targetPlayer = otherPlayer // First player without Second Chance
				}
			}

			// Create mock selector
			selector := &mockTargetSelector{targetToReturn: targetPlayer}

			// Execute
			result := handler.HandleSecondChance(player, activePlayers, selector)

			// Verify
			if result.ShouldDiscard != tt.expectedShouldDiscard {
				t.Errorf("ShouldDiscard = %v, want %v", result.ShouldDiscard, tt.expectedShouldDiscard)
			}
			if result.AddToHand != tt.expectedAddToHand {
				t.Errorf("AddToHand = %v, want %v", result.AddToHand, tt.expectedAddToHand)
			}
			if tt.expectedPassToPlayer && result.PassToPlayer == nil {
				t.Error("Expected PassToPlayer to be set, but it was nil")
			}
			if !tt.expectedPassToPlayer && result.PassToPlayer != nil {
				t.Errorf("Expected PassToPlayer to be nil, but it was %v", result.PassToPlayer.Name)
			}
		})
	}
}
