package rules

import (
	"flip7_strategy/internal/domain"
)

// CardSource abstracts where cards come from (deck, user input, etc.).
type CardSource interface {
	GetCard() (domain.Card, error)
}

// TargetSelector abstracts how a target is chosen for an action.
type TargetSelector interface {
	SelectTarget(actionType domain.ActionType, candidates []*domain.Player, source *domain.Player) *domain.Player
}

// GameEngine encapsulates the core rules of Flip 7.
type GameEngine struct{}

func NewGameEngine() *GameEngine {
	return &GameEngine{}
}

// ApplyResult contains the outcome of applying a card to a player.
type ApplyResult struct {
	Busted      bool
	Flip7       bool
	Stayed      bool
	BankedScore int
	ActionType  domain.ActionType // If an action was resolved immediately
	Target      *domain.Player    // Target of the action (if any)
	Discarded   []domain.Card     // Cards discarded (e.g. Second Chance passed but rejected)
}

// ApplyCard adds a card to the player's hand and resolves its immediate effects.
// It handles:
// - Adding card to hand
// - Checking for Bust
// - Checking for Flip 7
// - Handling duplicate Second Chance (passing to another player)
// - Resolving immediate Actions (Freeze, Flip Three) via TargetSelector
func (e *GameEngine) ApplyCard(round *domain.Round, player *domain.Player, card domain.Card, selector TargetSelector) (*ApplyResult, error) {
	result := &ApplyResult{}

	// Special handling for Second Chance Passing Logic
	// Rule: "If they are dealt another Second Chance card, they then choose another active player to give it to."
	if card.Type == domain.CardTypeAction && card.ActionType == domain.ActionSecondChance {
		if player.CurrentHand.HasSecondChance() {
			// Must pass it
			candidates := []*domain.Player{}
			for _, ap := range round.ActivePlayers {
				if ap.ID != player.ID {
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
					// Discard
					result.Discarded = append(result.Discarded, card)
				} else {
					target := selector.SelectTarget(domain.ActionGiveSecondChance, candidates, player)
					if target != nil {
						// Give to target
						target.CurrentHand.ActionCards = append(target.CurrentHand.ActionCards, card)
						result.ActionType = domain.ActionGiveSecondChance
						result.Target = target
					} else {
						// Should not happen if candidates > 0, but fallback to discard
						result.Discarded = append(result.Discarded, card)
					}
				}
			} else {
				// No candidates, discard
				result.Discarded = append(result.Discarded, card)
			}
			return result, nil
		}
	}

	// Add card to hand
	busted, flip7, discarded := player.CurrentHand.AddCard(card)
	if len(discarded) > 0 {
		result.Discarded = append(result.Discarded, discarded...)
	}

	if busted {
		result.Busted = true
		player.CurrentHand.Status = domain.HandStatusBusted
		round.RemoveActivePlayer(player)
	} else if flip7 {
		result.Flip7 = true
		player.CurrentHand.Status = domain.HandStatusStayed
		result.BankedScore = player.BankCurrentHand()
		round.RemoveActivePlayer(player)
		round.End(domain.RoundEndReasonFlip7)
	} else {
		// Resolve Immediate Actions
		if card.Type == domain.CardTypeAction {
			result.ActionType = card.ActionType
			e.resolveAction(round, player, card, selector, result)
		}
	}

	return result, nil
}

func (e *GameEngine) resolveAction(round *domain.Round, player *domain.Player, card domain.Card, selector TargetSelector, result *ApplyResult) {
	switch card.ActionType {
	case domain.ActionFreeze:
		candidates := []*domain.Player{}
		candidates = append(candidates, round.ActivePlayers...)
		target := selector.SelectTarget(domain.ActionFreeze, candidates, player)
		if target != nil {
			result.Target = target
			target.CurrentHand.Status = domain.HandStatusFrozen
			score := target.BankCurrentHand()
			// We don't store banked score for target in result.BankedScore (that's for the acting player usually),
			// but we could. For now, let's assume the caller handles logging this specific event details if needed.
			// Actually, let's add a field or just rely on the caller checking the target's state.
			_ = score
			round.RemoveActivePlayer(target)
		}

	case domain.ActionFlipThree:
		candidates := []*domain.Player{}
		candidates = append(candidates, round.ActivePlayers...)
		target := selector.SelectTarget(domain.ActionFlipThree, candidates, player)
		if target != nil {
			result.Target = target
			// The actual execution of Flip Three (drawing 3 cards) is complex and usually handled separately
			// because it involves multiple draws.
			// Here we just identify the target. The caller (Service) should call ExecuteFlipThree next.
		}
	}
}

// ExecuteFlipThree handles the logic of a player being forced to draw 3 cards.
// It returns a list of results for each card drawn/processed.
func (e *GameEngine) ExecuteFlipThree(round *domain.Round, target *domain.Player, source CardSource, selector TargetSelector) ([]*ApplyResult, error) {
	var results []*ApplyResult
	var pendingActions []domain.Card

	for i := 0; i < 3; i++ {
		if target.CurrentHand.Status != domain.HandStatusActive {
			break
		}

		card, err := source.GetCard()
		if err != nil {
			return results, err
		}

		// Handle cards drawn during Flip Three
		if card.Type == domain.CardTypeAction {
			if card.ActionType == domain.ActionSecondChance {
				// Process immediately
				res, err := e.ApplyCard(round, target, card, selector)
				if err != nil {
					return results, err
				}
				results = append(results, res)
				if target.CurrentHand.Status != domain.HandStatusActive {
					break
				}
			} else if card.ActionType == domain.ActionFreeze || card.ActionType == domain.ActionFlipThree {
				// Queue it
				pendingActions = append(pendingActions, card)
				// Add to hand without triggering immediate resolution
				// We manually add it to ActionCards to avoid ApplyCard triggering resolution
				target.CurrentHand.ActionCards = append(target.CurrentHand.ActionCards, card)

				// Check Flip 7 manually since we bypassed ApplyCard
				// (Logic duplicated from AddCard, but without side effects)
				// Actually, AddCard checks for 7 number cards. Action cards don't count towards number cards count directly for Flip 7?
				// Wait, Flip 7 is "Collect one of each number card 0-12".
				// Action cards are NOT number cards. So adding an action card never triggers Flip 7.
				// So we are safe.

				// Create a dummy result for logging
				results = append(results, &ApplyResult{
					ActionType: card.ActionType,
					// Indicate it was queued?
				})
			}
		} else {
			// Normal card
			res, err := e.ApplyCard(round, target, card, selector)
			if err != nil {
				return results, err
			}
			results = append(results, res)
			if target.CurrentHand.Status != domain.HandStatusActive {
				break
			}
		}
	}

	// Resolve Pending Actions
	if target.CurrentHand.Status == domain.HandStatusActive {
		for _, card := range pendingActions {
			// Now we resolve them. We can use resolveAction helper.
			res := &ApplyResult{ActionType: card.ActionType}
			e.resolveAction(round, target, card, selector, res)
			results = append(results, res)
			if target.CurrentHand.Status != domain.HandStatusActive {
				break
			}
		}
	}

	return results, nil
}
