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

### 4. Target Selection Optimization (Flip Three Risk Thresholds)

We ran a comprehensive simulation (1000 games per batch) to find the optimal "Flip Three" risk threshold for each strategy. The risk threshold determines when to target an opponent: if their estimated bust risk is greater than the threshold, they are targeted.

#### 1. Expected Value Strategy
*Optimal Threshold: 0.70*

| Risk Threshold | Win Rate | vs Standard Aggressive |
| :--- | :--- | :--- |
| **0.70** | **15.10%** | -0.75% |
| 0.80 | 14.65% | -1.20% |
| 0.65 | 13.95% | -1.90% |
| 0.50 | 12.40% | -3.45% |
| 0.90 | 12.20% | -3.65% |

#### 2. Probabilistic Strategy
*Optimal Threshold: 0.50*

| Risk Threshold | Win Rate | vs Standard Aggressive |
| :--- | :--- | :--- |
| **0.50** | **14.95%** | -0.80% |
| 0.80 | 14.45% | -1.30% |
| 0.90 | 13.80% | -1.95% |
| 0.70 | 13.65% | -2.10% |
| 0.65 | 11.30% | -4.45% |

#### 3. Heuristic Strategy (Threshold 27)
*Optimal Threshold: 0.65*

| Risk Threshold | Win Rate | vs Standard Aggressive |
| :--- | :--- | :--- |
| **0.65** | **15.05%** | **+0.70%** |
| 0.70 | 14.30% | -0.05% |
| 0.90 | 13.20% | -1.15% |
| 0.50 | 13.00% | -1.35% |
| 0.80 | 12.95% | -1.40% |

#### 4. Aggressive Strategy
*Optimal Threshold: 0.90*

| Risk Threshold | Win Rate | vs Standard Aggressive |
| :--- | :--- | :--- |
| **0.90** | **15.20%** | **+1.55%** |
| 0.70 | 14.40% | +0.75% |
| 0.80 | 13.85% | +0.20% |
| 0.50 | 13.85% | +0.20% |
| 0.65 | 13.55% | -0.10% |

#### Key Findings & Analysis
- **Diverse Optima**: There is no single "best" risk threshold for all strategies.
    - **Expected Value (0.70)**: Benefits from a balanced approach.
    - **Probabilistic (0.50)**: Benefits from aggressive targeting (targeting anyone with >50% risk), likely because the strategy itself is conservative in play, so aggressive targeting compensates.
    - **Heuristic (0.65)**: Benefits from a moderately aggressive threshold. This combination yielded the highest win rate against the standard Aggressive baseline in its batch.
    - **Aggressive (0.90)**: Benefits from a very conservative targeting threshold. Since the strategy plays aggressively (taking risks itself), it seems beneficial to only target opponents who are almost guaranteed to bust (>90%), ensuring the "Flip Three" is effective.

- **Synergy**: The results suggest a synergy between playing style and targeting style.
    - **Conservative Play + Aggressive Targeting**: Probabilistic (Conservative) + 0.50 (Aggressive Targeting) works well.
    - **Aggressive Play + Conservative Targeting**: Aggressive (Aggressive) + 0.90 (Conservative Targeting) works well.

### 5. Final Evaluation with Optimal Risk Thresholds

We re-ran the Multiplayer and Strategy Combination evaluations using the optimal risk thresholds identified above:
- **Expected Value**: 0.70
- **Probabilistic**: 0.50
- **Heuristic**: 0.65
- **Aggressive**: 0.90

#### Multiplayer Evaluation (Win Rates)

| Strategy | 2 Players | 3 Players | 4 Players | 5 Players |
| :--- | :--- | :--- | :--- | :--- |
| **Adaptive** | **19.95%** | **20.50%** | 19.00% | 19.05% |
| **ExpectedValue** | 17.00% | 19.55% | 19.45% | **20.10%** |
| **Heuristic-27** | 15.75% | 17.85% | **20.00%** | 19.45% |
| Probabilistic | 18.20% | 16.55% | 16.20% | 18.30% |
| Aggressive | 17.70% | 17.70% | 19.20% | 18.30% |
| Cautious | 11.40% | 7.85% | 6.15% | 4.80% |

*Analysis*:
- **Adaptive** remains dominant in 2 and 3 player games.
- **Heuristic-27** (with risk-based targeting) has surged to become the best strategy in 4-player games.
- **ExpectedValue** (with risk-based targeting) is the best strategy in 5-player games.
- The competition is much tighter with optimized targeting.

#### Strategy Combination Evaluation (1vs1 Matchups)

Win rates for Strategy A (Row) vs Strategy B (Column).

| vs | Cautious | Aggressive | Probabilistic | Heuristic-27 | ExpectedValue | Adaptive |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| **Cautious** | - | 38.30% | 25.35% | 27.35% | 23.05% | 24.70% |
| **Aggressive** | 61.70% | - | 47.40% | 45.00% | 43.75% | 42.65% |
| **Probabilistic** | **74.65%** | **52.60%** | - | 48.80% | 46.80% | 47.25% |
| **Heuristic-27** | **72.65%** | **55.00%** | **51.20%** | - | 48.65% | 46.50% |
| **ExpectedValue** | **76.95%** | **56.25%** | **53.20%** | **51.35%** | - | **51.55%** |
| **Adaptive** | **75.30%** | **57.35%** | **52.75%** | **53.50%** | 48.45% | - |

*Analysis*:
- **ExpectedValue** (optimized) has reclaimed the top spot in head-to-head matchups! It now beats **Adaptive** (51.55%) and all other strategies.
- **Adaptive** is a close second, beating everyone except ExpectedValue.
- **Heuristic-27** performs very well, beating Aggressive and Probabilistic.
- **Aggressive** struggles against optimized opponents in 1v1.

## Final Conclusion

Target selection optimization has significantly shifted the balance of power.

1.  **Expected Value (Risk 0.70)** is the **strongest 1v1 strategy**, defeating the Adaptive strategy. It also excels in large multiplayer games (5 players).
2.  **Adaptive Strategy** remains extremely competitive, dominating small multiplayer games (2-3 players) and being the second-best 1v1 strategy.
3.  **Heuristic-27 (Risk 0.65)** is a surprise contender, performing exceptionally well in 4-player games and holding its own in 1v1.
4.  **Risk-Based Targeting Matters**: Optimizing the "Flip Three" risk threshold provided a tangible edge, allowing ExpectedValue to overcome Adaptive in direct confrontation.
