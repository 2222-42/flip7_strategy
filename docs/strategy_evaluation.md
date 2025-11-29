# Strategy Evaluation Results

This document tracks the performance of different strategies in the Flip 7 game.

## Latest Simulation Run (Refined Freeze Strategy)

**Date**: 2025-11-30
**Changes**: Refined `Freeze` strategy to target Self if winning AND high risk, otherwise target opponent with highest score.

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
| **Adaptive** | **19.35%** | **21.60%** | **19.20%** | **21.70%** |
| ExpectedValue | 18.75% | 18.10% | 17.00% | 20.15% |
| Aggressive | 18.40% | 18.10% | 17.80% | 16.65% |
| Probabilistic | 18.10% | 16.50% | 17.75% | 17.15% |
| Heuristic-27 | 14.75% | 18.50% | 21.35% | 19.05% |
| Cautious | 10.65% | 7.20% | 6.90% | 5.30% |

*Adaptive strategy remains the most consistent performer, though Heuristic-27 spiked in 4-player games.*

### 3. Strategy Combination Evaluation (1vs1 Matchups)

Win rates for Strategy A (Row) vs Strategy B (Column).

| vs | Cautious | Aggressive | Probabilistic | Heuristic-27 | ExpectedValue | Adaptive |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **Cautious** | - | 13.90% | 6.70% | 8.80% | 7.60% | 8.40% |
| **Aggressive** | 86.10% | - | 43.05% | 41.10% | 41.30% | 42.35% |
| **Probabilistic** | 93.30% | 56.95% | - | 47.75% | 50.40% | 48.05% |
| **Heuristic-27** | 91.20% | 58.90% | 52.25% | - | 48.50% | 46.60% |
| **ExpectedValue** | 92.40% | 58.70% | 49.60% | 51.50% | - | 47.85% |
| **Adaptive** | **91.60%** | **57.65%** | **51.95%** | **53.40%** | **52.15%** | - |

*Adaptive strategy has regained its dominance, winning >50% against ALL other strategies, including ExpectedValue (52.15%).*

## Conclusions

The **Refined Freeze Strategy** (risk-aware targeting) has solidified the **Adaptive Strategy**'s position as the best overall strategy.

- **Adaptive Strategy** is the **undisputed champion**, winning all head-to-head matchups and performing consistently well in multiplayer.
- **ExpectedValue** remains strong but lost its slight edge against Adaptive.
- **Aggressive** and **Heuristic-27** are viable but less consistent.
- **Cautious** remains the weakest.
