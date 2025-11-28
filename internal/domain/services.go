package domain

// ScoreCalculator calculates the score for a hand.
type ScoreCalculator struct{}

func NewScoreCalculator() *ScoreCalculator {
	return &ScoreCalculator{}
}

func (sc *ScoreCalculator) Compute(hand *PlayerHand) PointValue {
	if hand.Status == HandStatusBusted {
		return PointValue{Total: 0}
	}

	baseSum := 0
	for _, val := range hand.RawNumberCards {
		baseSum += int(val)
	}

	multiplier := 1
	addModifiers := 0

	for _, mod := range hand.ModifierCards {
		switch mod.ModifierType {
		case ModifierPlus2:
			addModifiers += 2
		case ModifierPlus4:
			addModifiers += 4
		case ModifierPlus6:
			addModifiers += 6
		case ModifierPlus8:
			addModifiers += 8
		case ModifierPlus10:
			addModifiers += 10
		case ModifierX2:
			multiplier *= 2
		}
	}

	// Calculate total score
	// Formula: (BaseSum + AddModifiers) * Multiplier + Bonus

	// Flip 7 bonus: awards 15 points if the player has 7 or more unique number cards.
	bonus := 0
	// Check if Flip 7 achieved (7 unique number cards)
	totalCards := len(hand.NumberCards)
	if totalCards >= 7 {
		bonus = 15
	}

	total := (baseSum+addModifiers)*multiplier + bonus

	return PointValue{
		BaseSum:   baseSum,
		Modifiers: []int{addModifiers, multiplier}, // Simplified representation
		Bonus:     bonus,
		Total:     total,
	}
}
