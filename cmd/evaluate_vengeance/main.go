package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"flip7_strategy/internal/vengeance/application"
	"flip7_strategy/internal/vengeance/domain"
)

func main() {
	n := flag.Int("n", 200, "games per cell")
	heuristicGames := flag.Int("heuristic-n", 100, "games per heuristic threshold")
	brutal := flag.Bool("brutal", false, "evaluate Brutal Mode (does not overwrite standard tables)")
	flag.Parse()

	sim := application.NewSimulationService()
	if *brutal {
		sim.Rules = domain.BrutalRules()
	}
	started := time.Now()

	fmt.Fprintf(os.Stderr, "Heuristic sweep (%d games/threshold)...\n", *heuristicGames)
	sweep := sim.HeuristicSweep(*heuristicGames, 16, 32, 2)
	bestTh := 24
	bestAvg := 1e9
	for _, st := range sweep {
		var th int
		_, _ = fmt.Sscanf(st.Name, "Heuristic-%d", &th)
		if st.N > 0 && st.Avg < bestAvg {
			bestAvg = st.Avg
			bestTh = th
		}
	}

	catalog := application.CatalogWithHeuristic(bestTh)

	fmt.Fprintf(os.Stderr, "Single player (%d)...\n", *n)
	single := sim.SinglePlayer(*n, catalog)

	fmt.Fprintf(os.Stderr, "Multiplayer 2-5 (%d)...\n", *n)
	multi := sim.Multiplayer(*n, 2, 5, catalog)

	fmt.Fprintf(os.Stderr, "1v1 (%d)...\n", *n)
	pairs := sim.OneVsOne(*n, catalog)

	fmt.Fprintf(os.Stderr, "Flip Four thresholds (%d)...\n", *n)
	ff := sim.FlipFourThresholds(*n, []float64{0.30, 0.50, 0.70})

	var takeWin, attackWin float64
	if *brutal {
		fmt.Fprintf(os.Stderr, "Flip 7 take vs attack (%d)...\n", *n)
		takeWin, attackWin = sim.Flip7TakeVsAttack(*n)
	}

	elapsed := time.Since(started)
	names := make([]string, 0, len(catalog))
	for _, spec := range catalog {
		names = append(names, spec.Name)
	}
	printMarkdown(*n, *heuristicGames, bestTh, names, sweep, single, multi, pairs, ff, elapsed, *brutal, takeWin, attackWin)
}

func printMarkdown(n, heurN, bestTh int, names []string, sweep, single []application.NamedStat, multi []application.WinTable, pairs []application.PairResult, ff []application.TargetResult, elapsed time.Duration, brutal bool, takeWin, attackWin float64) {
	if brutal {
		fmt.Println("# Flip 7: With a Vengeance — Brutal Mode Strategy Evaluation")
	} else {
		fmt.Println("# Flip 7: With a Vengeance Strategy Evaluation")
	}
	fmt.Println()
	fmt.Printf("**Date**: %s\n", time.Now().UTC().Format("2006-01-02"))
	fmt.Printf("**Games per cell**: %d (heuristic sweep: %d per threshold)\n", n, heurN)
	fmt.Printf("**Runtime**: %s\n", elapsed.Round(time.Second))
	fmt.Println()
	if brutal {
		fmt.Println("Brutal overlay: round scores may go negative; modifiers may land on busted players (tax from 0); Flip 7 is +15 **or** −15 to another cumulative total. Catalog Flip 7 policy: Cautious/Heuristic take +15; Aggressive/EV/Adaptive subtract 15 from the score leader.")
	} else {
		fmt.Println("Strategies: Cautious, Aggressive, Heuristic-N (number-sum stop), ExpectedValue (remaining-deck EEV), Adaptive (Aggressive when behind, otherwise EV).")
		fmt.Println("Targeting is shared: negative modifiers and Just One More go to the most threatening non-busted opponent; Flip Four prefers a high bust-risk opponent; Steal/Swap/Discard operate on face-up cards (dump The Zero, take Lucky 13 / high ranks).")
	}
	fmt.Println()

	fmt.Println("## 1. Heuristic stopping threshold (solo, fastest to 200)")
	fmt.Println()
	fmt.Println("| Threshold | Avg Rounds | Median | Finished |")
	fmt.Println("| :--- | ---: | ---: | ---: |")
	for _, st := range sweep {
		fmt.Printf("| %s | %.2f | %.2f | %d |\n", st.Name, st.Avg, st.Median, st.N)
	}
	fmt.Printf("\nSelected **Heuristic-%d** for the remaining tables (lowest average rounds among finished solos).\n\n", bestTh)

	fmt.Println("## 2. Single player (fastest to 200)")
	fmt.Println()
	fmt.Println("| Strategy | Avg Rounds | Median Rounds | Finished |")
	fmt.Println("| :--- | ---: | ---: | ---: |")
	for _, st := range single {
		fmt.Printf("| %s | %.2f | %.2f | %d |\n", st.Name, st.Avg, st.Median, st.N)
	}
	fmt.Println()

	fmt.Println("## 3. Multiplayer win rates")
	fmt.Println()
	fmt.Println("Each game assigns catalog strategies round-robin so every name appears even when N is smaller than the catalog.")
	fmt.Println()
	fmt.Print("| Strategy |")
	for _, t := range multi {
		fmt.Printf(" %d Players |", t.PlayerCount)
	}
	fmt.Println()
	fmt.Print("| :--- |")
	for range multi {
		fmt.Print(" ---: |")
	}
	fmt.Println()
	for _, name := range names {
		fmt.Printf("| %s |", name)
		for _, t := range multi {
			fmt.Printf(" %.2f%% |", t.Rates[name])
		}
		fmt.Println()
	}
	fmt.Println()

	fmt.Println("## 4. 1v1 matchups")
	fmt.Println()
	fmt.Println("Cell is row-strategy win rate against the column strategy.")
	fmt.Println()
	fmt.Print("| vs |")
	for _, name := range names {
		fmt.Printf(" %s |", name)
	}
	fmt.Println()
	fmt.Print("| :--- |")
	for range names {
		fmt.Print(" ---: |")
	}
	fmt.Println()
	rate := map[string]map[string]float64{}
	for _, p := range pairs {
		if rate[p.A] == nil {
			rate[p.A] = map[string]float64{}
		}
		if rate[p.B] == nil {
			rate[p.B] = map[string]float64{}
		}
		rate[p.A][p.B] = p.AWin
		rate[p.B][p.A] = p.BWin
	}
	for _, row := range names {
		fmt.Printf("| **%s** |", row)
		for _, col := range names {
			if row == col {
				fmt.Print(" - |")
				continue
			}
			fmt.Printf(" %.2f%% |", rate[row][col])
		}
		fmt.Println()
	}
	fmt.Println()

	fmt.Println("## 5. Flip Four risk threshold (ExpectedValue vs Aggressive, 1v1)")
	fmt.Println()
	fmt.Println("| Selector | EV win rate vs Aggressive |")
	fmt.Println("| :--- | ---: |")
	for _, r := range ff {
		fmt.Printf("| %s | %.2f%% |\n", r.Name, r.WinRate)
	}
	fmt.Println()

	fmt.Println("## Notes")
	fmt.Println()
	fmt.Println("- Stay does **not** lock the line. A high number total can still be stolen, swapped, or hit with a negative modifier before the round ends. Hit/Stay EV that ignores that understates take-that.")
	fmt.Println("- Solo play forces every Modifier and most Actions onto yourself (only non-busted player). Rankings there are not the same as 3–5 player tables.")
	fmt.Println("- Stub-level targeting is replaced here; these AIs dump −N / ÷2 and Just One More on the current threat, and treat Lucky 13 / The Zero as card-level tactics.")
	fmt.Println("- Original Flip 7 results remain in `docs/strategy_evaluation.md`.")
	if !brutal {
		return
	}
	fmt.Println()
	fmt.Println("## 6. Flip 7 +15 vs −15 (ExpectedValue, 1v1, Brutal)")
	fmt.Println()
	fmt.Println("| Policy | Win rate |")
	fmt.Println("| :--- | ---: |")
	fmt.Printf("| EV-Take15 | %.2f%% |\n", takeWin)
	fmt.Printf("| EV-Attack15 | %.2f%% |\n", attackWin)
	fmt.Println()
	if attackWin > takeWin {
		fmt.Println("Attacking the leader (−15 to their total) beat taking +15 in this 1v1.")
	} else if takeWin > attackWin {
		fmt.Println("Taking +15 beat attacking the leader in this 1v1.")
	} else {
		fmt.Println("Take +15 and attack the leader tied in this 1v1.")
	}
}
