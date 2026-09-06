package main

import (
	"fmt"

	"flip7_strategy/internal/vengeance/application"
	"flip7_strategy/internal/vengeance/domain"
	"flip7_strategy/internal/vengeance/domain/strategy"
)

func main() {
	fmt.Println("Flip 7: With a Vengeance — automatic sample")
	fmt.Println("Rules: https://cdn.shopify.com/s/files/1/0611/3958/3198/files/26_FLIP_7_VENGEANCE_RULES_C.pdf?v=1770853609")

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
