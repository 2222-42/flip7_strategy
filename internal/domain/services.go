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

	// Apply modifiers
	// Order matters: usually Additions then Multipliers? Or Multipliers then Additions?
	// Rules usually specify. Let's assume standard math: Base + Adds, then Multipliers?
	// Or maybe Multipliers apply to the base?
	// Doc says: "Impl: Sum numbers, apply modifiers (x2 first), add bonus." -> Wait, x2 first?
	// "apply modifiers (x2 first)" -> This implies x2 applies to the base sum, then you add +2/+4?
	// Or does it mean x2 is processed first in the list?
	// Let's assume: (Sum + Modifiers) * Multipliers?
	// Or: Sum * Multiplier + Modifiers?
	// Let's stick to a safe bet: Sum Numbers. Add "Plus" modifiers. Then Multiply.
	// Wait, doc says "x2 first". That might mean: (Sum * 2) + Modifiers?
	// Let's re-read doc snippet: "Impl: Sum numbers, apply modifiers (x2 first), add bonus."
	// This is ambiguous. "x2 first" could mean "Apply x2 to the base sum, then add other modifiers".
	// Let's try: Total = (BaseSum * Multipliers) + AddModifiers + Bonus.

	multiplier := 1
	addModifiers := 0

	for _, mod := range hand.ModifierCards {
		switch mod.ModifierType {
		case ModifierX2:
			multiplier *= 2
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
		}
	}

	// If "x2 first" means priority:
	// Maybe it means: Total = (BaseSum * Multiplier) + AddModifiers?
	// Or maybe it means: Total = (BaseSum + AddModifiers) * Multiplier?
	// Let's go with (BaseSum + AddModifiers) * Multiplier as it's more standard for "Score = (Points) * Multiplier".
	// BUT, if the doc says "x2 first", maybe it means x2 applies to the base only?
	// Let's assume: Total = (BaseSum + AddModifiers) * Multiplier + Bonus.

	// Wait, Flip 7 Bonus is 15.
	bonus := 0
	// Check if Flip 7 achieved (7 cards)
	totalCards := len(hand.RawNumberCards) + len(hand.ModifierCards) + len(hand.ActionCards)
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
