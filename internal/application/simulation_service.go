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

func (s *SimulationService) RunSinglePlayerOptimization(n int) {
	fmt.Printf("Running Single Player Optimization (%d games per strategy)...\n", n)
	fmt.Println("Strategy | Avg Rounds | Median Rounds")
	fmt.Println("---------|------------|--------------")

	strategies := []struct {
		Name  string
		Strat domain.Strategy
	}{
		{"Cautious", &strategy.CautiousStrategy{}},
		{"Aggressive", &strategy.AggressiveStrategy{}},
		{"Probabilistic", &strategy.ProbabilisticStrategy{}},
		{"Heuristic-27", strategy.NewHeuristicStrategy(27)},
	}

	for _, strat := range strategies {
		var rounds []int
		for i := 0; i < n; i++ {
			p := domain.NewPlayer("Player", strat.Strat)
			players := []*domain.Player{p}
			game := domain.NewGame(players)
			svc := NewGameService(game)
			svc.Silent = true
			svc.RunGame()

			// Check if player reached 200 points
			if p.TotalScore >= 200 {
				rounds = append(rounds, game.RoundCount)
			}
		}

		if len(rounds) == 0 {
			fmt.Printf("%-15s | N/A | N/A\n", strat.Name)
			continue
		}

		sum := 0
		for _, r := range rounds {
			sum += r
		}
		avg := float64(sum) / float64(len(rounds))

		// Calculate median
		// Simple bubble sort for median (n is small enough or we can implement sort)
		// Or just use a quick sort implementation or import sort.
		// Let's import sort.
		// Wait, imports are at the top. I can't easily add imports with replace_file_content unless I replace the whole file or use multi_replace.
		// I'll use a simple selection sort here since n is likely 1000-ish, O(n^2) is fine for simulation or just implement a quick helper.
		// Actually, let's just use a simple loop to find median or just average for now if sort is hard.
		// But the requirement says "median".
		// I will use a simple bubble sort for now, it's easy to write inline.
		for i := 0; i < len(rounds)-1; i++ {
			for j := 0; j < len(rounds)-i-1; j++ {
				if rounds[j] > rounds[j+1] {
					rounds[j], rounds[j+1] = rounds[j+1], rounds[j]
				}
			}
		}
		median := float64(rounds[len(rounds)/2])
		if len(rounds)%2 == 0 {
			median = float64(rounds[len(rounds)/2-1]+rounds[len(rounds)/2]) / 2.0
		}

		fmt.Printf("%-15s | %10.2f | %13.2f\n", strat.Name, avg, median)
	}
}

func (s *SimulationService) RunMultiplayerEvaluation(n int) {
	fmt.Printf("Running Multiplayer Evaluation (%d games per player count)...\n", n)

	// Strategies pool
	strats := []domain.Strategy{
		&strategy.CautiousStrategy{},
		&strategy.AggressiveStrategy{},
		&strategy.ProbabilisticStrategy{},
		strategy.NewHeuristicStrategy(27),
	}

	for playerCount := 1; playerCount <= 5; playerCount++ {
		fmt.Printf("\n--- %d Players ---\n", playerCount)
		wins := make(map[string]float64)

		for i := 0; i < n; i++ {
			var players []*domain.Player
			for j := 0; j < playerCount; j++ {
				// Assign strategies in round-robin or random?
				// Let's do round-robin from the pool
				strat := strats[j%len(strats)]
				name := fmt.Sprintf("P%d-%s", j+1, strat.Name())
				players = append(players, domain.NewPlayer(name, strat))
			}

			game := domain.NewGame(players)
			svc := NewGameService(game)
			svc.Silent = true
			svc.RunGame()

			if len(game.Winners) > 0 {
				points := 1.0 / float64(len(game.Winners))
				for _, winner := range game.Winners {
					// We want to aggregate by strategy name, not player name
					// Player name includes strategy name, so we can parse or just use strategy name directly if we had access.
					// But we constructed the player with the strategy.
					// Let's just use the strategy name from the winner's strategy instance.
					wins[winner.Strategy.Name()] += points
				}
			}
		}

		for name, count := range wins {
			percentage := count / float64(n) * 100
			fmt.Printf("%s: %.2f wins (%.2f%%)\n", name, count, percentage)
		}
	}
}

func (s *SimulationService) RunStrategyCombinationEvaluation(n int) {
	fmt.Printf("Running Strategy Combination Evaluation (%d games per pair)...\n", n)

	strategies := []struct {
		Name  string
		Strat domain.Strategy
	}{
		{"Cautious", &strategy.CautiousStrategy{}},
		{"Aggressive", &strategy.AggressiveStrategy{}},
		{"Probabilistic", &strategy.ProbabilisticStrategy{}},
		{"Heuristic-27", strategy.NewHeuristicStrategy(27)},
	}

	for i := 0; i < len(strategies); i++ {
		for j := i + 1; j < len(strategies); j++ {
			s1 := strategies[i]
			s2 := strategies[j]

			fmt.Printf("\n--- %s vs %s ---\n", s1.Name, s2.Name)
			wins := make(map[string]float64)

			for k := 0; k < n; k++ {
				// Create fresh players for each game
				p1 := domain.NewPlayer(s1.Name, s1.Strat)
				p2 := domain.NewPlayer(s2.Name, s2.Strat)
				players := []*domain.Player{p1, p2}

				game := domain.NewGame(players)
				svc := NewGameService(game)
				svc.Silent = true
				svc.RunGame()

				if len(game.Winners) > 0 {
					points := 1.0 / float64(len(game.Winners))
					for _, winner := range game.Winners {
						// Use the strategy name from our struct to ensure consistency
						// winner.Strategy.Name() might differ slightly (e.g. Heuristic-27) but should be close.
						// Let's map back to our names based on player name or just use winner.Name since we set it to strat name.
						wins[winner.Name] += points
					}
				}
			}

			// Print results for this pair
			count1 := wins[s1.Name]
			pct1 := count1 / float64(n) * 100
			count2 := wins[s2.Name]
			pct2 := count2 / float64(n) * 100

			fmt.Printf("%s: %.2f wins (%.2f%%)\n", s1.Name, count1, pct1)
			fmt.Printf("%s: %.2f wins (%.2f%%)\n", s2.Name, count2, pct2)
		}
	}
}
