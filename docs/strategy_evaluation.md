# Strategy Evaluation Results

This document tracks the performance of different strategies in the Flip 7 game.

## Latest Simulation Run (Freeze Strategy Update)

**Date**: 2025-11-30
**Changes**: Updated `Freeze` strategy to target the opponent with the highest score instead of self.

### 1. Single Player Optimization (Fastest to 200 Points)

*Note: Single player optimization results are unchanged as Freeze behavior in single player (fallback to self) remains the same.*

| Strategy | Avg Rounds | Median Rounds |
| :--- | :--- | :--- |
| **Adaptive** | **9.72** | **9.00** |
| Probabilistic | 9.79 | 10.00 |
| ExpectedValue | 9.81 | 9.00 |
| Heuristic-27 | 9.88 | 10.00 |
| Aggressive | 11.18 | 11.00 |
| Cautious | 12.26 | 12.00 |

### 2. Multiplayer Evaluation (Win Rates)

Win rates in games with N players (1000 games each).

| Strategy | 2 Players | 3 Players | 4 Players | 5 Players |
| :--- | :--- | :--- | :--- | :--- |
| **Adaptive** | **21.10%** | **20.80%** | **20.65%** | **21.05%** |
| ExpectedValue | 18.50% | 19.20% | 17.15% | 19.95% |
| Aggressive | 18.00% | 17.20% | 21.10% | 19.00% |
| Heuristic-27 | 17.20% | 17.65% | 21.20% | 17.65% |
| Probabilistic | 16.15% | 17.25% | 14.95% | 17.15% |
| Cautious | 9.05% | 7.90% | 4.95% | 5.20% |

*Adaptive strategy continues to perform well, maintaining a lead in most configurations.*

### 3. Strategy Combination Evaluation (1vs1 Matchups)

Win rates for Strategy A (Row) vs Strategy B (Column).

| vs | Cautious | Aggressive | Probabilistic | Heuristic-27 | ExpectedValue | Adaptive |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **Cautious** | - | 13.90% | 6.70% | 8.80% | 7.60% | 8.40% |
| **Aggressive** | 86.10% | - | 44.75% | 44.60% | 43.25% | 40.55% |
| **Probabilistic** | 93.30% | 55.25% | - | 49.00% | 49.30% | 47.10% |
| **Heuristic-27** | 91.20% | 55.40% | 51.00% | - | 48.80% | 46.65% |
| **ExpectedValue** | 92.40% | 56.75% | 50.70% | 51.20% | - | 50.70% |
| **Adaptive** | **91.60%** | **59.45%** | **52.90%** | **53.35%** | **49.30%** | - |

*Adaptive strategy remains dominant, though ExpectedValue has a slight edge against it in direct matchups (50.70% vs 49.30%).*

## Conclusions

The update to the **Freeze Strategy** (targeting opponents) has maintained the overall balance while potentially tightening the competition between top strategies.

- **Adaptive Strategy** is still the strongest overall performer.
- **ExpectedValue** has shown strong performance, slightly edging out Adaptive in direct 1v1.
- **Aggressive** and **Heuristic-27** remain competitive in larger groups.
- **Cautious** remains the weakest strategy.
