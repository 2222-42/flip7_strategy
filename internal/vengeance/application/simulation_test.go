package application

import (
	"testing"

	"flip7_strategy/internal/vengeance/domain"
	"flip7_strategy/internal/vengeance/domain/strategy"
)

func TestSinglePlayerOneGameFinishes(t *testing.T) {
	sim := NewSimulationService()
	stats := sim.SinglePlayer(1, []StrategySpec{
		{Name: "Heuristic-24", New: func() domain.Strategy { return strategy.NewHeuristicStrategy(24) }},
	})
	if len(stats) != 1 {
		t.Fatalf("stats=%d", len(stats))
	}
	if stats[0].N != 1 {
		t.Fatalf("finished=%d, want 1 (score=%v)", stats[0].N, stats[0])
	}
}

func TestOneVsOneRuns(t *testing.T) {
	sim := NewSimulationService()
	pairs := sim.OneVsOne(1, []StrategySpec{
		{Name: "Cautious", New: func() domain.Strategy { return strategy.NewCautiousStrategy() }},
		{Name: "Aggressive", New: func() domain.Strategy { return strategy.NewAggressiveStrategy() }},
	})
	if len(pairs) != 1 {
		t.Fatalf("pairs=%d", len(pairs))
	}
}
