package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"flip7_strategy/internal/application"
	"flip7_strategy/internal/domain"
	"flip7_strategy/internal/domain/strategy"
	"flip7_strategy/internal/infrastructure/console"
)

func main() {
	fmt.Println("Welcome to Flip 7 Strategy!")
	fmt.Println("Select Mode:")
	fmt.Println("1. Automatic Play (Sample Game)")
	fmt.Println("2. Participating (Interactive)")
	fmt.Println("3. Counting (Monte Carlo Simulation)")

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter choice (1-3): ")
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)

	switch input {
	case "1":
		runAutomatic()
	case "2":
		runInteractive()
	case "3":
		runCounting()
	default:
		fmt.Println("Invalid choice. Defaulting to Automatic.")
		runAutomatic()
	}
}

func runAutomatic() {
	fmt.Println("\n--- Automatic Play ---")
	p1 := domain.NewPlayer("Alice (Cautious)", &strategy.CautiousStrategy{})
	p2 := domain.NewPlayer("Bob (Aggressive)", &strategy.AggressiveStrategy{})
	p3 := domain.NewPlayer("Charlie (Probabilistic)", &strategy.ProbabilisticStrategy{})

	players := []*domain.Player{p1, p2, p3}
	game := domain.NewGame(players)
	svc := application.NewGameService(game)
	svc.RunGame()

	printWinner(game)
}

func runInteractive() {
	fmt.Println("\n--- Interactive Play ---")
	p1 := domain.NewPlayer("Alice (Cautious)", &strategy.CautiousStrategy{})
	p2 := domain.NewPlayer("Bob (Aggressive)", &strategy.AggressiveStrategy{})
	p3 := domain.NewPlayer("You (Human)", console.NewHumanStrategy())

	players := []*domain.Player{p3, p1, p2}
	game := domain.NewGame(players)
	svc := application.NewGameService(game)
	svc.RunGame()

	printWinner(game)
}

func runCounting() {
	fmt.Println("\n--- Counting Mode ---")
	sim := application.NewSimulationService()
	sim.RunMonteCarlo(1000) // Run 1000 games
}

func printWinner(game *domain.Game) {
	if len(game.Winners) > 0 {
		fmt.Printf("\nGame Over! Winners:\n")
		for _, winner := range game.Winners {
			fmt.Printf("- %s with %d points!\n", winner.Name, winner.TotalScore)
		}
	} else {
		fmt.Println("\nGame Over! No winner?")
	}
	fmt.Println("Final Scores:")
	for _, p := range game.Players {
		fmt.Printf("- %s: %d\n", p.Name, p.TotalScore)
	}
}
