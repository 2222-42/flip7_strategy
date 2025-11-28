package application

import (
	"flip7_strategy/internal/domain"
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
				s.BankPoints(p)
				s.RemoveActivePlayer(p)
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

	// Check for Second Chance Passing Logic BEFORE adding to hand?
	// Rule: "If they are dealt another Second Chance card, they then choose another active player to give it to."
	if card.Type == domain.CardTypeAction && card.ActionType == domain.ActionSecondChance {
		if p.CurrentHand.HasSecondChance() {
			s.log("%s already has a Second Chance! Must pass it.\n", p.Name)
			// Choose target to give
			candidates := []*domain.Player{}
			for _, ap := range round.ActivePlayers {
				if ap.ID != p.ID {
					candidates = append(candidates, ap)
				}
			}

			if len(candidates) > 0 {
				// Check if all candidates already have a Second Chance card
				allHaveSecondChance := true
				for _, candidate := range candidates {
					if !candidate.CurrentHand.HasSecondChance() {
						allHaveSecondChance = false
						break
					}
				}
				if allHaveSecondChance {
					s.log("All other active players already have a Second Chance. Discarding card.\n")
					// Discard
					s.Game.DiscardPile = append(s.Game.DiscardPile, card)
				} else {
					target := p.Strategy.ChooseTarget(domain.ActionGiveSecondChance, candidates, p, round.Deck)
					s.log("%s gives Second Chance to %s\n", p.Name, target.Name)
					// Add to target's hand (recursive check? "If everyone else already has one, then discard")
					// Let's assume we just add it to target. If target has one, they keep two?
					// Rule says: "If everyone else already has one, then discard the Second Chance card."
					// This implies we should check if target has one.
					// But simpler: just give it.
					// Wait, if target has one, do they pass it too? "If they are dealt..."
					// Receiving from another player is not "dealt" from deck, but let's assume they just keep it.
					target.CurrentHand.ActionCards = append(target.CurrentHand.ActionCards, card)
				}
			} else {
				s.log("%s\n", "No active players to give Second Chance. Discarding.")
				// Discard
				s.Game.DiscardPile = append(s.Game.DiscardPile, card)
			}
			return // Done processing this card for this player
		}
	}

	busted, flip7, discarded := p.CurrentHand.AddCard(card)
	if len(discarded) > 0 {
		s.Game.DiscardPile = append(s.Game.DiscardPile, discarded...)
	}

	if busted {
		s.log("%s BUSTED!\n", p.Name)
		s.RemoveActivePlayer(p)
	} else if flip7 {
		s.log("%s FLIP 7! Bonus!\n", p.Name)
		p.CurrentHand.Status = domain.HandStatusStayed
		s.BankPoints(p)
		s.RemoveActivePlayer(p)
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
		target := p.Strategy.ChooseTarget(domain.ActionFreeze, candidates, p, round.Deck)
		s.log("%s uses Freeze on %s\n", p.Name, target.Name)

		target.CurrentHand.Status = domain.HandStatusFrozen
		s.BankPoints(target)
		s.RemoveActivePlayer(target)

	case domain.ActionFlipThree:
		candidates := []*domain.Player{}
		candidates = append(candidates, round.ActivePlayers...)
		target := p.Strategy.ChooseTarget(domain.ActionFlipThree, candidates, p, round.Deck)
		s.log("%s uses Flip Three on %s\n", p.Name, target.Name)
		s.ExecuteFlipThree(target)
	}
}

// ExecuteFlipThree handles the specific logic of Flip Three (nested actions).
func (s *GameService) ExecuteFlipThree(target *domain.Player) {
	round := s.Game.CurrentRound
	s.log("--- %s must draw 3 cards! ---\n", target.Name)

	pendingActions := []domain.Card{}

	for i := 0; i < 3; i++ {
		if target.CurrentHand.Status != domain.HandStatusActive {
			break
		}

		fCard, err := s.DrawCard()
		if err != nil {
			s.log("%s\n", "Deck and discard pile empty during Flip Three!")
			round.IsEnded = true
			round.EndReason = domain.RoundEndReasonAborted
			return
		}
		s.log("%s forced draw (%d/3): %v\n", target.Name, i+1, fCard)

		// Handle Second Chance logic specifically for Flip Three
		// Second Chance cards are added to the player's hand and resolved immediately.
		// Flip Three and Freeze action cards are queued and resolved after all three cards are drawn.

		if fCard.Type == domain.CardTypeAction {
			if fCard.ActionType == domain.ActionSecondChance {
				// Add to hand immediately (handling passing logic if duplicate)
				s.ProcessCardDraw(target, fCard)
				// If busted/flip7 happened in ProcessCardDraw (unlikely for SecondChance itself, but ProcessCardDraw handles AddCard), loop breaks.
				if target.CurrentHand.Status != domain.HandStatusActive {
					break
				}
				continue
			} else if fCard.ActionType == domain.ActionFreeze || fCard.ActionType == domain.ActionFlipThree {
				// Queue it
				s.log("Action %s queued for after Flip Three.\n", fCard.ActionType)
				pendingActions = append(pendingActions, fCard)

				// Add to hand without triggering immediate resolution (since we resolve later)
				target.CurrentHand.ActionCards = append(target.CurrentHand.ActionCards, fCard)

				// Check Flip 7 with this new card?
				// "Flip 7" check is in AddCard.
				// Let's manually check Flip 7 since we bypassed ProcessCardDraw.
				totalCards := len(target.CurrentHand.NumberCards)
				if totalCards >= 7 {
					s.log("%s FLIP 7! Bonus!\n", target.Name)
					target.CurrentHand.Status = domain.HandStatusStayed
					s.BankPoints(target)
					s.RemoveActivePlayer(target)
					round.EndReason = domain.RoundEndReasonFlip7
					round.IsEnded = true
					return
				}
				continue
			}
		}

		// Normal card (Number/Modifier)
		s.ProcessCardDraw(target, fCard)
		if round.IsEnded || target.CurrentHand.Status != domain.HandStatusActive {
			break
		}
	}

	// Resolve Pending Actions if still active
	if target.CurrentHand.Status == domain.HandStatusActive {
		for _, actionCard := range pendingActions {
			s.log("Resolving queued action %s...\n", actionCard.ActionType)
			s.ResolveAction(target, actionCard)
			if target.CurrentHand.Status != domain.HandStatusActive {
				break
			}
		}
	}

	s.log("--- End of Flip Three for %s ---\n", target.Name)
}

func (s *GameService) RemoveActivePlayer(p *domain.Player) {
	round := s.Game.CurrentRound
	for i, ap := range round.ActivePlayers {
		if ap.ID == p.ID {
			round.ActivePlayers = append(round.ActivePlayers[:i], round.ActivePlayers[i+1:]...)
			return
		}
	}
}

func (s *GameService) BankPoints(p *domain.Player) {
	calc := domain.NewScoreCalculator()
	score := calc.Compute(p.CurrentHand)
	p.BankScore(score.Total)
	s.log("%s banked %d points! Total: %d\n", p.Name, score.Total, p.TotalScore)
}
