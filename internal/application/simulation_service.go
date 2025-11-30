package application

import (
	"flip7_strategy/internal/domain"
	"flip7_strategy/internal/domain/strategy"
	"fmt"
	"sort"
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
		p2 := domain.NewPlayer("Bob (Aggressive)", strategy.NewAggressiveStrategy())
		p3 := domain.NewPlayer("Charlie (Probabilistic)", &strategy.ProbabilisticStrategy{})
		p4 := domain.NewPlayer("Dave (Heuristic)", strategy.NewHeuristicStrategy(strategy.DefaultHeuristicThreshold))
		p5 := domain.NewPlayer("Eve (ExpectedValue)", &strategy.ExpectedValueStrategy{})
		p6 := domain.NewPlayer("Frank (Adaptive)", strategy.NewAdaptiveStrategy())

		players := []*domain.Player{p1, p2, p3, p4, p5, p6}
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
			p2 := domain.NewPlayer("Bob", strategy.NewAggressiveStrategy())
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
		{"Aggressive", strategy.NewAggressiveStrategy()},
		{"Probabilistic", &strategy.ProbabilisticStrategy{}},
		{"Heuristic-27", strategy.NewHeuristicStrategy(27)},
		{"ExpectedValue", &strategy.ExpectedValueStrategy{}},
		{"Adaptive", strategy.NewAdaptiveStrategy()},
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

		sort.Ints(rounds)
		var median float64
		if len(rounds)%2 == 0 {
			median = float64(rounds[len(rounds)/2-1]+rounds[len(rounds)/2]) / 2.0
		} else {
			median = float64(rounds[len(rounds)/2])
		}

		fmt.Printf("%-15s | %10.2f | %13.2f\n", strat.Name, avg, median)
	}
}

func (s *SimulationService) RunMultiplayerEvaluation(n int) {
	fmt.Printf("Running Multiplayer Evaluation (%d games per player count)...\n", n)

	// Strategies pool
	strats := []domain.Strategy{
		&strategy.CautiousStrategy{},
		strategy.NewAggressiveStrategyWithSelector(strategy.NewRiskBasedTargetSelector(0.90)),
		strategy.NewProbabilisticStrategyWithSelector(strategy.NewRiskBasedTargetSelector(0.50)),
		strategy.NewHeuristicStrategyWithSelector(27, strategy.NewRiskBasedTargetSelector(0.65)),
		strategy.NewExpectedValueStrategyWithSelector(strategy.NewRiskBasedTargetSelector(0.70)),
		strategy.NewOptimizedAdaptiveStrategy(),
	}

	for playerCount := 1; playerCount <= 5; playerCount++ {
		fmt.Printf("\n--- %d Players ---\n", playerCount)
		wins := make(map[string]float64)

		for i := 0; i < n; i++ {
			var players []*domain.Player
			for j := 0; j < playerCount; j++ {
				// Assign strategies in round-robin with rotation based on game index
				// This ensures all strategies get played even if playerCount < len(strats)
				stratIndex := (i + j) % len(strats)
				strat := strats[stratIndex]
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
		{"Aggressive", strategy.NewAggressiveStrategyWithSelector(strategy.NewRiskBasedTargetSelector(0.90))},
		{"Probabilistic", strategy.NewProbabilisticStrategyWithSelector(strategy.NewRiskBasedTargetSelector(0.50))},
		{"Heuristic-27", strategy.NewHeuristicStrategyWithSelector(27, strategy.NewRiskBasedTargetSelector(0.65))},
		{"ExpectedValue", strategy.NewExpectedValueStrategyWithSelector(strategy.NewRiskBasedTargetSelector(0.70))},
		{"Adaptive", strategy.NewOptimizedAdaptiveStrategy()},
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

func (s *SimulationService) RunTargetSelectionSimulation(n int) {
	fmt.Printf("Running Target Selection Simulation (%d games)...\n", n)

	// Define strategies with different target selectors
	type StrategyConfig struct {
		Name  string
		Strat domain.Strategy
	}
	// Define thresholds
	thresholds := []float64{0.5, 0.65, 0.7, 0.8, 0.9}

	// Helper to run a batch
	runBatch := func(batchName string, targetStrategies []StrategyConfig) {
		fmt.Printf("\n--- Batch: %s ---\n", batchName)
		wins := make(map[string]float64)

		for i := 0; i < n; i++ {
			var players []*domain.Player
			// Add target strategies
			for _, s := range targetStrategies {
				players = append(players, domain.NewPlayer(s.Name, s.Strat))
			}
			// Add some standard opponents to fill the table and provide a baseline
			// 5 target strategies + 3 standard = 8 players
			players = append(players, domain.NewPlayer("Standard-Cautious", &strategy.CautiousStrategy{}))
			players = append(players, domain.NewPlayer("Standard-Aggressive", strategy.NewAggressiveStrategy()))
			players = append(players, domain.NewPlayer("Standard-Probabilistic", &strategy.ProbabilisticStrategy{}))

			game := domain.NewGame(players)
			svc := NewGameService(game)
			svc.Silent = true
			svc.RunGame()

			if len(game.Winners) > 0 {
				points := 1.0 / float64(len(game.Winners))
				for _, winner := range game.Winners {
					wins[winner.Name] += points
				}
			}
		}

		// Sort and print results for this batch
		type Result struct {
			Name string
			Wins float64
		}
		var results []Result
		for name, count := range wins {
			results = append(results, Result{Name: name, Wins: count})
		}
		sort.Slice(results, func(i, j int) bool {
			return results[i].Wins > results[j].Wins
		})

		for _, res := range results {
			percentage := res.Wins / float64(n) * 100
			fmt.Printf("%-25s: %.2f wins (%.2f%%)\n", res.Name, res.Wins, percentage)
		}
	}

	// 1. Expected Value Batch
	var evStrategies []StrategyConfig
	for _, t := range thresholds {
		name := fmt.Sprintf("EV-Risk-%.2f", t)
		evStrategies = append(evStrategies, StrategyConfig{Name: name, Strat: strategy.NewExpectedValueStrategyWithSelector(strategy.NewRiskBasedTargetSelector(t))})
	}
	runBatch("Expected Value", evStrategies)

	// 2. Probabilistic Batch
	var probStrategies []StrategyConfig
	for _, t := range thresholds {
		name := fmt.Sprintf("Prob-Risk-%.2f", t)
		probStrategies = append(probStrategies, StrategyConfig{Name: name, Strat: strategy.NewProbabilisticStrategyWithSelector(strategy.NewRiskBasedTargetSelector(t))})
	}
	runBatch("Probabilistic", probStrategies)

	// 3. Heuristic Batch
	var heurStrategies []StrategyConfig
	for _, t := range thresholds {
		name := fmt.Sprintf("Heur-Risk-%.2f", t)
		heurStrategies = append(heurStrategies, StrategyConfig{Name: name, Strat: strategy.NewHeuristicStrategyWithSelector(27, strategy.NewRiskBasedTargetSelector(t))})
	}
	runBatch("Heuristic", heurStrategies)

	// 4. Aggressive Batch
	var aggrStrategies []StrategyConfig
	for _, t := range thresholds {
		name := fmt.Sprintf("Aggr-Risk-%.2f", t)
		aggrStrategies = append(aggrStrategies, StrategyConfig{Name: name, Strat: strategy.NewAggressiveStrategyWithSelector(strategy.NewRiskBasedTargetSelector(t))})
	}
	runBatch("Aggressive", aggrStrategies)
}

func (s *SimulationService) RunAdaptiveOptimizationSimulation(n int) {
	fmt.Printf("Running Adaptive Strategy Optimization (%d games)...\n", n)

	strategies := []struct {
		Name  string
		Strat domain.Strategy
	}{
		{"Adaptive-Standard", strategy.NewAdaptiveStrategy()},
		{"Adaptive-Optimized", strategy.NewOptimizedAdaptiveStrategy()},
		{"ExpectedValue-Opt", strategy.NewExpectedValueStrategyWithSelector(strategy.NewRiskBasedTargetSelector(0.70))},
		{"Aggressive-Opt", strategy.NewAggressiveStrategyWithSelector(strategy.NewRiskBasedTargetSelector(0.90))},
	}

	wins := make(map[string]float64)

	for i := 0; i < n; i++ {
		var players []*domain.Player
		for _, s := range strategies {
			players = append(players, domain.NewPlayer(s.Name, s.Strat))
		}
		// Add some standard opponents to fill the table
		players = append(players, domain.NewPlayer("Standard-Cautious", &strategy.CautiousStrategy{}))
		players = append(players, domain.NewPlayer("Standard-Probabilistic", &strategy.ProbabilisticStrategy{}))

		game := domain.NewGame(players)
		svc := NewGameService(game)
		svc.Silent = true
		svc.RunGame()

		if len(game.Winners) > 0 {
			points := 1.0 / float64(len(game.Winners))
			for _, winner := range game.Winners {
				wins[winner.Name] += points
			}
		}
	}

	fmt.Println("\n--- Adaptive Optimization Results ---")
	// Sort by wins
	type Result struct {
		Name string
		Wins float64
	}
	var results []Result
	for name, count := range wins {
		results = append(results, Result{Name: name, Wins: count})
	}
	sort.Slice(results, func(i, j int) bool {
		return results[i].Wins > results[j].Wins
	})

	for _, res := range results {
		percentage := res.Wins / float64(n) * 100
		fmt.Printf("%-20s: %.2f wins (%.2f%%)\n", res.Name, res.Wins, percentage)
	}
}
