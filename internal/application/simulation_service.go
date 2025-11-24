package application

import (
	"flip7_strategy/internal/domain"
	"flip7_strategy/internal/domain/strategy"
	"fmt"
)

type SimulationService struct{}

func NewSimulationService() *SimulationService {
	return &SimulationService{}
}

func (s *SimulationService) RunMonteCarlo(n int) {
	fmt.Printf("Running %d games (Counting Mode)...\n", n)

	wins := make(map[string]int)

	// Define strategies to test
	// We need to create fresh players for each game to reset state,
	// but we want to track stats by strategy name.

	for i := 0; i < n; i++ {
		// Create players
		p1 := domain.NewPlayer("Alice (Cautious)", &strategy.CautiousStrategy{})
		p2 := domain.NewPlayer("Bob (Aggressive)", &strategy.AggressiveStrategy{})
		p3 := domain.NewPlayer("Charlie (Probabilistic)", &strategy.ProbabilisticStrategy{})

		players := []*domain.Player{p1, p2, p3}
		game := domain.NewGame(players)

		svc := NewGameService(game)
		svc.Silent = true // Run silently
		svc.RunGame()

		if game.Winner != nil {
			wins[game.Winner.Strategy.Name()]++
		}
	}

	fmt.Println("\n--- Simulation Results ---")
	for name, count := range wins {
		percentage := float64(count) / float64(n) * 100
		fmt.Printf("%s: %d wins (%.2f%%)\n", name, count, percentage)
	}
}
