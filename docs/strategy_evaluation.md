# Strategy Evaluation Results

## Single Player Optimization (Fastest to 200 Points)

We evaluated how many rounds it takes for a single player to reach 200 points using different strategies.

| Strategy | Average Rounds | Median Rounds |
| :--- | :--- | :--- |
| **Heuristic-27** | **10.65** | **10.00** |
| Probabilistic | 10.87 | 11.00 |
| Aggressive | 11.16 | 11.00 |
| Cautious | 13.62 | 14.00 |

*Result: The **Heuristic-27** strategy (stopping at sum >= 27) is the most efficient for speed, followed closely by the Probabilistic strategy.*

## Multiplayer Evaluation (Win Rates)

We simulated 1000 games for each player count to observe strategy performance in a competitive setting.

### 3 Players
- **Aggressive**: 46.20%
- **Probabilistic**: 43.10%
- **Cautious**: 10.70%

### 4 Players
- **Aggressive**: 37.15%
- **Heuristic-27**: 28.90%
- **Probabilistic**: 28.05%
- **Cautious**: 5.90%


*Result: The **Aggressive** strategy tends to dominate in multiplayer settings, likely because it pushes for higher scores per round, which is necessary to beat opponents before they reach 200. The **Cautious** strategy falls behind significantly as the number of players increases.*

## Strategy Combination Evaluation (1vs1)

We evaluated all unique pairs of strategies in 1vs1 matchups (1000 games each).

| Matchup | Winner | Win Rate | Loser | Win Rate |
| :--- | :--- | :--- | :--- | :--- |
| Cautious vs Aggressive | Aggressive | 72.90% | Cautious | 27.10% |
| Cautious vs Probabilistic | Probabilistic | 79.90% | Cautious | 20.10% |
| Cautious vs Heuristic-27 | Heuristic-27 | 76.60% | Cautious | 23.40% |
| Aggressive vs Probabilistic | Aggressive | 50.30% | Probabilistic | 49.70% |
| Aggressive vs Heuristic-27 | Aggressive | 54.65% | Heuristic-27 | 45.35% |
| Probabilistic vs Heuristic-27 | Heuristic-27 | 50.50% | Probabilistic | 49.50% |

*Result:*
- **Cautious** is consistently beaten by all other strategies in 1vs1.
- **Aggressive** has a slight edge over **Probabilistic** and **Heuristic-27**.
- **Probabilistic** and **Heuristic-27** are very evenly matched.

