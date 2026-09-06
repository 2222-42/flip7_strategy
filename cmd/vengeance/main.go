package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"flip7_strategy/internal/vengeance/application"
	"flip7_strategy/internal/vengeance/domain"
	"flip7_strategy/internal/vengeance/domain/strategy"
)

func main() {
	fmt.Println("Flip 7: With a Vengeance")
	fmt.Println("1. Automatic Play (sample AI game)")
	fmt.Println("2. Manual Mode (physical game helper)")
	fmt.Print("Enter choice (1-2): ")

	reader := bufio.NewReader(os.Stdin)
	choice, _ := reader.ReadString('\n')
	switch strings.TrimSpace(choice) {
	case "2":
		application.NewManualGameService(reader).Run()
	default:
		runAutomatic()
	}
}

func runAutomatic() {
	fmt.Println("\n--- Automatic Play ---")
	players := []*domain.Player{
		domain.NewPlayer("Ava (EV)", strategy.NewExpectedValueStrategy()),
		domain.NewPlayer("Ben (Heuristic-26)", strategy.NewHeuristicStrategy(26)),
		domain.NewPlayer("Cara (Adaptive)", strategy.NewAdaptiveStrategy()),
	}
	game := domain.NewGame(players)
	svc := application.NewGameService(game)
	svc.RunGame()

	fmt.Println()
	if len(game.Winners) == 0 {
		fmt.Println("No winner.")
		return
	}
	fmt.Print("Winner(s): ")
	for i, w := range game.Winners {
		if i > 0 {
			fmt.Print(", ")
		}
		fmt.Printf("%s (%d)", w.Name, w.TotalScore)
	}
	fmt.Printf(" after %d rounds\n", game.RoundCount)
}
