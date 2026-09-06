package domain

type PointValue struct {
	BaseSum             int   `json:"base_sum"`
	Divided             bool  `json:"divided"`
	AfterDivide         int   `json:"after_divide"`
	SubtractedModifiers []int `json:"subtracted_modifiers"`
	AfterModifiers      int   `json:"after_modifiers"`
	ZeroOverride        bool  `json:"zero_override"`
	Bonus               int   `json:"bonus"`
	Total               int   `json:"total"`
}

type ScoreCalculator struct{}

func NewScoreCalculator() *ScoreCalculator {
	return &ScoreCalculator{}
}

func (sc *ScoreCalculator) Compute(hand *PlayerHand) PointValue {
	if hand == nil || hand.Status == HandStatusBusted {
		return PointValue{Total: 0}
	}

	flip7 := hand.HasFlip7()
	if hand.HasZero() && !flip7 {
		return PointValue{ZeroOverride: true, Total: 0}
	}

	baseSum := 0
	for _, c := range hand.NumberLine {
		if rank, ok := c.Spec.Rank(); ok {
			baseSum += int(rank)
		}
	}

	divided := false
	afterDivide := baseSum
	subtracted := make([]int, 0)
	for _, m := range hand.ModifierLine {
		if m.Spec.ModifierType.IsDivide() {
			divided = true
		} else {
			subtracted = append(subtracted, m.Spec.ModifierType.Amount())
		}
	}
	if divided {
		afterDivide = baseSum / 2
	}

	afterMod := afterDivide
	for _, v := range subtracted {
		afterMod += v
	}
	if afterMod < 0 {
		afterMod = 0
	}

	bonus := 0
	if flip7 {
		bonus = Flip7Bonus
	}

	return PointValue{
		BaseSum:             baseSum,
		Divided:             divided,
		AfterDivide:         afterDivide,
		SubtractedModifiers: subtracted,
		AfterModifiers:      afterMod,
		Bonus:               bonus,
		Total:               afterMod + bonus,
	}
}
