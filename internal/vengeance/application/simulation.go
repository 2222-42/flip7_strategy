package application

import (
	"fmt"
	"sort"
	"strings"

	"flip7_strategy/internal/vengeance/domain"
	"flip7_strategy/internal/vengeance/domain/strategy"
)

type StrategySpec struct {
	Name string
	New  func() domain.Strategy
}

func DefaultCatalog() []StrategySpec {
	return []StrategySpec{
		{Name: "Cautious", New: func() domain.Strategy { return strategy.NewCautiousStrategy() }},
		{Name: "Aggressive", New: func() domain.Strategy { return strategy.NewAggressiveStrategy() }},
		{Name: "Heuristic-26", New: func() domain.Strategy { return strategy.NewHeuristicStrategy(strategy.DefaultHeuristicThreshold) }},
		{Name: "ExpectedValue", New: func() domain.Strategy { return strategy.NewExpectedValueStrategy() }},
		{Name: "Adaptive", New: func() domain.Strategy { return strategy.NewAdaptiveStrategy() }},
	}
}

func CatalogWithHeuristic(threshold int) []StrategySpec {
	c := DefaultCatalog()
	for i := range c {
		if strings.HasPrefix(c[i].Name, "Heuristic-") {
			c[i].Name = fmt.Sprintf("Heuristic-%d", threshold)
			th := threshold
			c[i].New = func() domain.Strategy { return strategy.NewHeuristicStrategy(th) }
		}
	}
	return c
}

type SimulationService struct {
	Rules domain.GameRules
}

func NewSimulationService() *SimulationService {
	return &SimulationService{}
}

func (s *SimulationService) runGame(players []*domain.Player) *domain.Game {
	game := domain.NewGame(players)
	game.Rules = s.Rules
	svc := NewGameService(game)
	svc.Silent = true
	svc.RunGame()
	return game
}

type NamedStat struct {
	Name   string
	Avg    float64
	Median float64
	N      int
}

func median(vals []int) float64 {
	if len(vals) == 0 {
		return 0
	}
	sort.Ints(vals)
	if len(vals)%2 == 0 {
		return float64(vals[len(vals)/2-1]+vals[len(vals)/2]) / 2
	}
	return float64(vals[len(vals)/2])
}

func (s *SimulationService) SinglePlayer(n int, catalog []StrategySpec) []NamedStat {
	out := make([]NamedStat, 0, len(catalog))
	for _, spec := range catalog {
		rounds := make([]int, 0, n)
		for i := 0; i < n; i++ {
			p := domain.NewPlayer(spec.Name, spec.New())
			g := s.runGame([]*domain.Player{p})
			if p.TotalScore >= domain.WinningThreshold {
				rounds = append(rounds, g.RoundCount)
			}
		}
		sum := 0
		for _, r := range rounds {
			sum += r
		}
		avg := 0.0
		if len(rounds) > 0 {
			avg = float64(sum) / float64(len(rounds))
		}
		out = append(out, NamedStat{Name: spec.Name, Avg: avg, Median: median(rounds), N: len(rounds)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Avg < out[j].Avg })
	return out
}

func (s *SimulationService) HeuristicSweep(gamesPerThreshold int, from, to, step int) []NamedStat {
	var out []NamedStat
	for th := from; th <= to; th += step {
		threshold := th
		spec := StrategySpec{
			Name: fmt.Sprintf("Heuristic-%d", threshold),
			New:  func() domain.Strategy { return strategy.NewHeuristicStrategy(threshold) },
		}
		stats := s.SinglePlayer(gamesPerThreshold, []StrategySpec{spec})
		out = append(out, stats...)
	}
	return out
}

type WinTable struct {
	PlayerCount int
	Rates       map[string]float64
}

func (s *SimulationService) Multiplayer(n, minPlayers, maxPlayers int, catalog []StrategySpec) []WinTable {
	var tables []WinTable
	for pc := minPlayers; pc <= maxPlayers; pc++ {
		wins := make(map[string]float64)
		for i := 0; i < n; i++ {
			players := make([]*domain.Player, 0, pc)
			for j := 0; j < pc; j++ {
				spec := catalog[(i+j)%len(catalog)]
				players = append(players, domain.NewPlayer(fmt.Sprintf("P%d-%s", j+1, spec.Name), spec.New()))
			}
			g := s.runGame(players)
			if len(g.Winners) == 0 {
				continue
			}
			pts := 1.0 / float64(len(g.Winners))
			for _, w := range g.Winners {
				wins[w.Strategy.Name()] += pts
			}
		}
		rates := make(map[string]float64, len(wins))
		for name, c := range wins {
			rates[name] = c / float64(n) * 100
		}
		tables = append(tables, WinTable{PlayerCount: pc, Rates: rates})
	}
	return tables
}

type PairResult struct {
	A, B       string
	AWin, BWin float64
}

func (s *SimulationService) OneVsOne(n int, catalog []StrategySpec) []PairResult {
	var out []PairResult
	for i := 0; i < len(catalog); i++ {
		for j := i + 1; j < len(catalog); j++ {
			a, b := catalog[i], catalog[j]
			wins := map[string]float64{}
			for k := 0; k < n; k++ {
				p1 := domain.NewPlayer(a.Name, a.New())
				p2 := domain.NewPlayer(b.Name, b.New())
				g := s.runGame([]*domain.Player{p1, p2})
				if len(g.Winners) == 0 {
					continue
				}
				pts := 1.0 / float64(len(g.Winners))
				for _, w := range g.Winners {
					wins[w.Name] += pts
				}
			}
			out = append(out, PairResult{
				A:    a.Name,
				B:    b.Name,
				AWin: wins[a.Name] / float64(n) * 100,
				BWin: wins[b.Name] / float64(n) * 100,
			})
		}
	}
	return out
}

type TargetResult struct {
	Name    string
	WinRate float64
}

func (s *SimulationService) FlipFourThresholds(n int, thresholds []float64) []TargetResult {
	var out []TargetResult
	for _, th := range thresholds {
		threshold := th
		name := fmt.Sprintf("EV-FF-%.2f", threshold)
		wins := 0.0
		for i := 0; i < n; i++ {
			ev := strategy.NewExpectedValueStrategyWithRisk(threshold)
			agg := strategy.NewAggressiveStrategy()
			p1 := domain.NewPlayer(name, ev)
			p2 := domain.NewPlayer("Aggressive", agg)
			g := s.runGame([]*domain.Player{p1, p2})
			for _, w := range g.Winners {
				if w.Name == name {
					wins += 1.0 / float64(len(g.Winners))
				}
			}
		}
		out = append(out, TargetResult{Name: name, WinRate: wins / float64(n) * 100})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].WinRate > out[j].WinRate })
	return out
}

func (s *SimulationService) Flip7TakeVsAttack(n int) (takeWin, attackWin float64) {
	takeWins, attackWins := 0.0, 0.0
	for i := 0; i < n; i++ {
		take := strategy.NewExpectedValueStrategyTakeBonus()
		attack := strategy.NewExpectedValueStrategy()
		p1 := domain.NewPlayer("EV-Take15", take)
		p2 := domain.NewPlayer("EV-Attack15", attack)
		g := s.runGame([]*domain.Player{p1, p2})
		for _, w := range g.Winners {
			pts := 1.0 / float64(len(g.Winners))
			if w.Name == "EV-Take15" {
				takeWins += pts
			}
			if w.Name == "EV-Attack15" {
				attackWins += pts
			}
		}
	}
	return takeWins / float64(n) * 100, attackWins / float64(n) * 100
}
