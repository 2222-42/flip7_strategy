package application

import (
	"flip7_strategy/internal/domain"
	"flip7_strategy/internal/domain/rules"
	"fmt"
)

// GameService orchestrates the game.
type GameService struct {
	Game   *domain.Game
	Silent bool
}

func NewGameService(game *domain.Game) *GameService {
	return &GameService{Game: game}
}

func (s *GameService) log(format string, a ...interface{}) {
	if !s.Silent {
		fmt.Printf(format, a...)
	}
}

// RunGame loops until a winner is found.
func (s *GameService) RunGame() {
	deck := domain.NewDeck()

	for !s.Game.IsCompleted {
		s.Game.RoundCount++
		s.Game.CurrentRound = domain.NewRound(s.Game.Players, s.Game.Players[s.Game.DealerIndex], deck)
		s.PlayRound()

		if s.Game.CurrentRound.EndReason == domain.RoundEndReasonAborted {
			s.log("Game aborted due to empty deck/discard.\n")
			s.Game.IsCompleted = true
			break
		}

		// Move all cards from players' hands to the discard pile.
		// The deck persists across rounds and is passed to the next dealer.

		// Collect cards from players' hands
		for _, p := range s.Game.Players {
			// Number cards
			for _, val := range p.CurrentHand.RawNumberCards {
				s.Game.DiscardPile = append(s.Game.DiscardPile, domain.Card{Type: domain.CardTypeNumber, Value: val})
			}
			// Modifier cards
			s.Game.DiscardPile = append(s.Game.DiscardPile, p.CurrentHand.ModifierCards...)
			// Action cards
			s.Game.DiscardPile = append(s.Game.DiscardPile, p.CurrentHand.ActionCards...)
		}

		// Check for winner
		winners := s.Game.DetermineWinners()
		if len(winners) > 0 {
			s.Game.IsCompleted = true
			s.Game.Winners = winners
			break
		}

		// Rotate dealer
		s.Game.DealerIndex = (s.Game.DealerIndex + 1) % len(s.Game.Players)

		// Update deck reference for the next round
		// If a reshuffle happened during PlayRound, s.Game.CurrentRound.Deck points to the new deck.
		// We must update our local 'deck' variable so the next round uses the valid deck.
		deck = s.Game.CurrentRound.Deck
	}
}

// DrawCard handles drawing a card, reshuffling from discard pile if necessary.
func (s *GameService) DrawCard() (domain.Card, error) {
	round := s.Game.CurrentRound
	card, err := round.Deck.Draw()
	if err == nil {
		return card, nil
	}

	// Deck is empty, try to reshuffle
	if len(s.Game.DiscardPile) == 0 {
		return domain.Card{}, fmt.Errorf("deck is empty and discard pile is empty")
	}

	s.log("Deck empty. Reshuffling %d cards from discard pile...\n", len(s.Game.DiscardPile))
	round.Deck = domain.NewDeckFromCards(s.Game.DiscardPile)
	s.Game.DiscardPile = []domain.Card{} // Clear discard pile

	// Try drawing again
	return round.Deck.Draw()
}

// GetCard implements rules.CardSource
func (s *GameService) GetCard() (domain.Card, error) {
	return s.DrawCard()
}

// SelectTarget implements rules.TargetSelector
func (s *GameService) SelectTarget(actionType domain.ActionType, candidates []*domain.Player, source *domain.Player) *domain.Player {
	if ds, ok := source.Strategy.(interface{ SetDeck(*domain.Deck) }); ok {
		ds.SetDeck(s.Game.CurrentRound.Deck)
	}
	target := source.Strategy.ChooseTarget(actionType, candidates, source)
	// Log the choice
	s.log("%s targets %s with %s\n", source.Name, target.Name, actionType)
	return target
}

// PlayRound handles a single round.
func (s *GameService) PlayRound() {
	round := s.Game.CurrentRound
	s.log("--- New Round! Dealer: %s ---\n", round.Dealer.Name)

	// Initial Deal: Sequential starting from Dealer (which is ActivePlayers[0])
	// "If draw an action card(i.e., flip_three, freeze) then he choose immediately even if other one doesn't draw yet."

	// Iterate through a copy of active players because ResolveAction might remove players (Freeze)
	// But for initial deal, we iterate through the list as established by NewRound.
	// Note: NewRound already ordered ActivePlayers starting with Dealer.

	initialActive := make([]*domain.Player, len(round.ActivePlayers))
	copy(initialActive, round.ActivePlayers)

	for _, p := range initialActive {
		// Check if player is still active (might have been removed by someone else's action? Unlikely in initial deal unless FlipThree targets them?)
		// Actually, FlipThree could target another player and bust them.
		if p.CurrentHand.Status != domain.HandStatusActive {
			continue
		}

		card, err := s.DrawCard()
		if err != nil {
			s.log("%s\n", "Deck and discard pile empty during initial deal!")
			round.IsEnded = true
			round.EndReason = domain.RoundEndReasonAborted
			return
		}
		s.log("%s dealt: %v\n", p.Name, card)

		s.ProcessCardDraw(p, card)
		if round.IsEnded {
			return
		}
	}

	// Turns
	for len(round.ActivePlayers) > 0 {
		active := make([]*domain.Player, len(round.ActivePlayers))
		copy(active, round.ActivePlayers)

		for _, p := range active {
			if p.CurrentHand.Status != domain.HandStatusActive {
				continue
			}

			// Strategy Decision
			choice := p.Strategy.Decide(round.Deck, p.CurrentHand, p.TotalScore, round.Players)
			s.log("%s decides to %s\n", p.Name, choice)

			if choice == domain.TurnChoiceStay {
				p.CurrentHand.Status = domain.HandStatusStayed
				score := p.BankCurrentHand()
				s.log("%s banked %d points! Total: %d\n", p.Name, score, p.TotalScore)
				s.Game.CurrentRound.RemoveActivePlayer(p)
			} else {
				// Hit
				card, err := s.DrawCard()
				if err != nil {
					s.log("%s\n", "Deck and discard pile empty!")
					round.IsEnded = true
					round.EndReason = domain.RoundEndReasonAborted
					return
				}
				s.log("%s drew: %v\n", p.Name, card)

				s.ProcessCardDraw(p, card)
				if round.IsEnded {
					return
				}
			}
		}
	}
}

// ProcessCardDraw handles adding a card and resolving its effects.
func (s *GameService) ProcessCardDraw(p *domain.Player, card domain.Card) {
	engine := rules.NewGameEngine()
	result, err := engine.ApplyCard(s.Game.CurrentRound, p, card, s)
	if err != nil {
		s.log("Error applying card: %v\n", err)
		return
	}

	// Log Discards
	if len(result.Discarded) > 0 {
		for _, d := range result.Discarded {
			s.log("Discarded: %v\n", d)
			s.Game.DiscardPile = append(s.Game.DiscardPile, d)
		}
	}

	// Log Outcomes
	if result.Busted {
		s.log("%s BUSTED!\n", p.Name)
	}
	if result.Flip7 {
		s.log("%s FLIP 7! Bonus!\n", p.Name)
		s.log("%s banked %d points! Total: %d\n", p.Name, result.BankedScore, p.TotalScore)
	}
	if result.Stayed && !result.Flip7 { // Explicit stay handled in loop, but Flip7 causes stay
		// Flip7 logged above
	}

	// Handle Actions
	if result.ActionType != "" {
		switch result.ActionType {
		case domain.ActionGiveSecondChance:
			s.log("%s gives Second Chance to %s\n", p.Name, result.Target.Name)
		case domain.ActionFreeze:
			s.log("%s uses Freeze on %s\n", p.Name, result.Target.Name)
			s.log("%s banked %d points! Total: %d\n", result.Target.Name, result.Target.TotalScore, result.Target.TotalScore) // Score already updated in engine
		case domain.ActionFlipThree:
			s.log("%s uses Flip Three on %s\n", p.Name, result.Target.Name)
			s.ExecuteFlipThree(result.Target)
		}
	}
}

// ExecuteFlipThree handles the specific logic of Flip Three (nested actions).
func (s *GameService) ExecuteFlipThree(target *domain.Player) {
	s.log("--- %s must draw 3 cards! ---\n", target.Name)
	engine := rules.NewGameEngine()
	results, err := engine.ExecuteFlipThree(s.Game.CurrentRound, target, s, s)
	if err != nil {
		s.log("Error executing Flip Three: %v\n", err)
		return
	}

	for i, res := range results {
		s.log("%s forced draw (%d/3) result: Busted=%v, Action=%s\n", target.Name, i+1, res.Busted, res.ActionType)
		if len(res.Discarded) > 0 {
			s.Game.DiscardPile = append(s.Game.DiscardPile, res.Discarded...)
		}
		if res.Busted {
			s.log("%s BUSTED in Flip Three!\n", target.Name)
		}
		if res.Flip7 {
			s.log("%s FLIP 7 in Flip Three!\n", target.Name)
			s.log("%s banked %d points! Total: %d\n", target.Name, res.BankedScore, target.TotalScore)
		}
		if res.ActionType == domain.ActionFreeze && res.Target != nil {
			s.log("Freeze resolved on %s\n", res.Target.Name)
			s.log("%s banked %d points! Total: %d\n", res.Target.Name, res.Target.TotalScore, res.Target.TotalScore)
		}
		if res.ActionType == domain.ActionFlipThree && res.Target != nil {
			s.log("Nested Flip Three on %s\n", res.Target.Name)
			s.ExecuteFlipThree(res.Target) // Recursive
		}
	}

	s.log("--- End of Flip Three for %s ---\n", target.Name)
}
