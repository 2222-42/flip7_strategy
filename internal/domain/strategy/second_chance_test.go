package strategy_test

import (
	"testing"

	"flip7_strategy/internal/domain"
	"flip7_strategy/internal/domain/strategy"
)

func TestStrategies_HitWithSecondChance(t *testing.T) {
	strategies := []domain.Strategy{
		&strategy.CautiousStrategy{},
		strategy.NewAggressiveStrategy(),
		&strategy.ProbabilisticStrategy{},
		strategy.NewHeuristicStrategy(27),
		strategy.NewExpectedValueStrategy(),
		strategy.NewAdaptiveStrategy(),
	}

	for _, s := range strategies {
		t.Run(s.Name(), func(t *testing.T) {
			hand := domain.NewPlayerHand()
			// Add a Second Chance card
			hand.ActionCards = append(hand.ActionCards, domain.Card{Type: domain.CardTypeAction, ActionType: domain.ActionSecondChance})

			// Add some number cards to make it risky (normally they might stay)
			hand.NumberCards[domain.NumberValue(10)] = struct{}{}
			hand.RawNumberCards = append(hand.RawNumberCards, domain.NumberValue(10))
			hand.NumberCards[domain.NumberValue(11)] = struct{}{}
			hand.RawNumberCards = append(hand.RawNumberCards, domain.NumberValue(11))
			hand.NumberCards[domain.NumberValue(12)] = struct{}{}
			hand.RawNumberCards = append(hand.RawNumberCards, domain.NumberValue(12))

			deck := domain.NewDeck()
			choice := s.Decide(deck, hand, 0, nil)

			if choice != domain.TurnChoiceHit {
				t.Errorf("Strategy %s should HIT with Second Chance card, but chose %s", s.Name(), choice)
			}
		})
	}
}
