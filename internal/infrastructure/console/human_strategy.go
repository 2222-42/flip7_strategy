package console

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"flip7_strategy/internal/domain"
)

// HumanStrategy allows a human to play via CLI.
type HumanStrategy struct {
	reader *bufio.Reader
}

func NewHumanStrategy() *HumanStrategy {
	return &HumanStrategy{
		reader: bufio.NewReader(os.Stdin),
	}
}

func (s *HumanStrategy) Name() string {
	return "Human"
}

func (s *HumanStrategy) Decide(deck *domain.Deck, hand *domain.PlayerHand, playerScore int, otherPlayers []*domain.Player) domain.TurnChoice {
	fmt.Printf("\n--- Your Turn ---\n")
	fmt.Printf("Your Hand: %v (Modifiers: %v, Actions: %v)\n", hand.RawNumberCards, hand.ModifierCards, hand.ActionCards)

	calc := domain.NewScoreCalculator()
	score := calc.Compute(hand)
	fmt.Printf("Current Hand Score: %d (Total Banked: %d)\n", score.Total, playerScore)

	risk := deck.EstimateHitRisk(hand.NumberCards)
	fmt.Printf("Estimated Risk of Bust: %.2f%%\n", risk*100)

	for {
		fmt.Print("Choose action (hit/stay): ")
		input, _ := s.reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))

		if input == "hit" || input == "h" {
			return domain.TurnChoiceHit
		}
		if input == "stay" || input == "s" {
			return domain.TurnChoiceStay
		}
		fmt.Println("Invalid input. Please enter 'hit' or 'stay'.")
	}
}

func (s *HumanStrategy) ChooseTarget(action domain.ActionType, candidates []*domain.Player, self *domain.Player) *domain.Player {
	fmt.Printf("\n--- Choose Target for %s ---\n", action)
	for i, p := range candidates {
		label := p.Name
		if p.ID == self.ID {
			label += " (You)"
		}
		fmt.Printf("%d: %s (Score: %d)\n", i+1, label, p.TotalScore)
	}

	for {
		fmt.Printf("Enter number (1-%d): ", len(candidates))
		input, _ := s.reader.ReadString('\n')
		input = strings.TrimSpace(input)

		idx, err := strconv.Atoi(input)
		if err == nil && idx >= 1 && idx <= len(candidates) {
			return candidates[idx-1]
		}
		fmt.Println("Invalid selection.")
	}
}
