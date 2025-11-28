# Strategy Evaluation Results

## Single Player Optimization (Fastest to 200 Points)

We evaluated how many rounds it takes for a single player to reach 200 points using different strategies.

| Strategy | Average Rounds | Median Rounds |
| :--- | :--- | :--- |
| **Adaptive** | **9.49** | **9.00** |
| **ExpectedValue** | **9.51** | **9.00** |
| Probabilistic | 9.59 | 9.00 |
| Heuristic-27 | 9.63 | 9.00 |
| Aggressive | 10.41 | 10.00 |
| Cautious | 12.07 | 12.00 |

*Result: The **Adaptive** and **ExpectedValue** strategies are the most efficient, reaching 200 points faster than others.*

## Multiplayer Evaluation (Win Rates)

We simulated 1000 games for each player count to observe strategy performance in a competitive setting.

### 5 Players
- **Adaptive**: 21.95%
- **ExpectedValue**: 19.75%
- **Probabilistic**: 19.25%
- **Aggressive**: 17.80%
- **Heuristic-27**: 17.15%
- **Cautious**: 4.10%

*Result: **Adaptive** emerges as the strongest strategy in a 5-player game, followed closely by **ExpectedValue** and **Probabilistic**. The rotation of strategies ensures a fair comparison.*

## Strategy Combination Evaluation (1vs1)

We evaluated all unique pairs of strategies in 1vs1 matchups (1000 games each).

| Matchup | Winner | Win Rate | Loser | Win Rate |
| :--- | :--- | :--- | :--- | :--- |
| Cautious vs Aggressive | Aggressive | 75.55% | Cautious | 24.45% |
| Cautious vs Probabilistic | Probabilistic | 79.40% | Cautious | 20.60% |
| Cautious vs Heuristic-27 | Heuristic-27 | 81.40% | Cautious | 18.60% |
| **Cautious vs ExpectedValue** | **ExpectedValue** | **85.15%** | Cautious | 14.85% |
| Aggressive vs Probabilistic | Probabilistic | 52.25% | Aggressive | 47.75% |
| Aggressive vs Heuristic-27 | Heuristic-27 | 53.65% | Aggressive | 46.35% |
| **Aggressive vs ExpectedValue** | **ExpectedValue** | **54.10%** | Aggressive | 45.90% |
| **Aggressive vs Adaptive** | **Adaptive** | **55.70%** | Aggressive | 44.30% |
| Probabilistic vs Heuristic-27 | Heuristic-27 | 50.90% | Probabilistic | 49.10% |
| **Probabilistic vs ExpectedValue** | **ExpectedValue** | **52.95%** | Probabilistic | 47.05% |
| **Probabilistic vs Adaptive** | **Adaptive** | **52.70%** | Probabilistic | 47.30% |
| **Heuristic-27 vs ExpectedValue** | **ExpectedValue** | **51.90%** | Heuristic-27 | 48.10% |
| **Heuristic-27 vs Adaptive** | **Adaptive** | **53.50%** | Heuristic-27 | 46.50% |
| **ExpectedValue vs Adaptive** | **Adaptive** | **51.00%** | ExpectedValue | 49.00% |

*Result:*
- **Adaptive** is the dominant strategy in 1vs1, defeating all other strategies, including **ExpectedValue**.
- **ExpectedValue** remains very strong, beating everyone except Adaptive.
- **Heuristic-27** performs surprisingly well, beating Aggressive and Probabilistic.
