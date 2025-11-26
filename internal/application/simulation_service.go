package application

import (
	"flip7_strategy/internal/domain"
	"flip7_strategy/internal/domain/strategy"
	"fmt"
)

const MinDeckSizeBeforeReshuffle = 10

type SimulationService struct{}

func NewSimulationService() *SimulationService {
	return &SimulationService{}
}

func (s *SimulationService) RunMonteCarlo(n int) {
	fmt.Printf("Running %d games (Counting Mode)...\n", n)

	wins := make(map[string]float64)

	// Define strategies to test
	// We need to create fresh players for each game to reset state,
	// but we want to track stats by strategy name.

	for i := 0; i < n; i++ {
		// Create players
		p1 := domain.NewPlayer("Alice (Cautious)", &strategy.CautiousStrategy{})
		p2 := domain.NewPlayer("Bob (Aggressive)", &strategy.AggressiveStrategy{})
		p3 := domain.NewPlayer("Charlie (Probabilistic)", &strategy.ProbabilisticStrategy{})
		p4 := domain.NewPlayer("Dave (Heuristic)", strategy.NewHeuristicStrategy(strategy.DefaultHeuristicThreshold))

		players := []*domain.Player{p1, p2, p3, p4}
		game := domain.NewGame(players)

		svc := NewGameService(game)
		svc.Silent = true // Run silently
		svc.RunGame()

		if len(game.Winners) > 0 {
			points := 1.0 / float64(len(game.Winners))
			for _, winner := range game.Winners {
				wins[winner.Strategy.Name()] += points
			}
		}
	}

	fmt.Println("\n--- Simulation Results ---")
	for name, count := range wins {
		percentage := count / float64(n) * 100
		fmt.Printf("%s: %.2f wins (%.2f%%)\n", name, count, percentage)
	}
}

func (s *SimulationService) RunHeuristicOptimization(gamesPerThreshold int) {
	fmt.Printf("Running Heuristic Optimization (%d games per threshold)...\n", gamesPerThreshold)
	fmt.Println("Threshold | Win Rate")
	fmt.Println("----------|----------")

	type Result struct {
		Threshold int
		WinRate   float64
	}
	var results []Result

	for threshold := 15; threshold <= 35; threshold++ {
		wins := 0.0
		for i := 0; i < gamesPerThreshold; i++ {
			p1 := domain.NewPlayer("Alice", &strategy.CautiousStrategy{})
			p2 := domain.NewPlayer("Bob", &strategy.AggressiveStrategy{})
			p3 := domain.NewPlayer("Charlie", &strategy.ProbabilisticStrategy{})
			p4 := domain.NewPlayer("Dave", strategy.NewHeuristicStrategy(threshold))

			players := []*domain.Player{p1, p2, p3, p4}
			game := domain.NewGame(players)

			svc := NewGameService(game)
			svc.Silent = true
			svc.RunGame()

			for _, winner := range game.Winners {
				if winner.Name == "Dave" {
					wins += 1.0 / float64(len(game.Winners))
				}
			}
		}
		winRate := (wins / float64(gamesPerThreshold)) * 100
		fmt.Printf("%9d | %7.2f%%\n", threshold, winRate)
		results = append(results, Result{Threshold: threshold, WinRate: winRate})
	}

	// Find best
	bestThreshold := 0
	maxWinRate := -1.0
	for _, res := range results {
		if res.WinRate > maxWinRate {
			maxWinRate = res.WinRate
			bestThreshold = res.Threshold
		}
	}
	fmt.Printf("\nBest Threshold: %d (Win Rate: %.2f%%)\n", bestThreshold, maxWinRate)
}
