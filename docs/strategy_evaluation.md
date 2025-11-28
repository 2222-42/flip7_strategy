# Strategy Evaluation Results

This document tracks the performance of different strategies in the Flip 7 game.

## Latest Simulation Run (Flip Three Bust Strategy Implemented)

**Date**: 2025-11-29
**Changes**: Implemented `EstimateFlipThreeRisk` and updated `Adaptive`, `ExpectedValue`, and `Heuristic` strategies to target high-risk opponents when using `Flip Three`.

### 1. Single Player Optimization (Fastest to 200 Points)

| Strategy | Avg Rounds | Median Rounds |
| :--- | :--- | :--- |
| **Adaptive** | **9.72** | **9.00** |
| Probabilistic | 9.79 | 10.00 |
| ExpectedValue | 9.81 | 9.00 |
| Heuristic-27 | 9.88 | 10.00 |
| Aggressive | 11.18 | 11.00 |
| Cautious | 12.26 | 12.00 |

*Adaptive strategy is now the fastest to reach 200 points.*

### 2. Multiplayer Evaluation (Win Rates)

Win rates in games with N players (1000 games each).

| Strategy | 2 Players | 3 Players | 4 Players | 5 Players |
| :--- | :--- | :--- | :--- | :--- |
| **Adaptive** | **21.80%** | **20.85%** | **20.60%** | **21.05%** |
| ExpectedValue | 15.25% | 19.80% | 20.25% | 20.20% |
| Heuristic-27 | 18.00% | 15.90% | 19.80% | 17.65% |
| Aggressive | 17.20% | 17.65% | 18.55% | 19.85% |
| Probabilistic | 16.60% | 17.85% | 15.75% | 16.40% |
| Cautious | 11.15% | 7.95% | 5.05% | 4.85% |

*Adaptive strategy consistently outperforms all others in multiplayer settings.*

### 3. Strategy Combination Evaluation (1vs1 Matchups)

Win rates for Strategy A (Row) vs Strategy B (Column).

| vs | Cautious | Aggressive | Probabilistic | Heuristic-27 | ExpectedValue | Adaptive |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **Cautious** | - | 13.90% | 6.70% | 8.80% | 7.60% | 8.40% |
| **Aggressive** | 86.10% | - | 46.20% | 43.55% | 42.85% | 40.45% |
| **Probabilistic** | 93.30% | 53.80% | - | 48.20% | 45.55% | 46.25% |
| **Heuristic-27** | 91.20% | 56.45% | 51.80% | - | 46.60% | 47.20% |
| **ExpectedValue** | 92.40% | 57.15% | 54.45% | 53.40% | - | 48.45% |
| **Adaptive** | **91.60%** | **59.55%** | **53.75%** | **52.80%** | **51.55%** | - |

*Adaptive strategy has a positive win rate (>50%) against ALL other strategies.*

## Conclusions

The implementation of the **Flip Three Bust Strategy** (targeting high-risk opponents) has significantly improved the performance of the strategies that use it (Adaptive, ExpectedValue, Heuristic, Probabilistic).

- **Adaptive Strategy** is currently the **dominant strategy**, winning both in speed and head-to-head matchups.
- **ExpectedValue** remains a very strong contender.
- **Cautious** strategy is too slow to compete effectively.
- **Aggressive** strategy is good against Cautious but loses to more calculated strategies.
