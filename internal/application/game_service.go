package application

import (
	"flip7_strategy/internal/domain"
	"fmt"
)

// GameService orchestrates the game.
type GameService struct {
	Game                *domain.Game
	Silent              bool
	secondChanceHandler *domain.SecondChanceHandler
}

// gameServiceFlipThreeCardSource implements FlipThreeCardSource for AI mode.
type gameServiceFlipThreeCardSource struct {
	service *GameService
}

func (gs *gameServiceFlipThreeCardSource) GetNextCard(cardNum int, target *domain.Player) (domain.Card, error) {
	return gs.service.DrawCard()
}

// gameServiceFlipThreeCardProcessor implements FlipThreeCardProcessor for AI mode.
type gameServiceFlipThreeCardProcessor struct {
	service *GameService
}

func (gp *gameServiceFlipThreeCardProcessor) ProcessImmediateCard(target *domain.Player, card domain.Card) error {
	gp.service.ProcessCardDraw(target, card)
	return nil
}

func (gp *gameServiceFlipThreeCardProcessor) ProcessQueuedAction(target *domain.Player, card domain.Card) error {
	gp.service.ResolveAction(target, card)
	return nil
}

// strategyTargetSelector wraps a Strategy to implement TargetSelector interface.
type strategyTargetSelector struct {
	strategy domain.Strategy
	deck     *domain.Deck
}

func (sts *strategyTargetSelector) SelectTarget(actionType domain.ActionType, candidates []*domain.Player, actor *domain.Player) *domain.Player {
	target := sts.strategy.ChooseTarget(actionType, candidates, actor)

	// Validate that the target is in the candidates list
	if target != nil {
		for _, candidate := range candidates {
			if candidate.ID == target.ID {
				return target
			}
		}
		// Target not in candidates - return first candidate as fallback
		if len(candidates) > 0 {
			return candidates[0]
		}
	}
	return nil
}

func NewGameService(game *domain.Game) *GameService {
	return &GameService{
		Game:                game,
		secondChanceHandler: domain.NewSecondChanceHandler(),
	}
}

func (s *GameService) log(format string, a ...interface{}) {
	if !s.Silent {
		fmt.Printf(format, a...)
	}
}

// RunGame loops until a winner is found.
func (s *GameService) RunGame() {
	if s.Game.Deck == nil {
		s.Game.Deck = domain.NewDeck()
	}

	for !s.Game.IsCompleted {
		s.Game.RoundCount++
		s.Game.CurrentRound = domain.NewRound(s.Game.Players, s.Game.Players[s.Game.DealerIndex], s.Game.Deck)
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
		s.Game.Deck = s.Game.CurrentRound.Deck
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
	round := s.Game.CurrentRound

	// Check for Second Chance Passing Logic BEFORE adding to hand
	// Rule: "If they are dealt another Second Chance card, they then choose another active player to give it to."
	if card.Type == domain.CardTypeAction && card.ActionType == domain.ActionSecondChance {
		// Create a selector for the strategy
		selector := &strategyTargetSelector{strategy: p.Strategy, deck: round.Deck}

		// Set deck for strategies that need it
		if ds, ok := p.Strategy.(interface{ SetDeck(*domain.Deck) }); ok {
			ds.SetDeck(round.Deck)
		}

		result := s.secondChanceHandler.HandleSecondChance(p, round.ActivePlayers, selector)

		if result.ShouldDiscard {
			s.log("All other active players already have a Second Chance. Discarding card.\n")
			s.Game.DiscardPile = append(s.Game.DiscardPile, card)
			return
		} else if result.PassToPlayer != nil {
			s.log("%s gives Second Chance to %s\n", p.Name, result.PassToPlayer.Name)
			result.PassToPlayer.CurrentHand.ActionCards = append(result.PassToPlayer.CurrentHand.ActionCards, card)
			return
		}
		// Otherwise, fall through to add to player's hand
	}

	busted, flip7, discarded := p.CurrentHand.AddCard(card)
	if len(discarded) > 0 {
		s.Game.DiscardPile = append(s.Game.DiscardPile, discarded...)
	}

	if busted {
		s.log("%s BUSTED!\n", p.Name)
		s.Game.CurrentRound.RemoveActivePlayer(p)
	} else if flip7 {
		s.log("%s FLIP 7! Bonus!\n", p.Name)
		p.CurrentHand.Status = domain.HandStatusStayed
		score := p.BankCurrentHand()
		s.log("%s banked %d points! Total: %d\n", p.Name, score, p.TotalScore)
		s.Game.CurrentRound.RemoveActivePlayer(p)
		round.EndReason = domain.RoundEndReasonFlip7
		round.IsEnded = true
	} else {
		// Resolve Immediate Actions
		if card.Type == domain.CardTypeAction {
			s.ResolveAction(p, card)
		}
	}
}

// ResolveAction handles the effect of an action card.
func (s *GameService) ResolveAction(p *domain.Player, card domain.Card) {
	round := s.Game.CurrentRound

	switch card.ActionType {
	case domain.ActionFreeze:
		candidates := []*domain.Player{}
		candidates = append(candidates, round.ActivePlayers...)
		if ds, ok := p.Strategy.(interface{ SetDeck(*domain.Deck) }); ok {
			ds.SetDeck(round.Deck)
		}
		target := p.Strategy.ChooseTarget(domain.ActionFreeze, candidates, p)
		s.log("%s uses Freeze on %s\n", p.Name, target.Name)

		target.CurrentHand.Status = domain.HandStatusFrozen
		score := target.BankCurrentHand()
		s.log("%s banked %d points! Total: %d\n", target.Name, score, target.TotalScore)
		s.Game.CurrentRound.RemoveActivePlayer(target)

	case domain.ActionFlipThree:
		candidates := []*domain.Player{}
		candidates = append(candidates, round.ActivePlayers...)
		if ds, ok := p.Strategy.(interface{ SetDeck(*domain.Deck) }); ok {
			ds.SetDeck(round.Deck)
		}
		target := p.Strategy.ChooseTarget(domain.ActionFlipThree, candidates, p)
		s.log("%s uses Flip Three on %s\n", p.Name, target.Name)
		s.ExecuteFlipThree(target)
	}
}

// ExecuteFlipThree handles the specific logic of Flip Three (nested actions).
func (s *GameService) ExecuteFlipThree(target *domain.Player) {
	// Create FlipThree executor with AI mode implementations
	source := &gameServiceFlipThreeCardSource{service: s}
	processor := &gameServiceFlipThreeCardProcessor{service: s}

	// Create logger function that uses the service's log method
	var logger domain.FlipThreeLogger
	if !s.Silent {
		logger = func(message string) {
			s.log("%s\n", message)
		}
	}

	executor := domain.NewFlipThreeExecutor(source, processor, logger)
	executor.Execute(target, s.Game.CurrentRound)
}
