package main

import (
	"fmt"
	"math/rand"
	"time"

	"flip7_strategy/internal/application"
	"flip7_strategy/internal/domain"
	"flip7_strategy/internal/domain/strategy"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	fmt.Println("Starting Flip 7 Simulation...")

	// Create Players with different strategies
	p1 := domain.NewPlayer("Alice (Cautious)", &strategy.CautiousStrategy{})
	p2 := domain.NewPlayer("Bob (Aggressive)", &strategy.AggressiveStrategy{})
	p3 := domain.NewPlayer("Charlie (Probabilistic)", &strategy.ProbabilisticStrategy{})

	players := []*domain.Player{p1, p2, p3}

	// Initialize Game
	game := domain.NewGame(players)

	// Create Service
	svc := application.NewGameService(game)

	// Run Game
	svc.RunGame()

	fmt.Printf("\nGame Over! Winner: %s with %d points!\n", game.Winner.Name, game.Winner.TotalScore)
	fmt.Println("Final Scores:")
	for _, p := range players {
		fmt.Printf("- %s: %d\n", p.Name, p.TotalScore)
	}
}
