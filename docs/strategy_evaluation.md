# Strategy Evaluation Results

## Single Player Optimization (Fastest to 200 Points)

We evaluated how many rounds it takes for a single player to reach 200 points using different strategies.

| Strategy | Average Rounds | Median Rounds |
| :--- | :--- | :--- |
| **ExpectedValue** | **9.40** | **9.00** |
| Heuristic-27 | 9.96 | 10.00 |
| Probabilistic | 10.04 | 10.00 |
| Aggressive | 10.27 | 10.00 |
| Cautious | 13.07 | 13.00 |

*Result: The **ExpectedValue** strategy is the most efficient, reaching 200 points faster than any other strategy.*

## Multiplayer Evaluation (Win Rates)

We simulated 1000 games for each player count to observe strategy performance in a competitive setting.

### 5 Players
- **Aggressive**: 27.40%
- **ExpectedValue**: 26.90%
- **Heuristic-27**: 22.40%
- **Probabilistic**: 20.95%
- **Cautious**: 2.35%

*Result: In a crowded game (5 players), **Aggressive** and **ExpectedValue** are the top contenders. Aggressive takes slightly more risks which pays off when multiple opponents are racing for the win.*

## Strategy Combination Evaluation (1vs1)

We evaluated all unique pairs of strategies in 1vs1 matchups (1000 games each).

| Matchup | Winner | Win Rate | Loser | Win Rate |
| :--- | :--- | :--- | :--- | :--- |
| Cautious vs Aggressive | Aggressive | 75.55% | Cautious | 24.45% |
| Cautious vs Probabilistic | Probabilistic | 79.40% | Cautious | 20.60% |
| Cautious vs Heuristic-27 | Heuristic-27 | 81.40% | Cautious | 18.60% |
| **Cautious vs ExpectedValue** | **ExpectedValue** | **85.15%** | Cautious | 14.85% |
| Aggressive vs Probabilistic | Probabilistic | 51.75% | Aggressive | 48.25% |
| Aggressive vs Heuristic-27 | Heuristic-27 | 50.95% | Aggressive | 49.05% |
| **Aggressive vs ExpectedValue** | **ExpectedValue** | **53.30%** | Aggressive | 46.70% |
| Probabilistic vs Heuristic-27 | Heuristic-27 | 51.05% | Probabilistic | 48.95% |
| **Probabilistic vs ExpectedValue** | **ExpectedValue** | **53.80%** | Probabilistic | 46.20% |
| **Heuristic-27 vs ExpectedValue** | **ExpectedValue** | **55.60%** | Heuristic-27 | 44.40% |
| **Adaptive vs ExpectedValue** | **Adaptive** | **52.50%** | ExpectedValue | 47.50% |
| **Adaptive vs Aggressive** | **Adaptive** | **54.70%** | Aggressive | 45.30% |
| **Adaptive vs Probabilistic** | **Adaptive** | **56.00%** | Probabilistic | 44.00% |
| **Adaptive vs Heuristic-27** | **Adaptive** | **54.60%** | Heuristic-27 | 45.40% |

*Result:*
- **Adaptive** is the new dominant strategy in 1vs1, defeating **ExpectedValue**, **Aggressive**, and all others.
- **ExpectedValue** remains very strong, beating everyone except Adaptive.
- **Adaptive** works by playing efficiently (like ExpectedValue) but switching to high-risk/high-reward (Aggressive) when an opponent threatens to win (score > 200).
