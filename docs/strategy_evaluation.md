# Strategy Evaluation Results

## Single Player Optimization (Fastest to 200 Points)

We evaluated how many rounds it takes for a single player to reach 200 points using different strategies.

| Strategy | Average Rounds | Median Rounds |
| :--- | :--- | :--- |
| **ExpectedValue** | **9.66** | **9.00** |
| **Adaptive** | **9.80** | **10.00** |
| Probabilistic | 9.80 | 10.00 |
| Heuristic-27 | 10.06 | 10.00 |
| Aggressive | 11.06 | 11.00 |
| Cautious | 12.27 | 12.00 |

*Result: The **ExpectedValue** strategy remains the most efficient, followed closely by **Adaptive** and **Probabilistic**.*

## Multiplayer Evaluation (Win Rates)

We simulated 1000 games for each player count to observe strategy performance in a competitive setting.

### 5 Players
- **ExpectedValue**: 21.70%
- **Adaptive**: 19.80%
- **Probabilistic**: 19.45%
- **Heuristic-27**: 18.45%
- **Aggressive**: 15.50%
- **Cautious**: 5.10%

*Result: **ExpectedValue** takes the lead in 5-player games, with **Adaptive** and **Probabilistic** following. **Cautious** remains the weakest despite the Second Chance improvement.*

## Strategy Combination Evaluation (1vs1)

We evaluated all unique pairs of strategies in 1vs1 matchups (1000 games each).

| Matchup | Winner | Win Rate | Loser | Win Rate |
| :--- | :--- | :--- | :--- | :--- |
| Cautious vs Aggressive | Aggressive | 75.30% | Cautious | 24.70% |
| Cautious vs Probabilistic | Probabilistic | 79.40% | Cautious | 20.60% |
| Cautious vs Heuristic-27 | Heuristic-27 | 81.65% | Cautious | 18.35% |
| **Cautious vs ExpectedValue** | **ExpectedValue** | **85.35%** | Cautious | 14.65% |
| Aggressive vs Probabilistic | Probabilistic | 56.60% | Aggressive | 43.40% |
| Aggressive vs Heuristic-27 | Heuristic-27 | 56.40% | Aggressive | 43.60% |
| **Aggressive vs ExpectedValue** | **ExpectedValue** | **56.65%** | Aggressive | 43.35% |
| **Aggressive vs Adaptive** | **Adaptive** | **58.80%** | Aggressive | 41.20% |
| Probabilistic vs Heuristic-27 | Heuristic-27 | 51.35% | Probabilistic | 48.65% |
| **Probabilistic vs ExpectedValue** | **ExpectedValue** | **52.95%** | Probabilistic | 47.05% |
| **Probabilistic vs Adaptive** | **Adaptive** | **52.55%** | Probabilistic | 47.45% |
| **Heuristic-27 vs ExpectedValue** | **ExpectedValue** | **52.25%** | Heuristic-27 | 47.75% |
| **Heuristic-27 vs Adaptive** | **Adaptive** | **50.70%** | Heuristic-27 | 49.30% |
| **ExpectedValue vs Adaptive** | **ExpectedValue** | **50.25%** | Adaptive | 49.75% |

*Result:*
- **ExpectedValue** and **Adaptive** are extremely close, with ExpectedValue slightly edging out Adaptive in their direct matchup (50.25% vs 49.75%).
- **ExpectedValue** wins against all other strategies.
- **Adaptive** wins against all strategies except ExpectedValue (narrow loss).
