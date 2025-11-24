package application

import (
	"flip7_strategy/internal/domain"
	"fmt"
)

// GameService orchestrates the game.
type GameService struct {
	Game *domain.Game
}

func NewGameService(game *domain.Game) *GameService {
	return &GameService{Game: game}
}

// RunGame loops until a winner is found.
func (s *GameService) RunGame() {
	deck := domain.NewDeck()

	for !s.Game.IsCompleted {
		s.Game.CurrentRound = domain.NewRound(s.Game.Players, s.Game.Players[s.Game.DealerIndex], deck)
		s.PlayRound()

		// Check for winner
		for _, p := range s.Game.Players {
			if p.TotalScore >= 200 {
				s.Game.IsCompleted = true
				s.Game.Winner = p
				break // First to 200 wins (or handle tie-break?)
			}
		}

		// Rotate dealer
		s.Game.DealerIndex = (s.Game.DealerIndex + 1) % len(s.Game.Players)

		// Reshuffle if deck is low?
		if len(deck.Cards) < 10 {
			fmt.Println("Reshuffling deck...")
			deck = domain.NewDeck()
		}
	}
}

// PlayRound handles a single round.
func (s *GameService) PlayRound() {
	round := s.Game.CurrentRound
	fmt.Printf("--- New Round! Dealer: %s ---\n", round.Dealer.Name)

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

		card, err := round.Deck.Draw()
		if err != nil {
			fmt.Println("Deck empty during initial deal!")
			return
		}
		fmt.Printf("%s dealt: %v\n", p.Name, card)

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
			fmt.Printf("%s decides to %s\n", p.Name, choice)

			if choice == domain.TurnChoiceStay {
				p.CurrentHand.Status = domain.HandStatusStayed
				s.BankPoints(p)
				s.RemoveActivePlayer(p)
			} else {
				// Hit
				card, err := round.Deck.Draw()
				if err != nil {
					fmt.Println("Deck empty!")
					round.IsEnded = true
					return
				}
				fmt.Printf("%s drew: %v\n", p.Name, card)

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
			fmt.Printf("%s already has a Second Chance! Must pass it.\n", p.Name)
			// Choose target to give
			candidates := []*domain.Player{}
			for _, ap := range round.ActivePlayers {
				if ap.ID != p.ID {
					candidates = append(candidates, ap)
				}
			}

			if len(candidates) > 0 {
				target := p.Strategy.ChooseTarget(domain.ActionGiveSecondChance, candidates, p)
				fmt.Printf("%s gives Second Chance to %s\n", p.Name, target.Name)
				// Add to target's hand (recursive check? "If everyone else already has one, then discard")
				// Let's assume we just add it to target. If target has one, they keep two?
				// Rule says: "If everyone else already has one, then discard the Second Chance card."
				// This implies we should check if target has one.
				// But simpler: just give it.
				// Wait, if target has one, do they pass it too? "If they are dealt..."
				// Receiving from another player is not "dealt" from deck, but let's assume they just keep it.
				target.CurrentHand.ActionCards = append(target.CurrentHand.ActionCards, card)
			} else {
				fmt.Println("No active players to give Second Chance. Discarding.")
				// Discard
			}
			return // Done processing this card for this player
		}
	}

	busted, flip7 := p.CurrentHand.AddCard(card)
	if busted {
		fmt.Printf("%s BUSTED!\n", p.Name)
		s.RemoveActivePlayer(p)
	} else if flip7 {
		fmt.Printf("%s FLIP 7! Bonus!\n", p.Name)
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
		for _, ap := range round.ActivePlayers {
			candidates = append(candidates, ap)
		}
		target := p.Strategy.ChooseTarget(domain.ActionFreeze, candidates, p)
		fmt.Printf("%s uses Freeze on %s\n", p.Name, target.Name)

		target.CurrentHand.Status = domain.HandStatusFrozen
		s.BankPoints(target)
		s.RemoveActivePlayer(target)

	case domain.ActionFlipThree:
		candidates := []*domain.Player{}
		for _, ap := range round.ActivePlayers {
			candidates = append(candidates, ap)
		}
		target := p.Strategy.ChooseTarget(domain.ActionFlipThree, candidates, p)
		fmt.Printf("%s uses Flip Three on %s\n", p.Name, target.Name)
		s.ExecuteFlipThree(target)
	}
}

// ExecuteFlipThree handles the specific logic of Flip Three (nested actions).
func (s *GameService) ExecuteFlipThree(target *domain.Player) {
	round := s.Game.CurrentRound
	fmt.Printf("--- %s must draw 3 cards! ---\n", target.Name)

	pendingActions := []domain.Card{}

	for i := 0; i < 3; i++ {
		if target.CurrentHand.Status != domain.HandStatusActive {
			break
		}

		fCard, err := round.Deck.Draw()
		if err != nil {
			fmt.Println("Deck empty during Flip Three!")
			round.IsEnded = true
			return
		}
		fmt.Printf("%s forced draw (%d/3): %v\n", target.Name, i+1, fCard)

		// Handle Second Chance logic specifically for Flip Three
		// "(3-a) if, a Second Chance card is revealed, it may be set aside and used"
		// This implies it's added to hand immediately?
		// "(3-b) If another Flip Three or Freeze card is revealed they are resolved AFTER all three cards are drawn."

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
				fmt.Printf("Action %s queued for after Flip Three.\n", fCard.ActionType)
				pendingActions = append(pendingActions, fCard)

				// Add to hand? "revealed... resolved AFTER".
				// Usually action cards are added to hand.
				// But if we add it now, ProcessCardDraw would trigger resolution.
				// We should AddCard ONLY (no resolution).
				// But AddCard checks for Bust/Flip7.
				// Action cards don't cause bust (unless they are also Number cards? No).
				// So we just add it.
				target.CurrentHand.ActionCards = append(target.CurrentHand.ActionCards, fCard)

				// Check Flip 7 with this new card?
				// "Flip 7" check is in AddCard.
				// Let's manually check Flip 7 since we bypassed ProcessCardDraw.
				totalCards := len(target.CurrentHand.RawNumberCards) + len(target.CurrentHand.ModifierCards) + len(target.CurrentHand.ActionCards)
				if totalCards >= 7 {
					fmt.Printf("%s FLIP 7! Bonus!\n", target.Name)
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
			fmt.Printf("Resolving queued action %s...\n", actionCard.ActionType)
			s.ResolveAction(target, actionCard)
			if target.CurrentHand.Status != domain.HandStatusActive {
				break
			}
		}
	}

	fmt.Printf("--- End of Flip Three for %s ---\n", target.Name)
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
	fmt.Printf("%s banked %d points! Total: %d\n", p.Name, score.Total, p.TotalScore)
}
