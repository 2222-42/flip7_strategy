package domain

// TargetSelector is an interface for selecting targets for action cards.
// In manual mode, this would prompt the user. In AI mode, this uses the strategy.
type TargetSelector interface {
	SelectTarget(actionType ActionType, candidates []*Player, actor *Player) *Player
}

// SecondChanceHandler handles the Second Chance passing logic.
type SecondChanceHandler struct{}

// NewSecondChanceHandler creates a new SecondChanceHandler.
func NewSecondChanceHandler() *SecondChanceHandler {
	return &SecondChanceHandler{}
}

// SecondChanceResult contains the result of processing a Second Chance card.
type SecondChanceResult struct {
	ShouldDiscard bool   // If true, the card should be discarded
	AddToHand     bool   // If true, add to the original player's hand
	PassToPlayer  *Player // If set, pass to this player
}

// HandleSecondChance handles the logic when a player draws a Second Chance card.
// Returns:
// - ShouldDiscard=true if the card should be discarded (all others have one)
// - PassToPlayer set if the card should be passed to another player
// - AddToHand=true if the card should be added to the original player's hand
func (sch *SecondChanceHandler) HandleSecondChance(
	p *Player,
	activePlayers []*Player,
	selector TargetSelector,
) SecondChanceResult {
	// If player doesn't already have a Second Chance, add it to their hand
	if !p.CurrentHand.HasSecondChance() {
		return SecondChanceResult{AddToHand: true}
	}

	// Player already has a Second Chance - must pass it to another active player
	// Filter candidates to only include players who don't already have Second Chance
	candidates := []*Player{}
	for _, ap := range activePlayers {
		if ap.ID != p.ID && !ap.CurrentHand.HasSecondChance() {
			candidates = append(candidates, ap)
		}
	}

	// No valid candidates - discard (either no other players or all have Second Chance)
	if len(candidates) == 0 {
		return SecondChanceResult{ShouldDiscard: true}
	}

	// Select a target to give the card to (candidates are already filtered)
	target := selector.SelectTarget(ActionGiveSecondChance, candidates, p)
	if target == nil {
		// If no target selected, discard the card
		return SecondChanceResult{ShouldDiscard: true}
	}
	return SecondChanceResult{PassToPlayer: target}
}
