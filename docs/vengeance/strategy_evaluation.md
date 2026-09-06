# Flip 7: With a Vengeance Strategy Evaluation

**Date**: 2026-09-05
**Games per cell**: 200 (heuristic sweep: 100 per threshold)
**Runtime**: 2s

Strategies: Cautious, Aggressive, Heuristic-N (number-sum stop), ExpectedValue (remaining-deck EEV), Adaptive (Aggressive when behind, otherwise EV).
Targeting is shared: negative modifiers and Just One More go to the most threatening non-busted opponent; Flip Four prefers a high bust-risk opponent; Steal/Swap/Discard operate on face-up cards (dump The Zero, take Lucky 13 / high ranks).

## 1. Heuristic stopping threshold (solo, fastest to 200)

| Threshold | Avg Rounds | Median | Finished |
| :--- | ---: | ---: | ---: |
| Heuristic-16 | 12.53 | 12.00 | 100 |
| Heuristic-18 | 12.08 | 12.00 | 100 |
| Heuristic-20 | 11.63 | 11.00 | 100 |
| Heuristic-22 | 11.13 | 11.00 | 100 |
| Heuristic-24 | 11.38 | 11.00 | 100 |
| Heuristic-26 | 10.89 | 10.50 | 100 |
| Heuristic-28 | 11.65 | 12.00 | 100 |
| Heuristic-30 | 11.69 | 11.00 | 100 |
| Heuristic-32 | 12.03 | 11.00 | 100 |

Selected **Heuristic-26** for the remaining tables (lowest average rounds among finished solos).

## 2. Single player (fastest to 200)

| Strategy | Avg Rounds | Median Rounds | Finished |
| :--- | ---: | ---: | ---: |
| ExpectedValue | 11.19 | 11.00 | 200 |
| Adaptive | 11.21 | 11.00 | 200 |
| Heuristic-26 | 11.37 | 11.00 | 200 |
| Cautious | 13.94 | 14.00 | 200 |
| Aggressive | 20.43 | 19.50 | 200 |

## 3. Multiplayer win rates

Each game assigns catalog strategies round-robin so every name appears even when N is smaller than the catalog.

| Strategy | 2 Players | 3 Players | 4 Players | 5 Players |
| :--- | ---: | ---: | ---: | ---: |
| Cautious | 11.00% | 8.00% | 2.50% | 4.00% |
| Aggressive | 16.00% | 11.00% | 13.50% | 9.50% |
| Heuristic-26 | 26.50% | 25.00% | 34.00% | 25.00% |
| ExpectedValue | 20.50% | 30.25% | 24.00% | 34.75% |
| Adaptive | 26.00% | 25.75% | 26.00% | 26.75% |

## 4. 1v1 matchups

Cell is row-strategy win rate against the column strategy.

| vs | Cautious | Aggressive | Heuristic-26 | ExpectedValue | Adaptive |
| :--- | ---: | ---: | ---: | ---: | ---: |
| **Cautious** | - | 55.50% | 14.50% | 12.75% | 23.00% |
| **Aggressive** | 44.50% | - | 21.25% | 28.00% | 25.75% |
| **Heuristic-26** | 85.50% | 78.75% | - | 45.50% | 45.25% |
| **ExpectedValue** | 87.25% | 72.00% | 54.50% | - | 50.00% |
| **Adaptive** | 77.00% | 74.25% | 54.75% | 50.00% | - |

## 5. Flip Four risk threshold (ExpectedValue vs Aggressive, 1v1)

| Selector | EV win rate vs Aggressive |
| :--- | ---: |
| EV-FF-0.30 | 78.50% |
| EV-FF-0.50 | 78.00% |
| EV-FF-0.70 | 76.00% |

## Notes

- Stay does **not** lock the line. A high number total can still be stolen, swapped, or hit with a negative modifier before the round ends. Hit/Stay EV that ignores that understates take-that.
- Solo play forces every Modifier and most Actions onto yourself (only non-busted player). Rankings there are not the same as 3–5 player tables.
- Stub-level targeting is replaced here; these AIs dump −N / ÷2 and Just One More on the current threat, and treat Lucky 13 / The Zero as card-level tactics.
- Original Flip 7 results remain in `docs/strategy_evaluation.md`.

## Conclusions

- **ExpectedValue** is the fastest solo (11.19 rounds) and the strongest at five players (34.75%). It also edges Heuristic-26 in 1v1 (54.50%).
- **Adaptive** matches EV when not behind, so the 1v1 vs EV is a coin flip (50%). It is the most even multiplayer profile (~26% at every table size).
- **Heuristic-26** is the best number-sum stop on the solo sweep and wins a lot of 4-player mixed tables (34%). It is simpler than EV and close on speed.
- **Aggressive** is punished: solo 20.4 rounds, and it loses 1v1 to Cautious (44.5%). Extra hits run into duplicates and Unlucky 7 without Freeze/Second Chance to bail out.

## Brutal Mode

**Date**: 2026-09-06  
**Games per cell**: 200 (heuristic sweep: 100 per threshold)  
**Runtime**: 7s  
**Command**: `go run cmd/evaluate_vengeance/main.go -n 200 -brutal`

Same catalog, overlay on: scores may go negative; modifiers may land on busted players (tax from base 0); Flip 7 is +15 **or** −15 to another cumulative total. Cautious and Heuristic take +15; Aggressive, ExpectedValue, and Adaptive subtract 15 from the score leader.

Standard tables above are unchanged.

### B1. Heuristic stopping threshold (solo)

| Threshold | Avg Rounds | Median | Finished |
| :--- | ---: | ---: | ---: |
| Heuristic-16 | 12.60 | 13.00 | 100 |
| Heuristic-18 | 12.03 | 12.00 | 100 |
| Heuristic-20 | 11.98 | 12.00 | 100 |
| Heuristic-22 | 11.35 | 11.00 | 100 |
| Heuristic-24 | 11.40 | 11.00 | 100 |
| Heuristic-26 | 12.22 | 12.00 | 100 |
| Heuristic-28 | 11.39 | 11.00 | 100 |
| Heuristic-30 | 11.79 | 11.00 | 100 |
| Heuristic-32 | 12.62 | 12.00 | 100 |

Selected **Heuristic-22** (lowest average rounds among finished solos). Standard's Heuristic-26 is slower here.

### B2. Single player (fastest to 200)

| Strategy | Avg Rounds | Median Rounds | Finished |
| :--- | ---: | ---: | ---: |
| Adaptive | 11.28 | 11.00 | 200 |
| Heuristic-22 | 11.46 | 11.00 | 200 |
| ExpectedValue | 11.64 | 12.00 | 200 |
| Cautious | 13.73 | 14.00 | 200 |
| Aggressive | 21.05 | 19.00 | 200 |

### B3. Multiplayer win rates

| Strategy | 2 Players | 3 Players | 4 Players | 5 Players |
| :--- | ---: | ---: | ---: | ---: |
| Cautious | 17.50% | 8.00% | 9.50% | 7.00% |
| Aggressive | 12.50% | 12.00% | 6.50% | 9.50% |
| Heuristic-22 | 22.25% | 23.50% | 25.75% | 23.50% |
| ExpectedValue | 23.75% | 28.25% | 30.25% | 35.50% |
| Adaptive | 24.00% | 28.25% | 28.00% | 24.50% |

### B4. 1v1 matchups

| vs | Cautious | Aggressive | Heuristic-22 | ExpectedValue | Adaptive |
| :--- | ---: | ---: | ---: | ---: | ---: |
| **Cautious** | - | 61.25% | 12.75% | 14.75% | 19.00% |
| **Aggressive** | 38.75% | - | 22.50% | 18.25% | 26.25% |
| **Heuristic-22** | 87.25% | 77.50% | - | 43.00% | 46.75% |
| **ExpectedValue** | 85.25% | 81.75% | 57.00% | - | 53.00% |
| **Adaptive** | 81.00% | 73.75% | 53.25% | 47.00% | - |

### B5. Flip Four risk (ExpectedValue vs Aggressive, 1v1)

| Selector | EV win rate vs Aggressive |
| :--- | ---: |
| EV-FF-0.70 | 84.50% |
| EV-FF-0.50 | 80.00% |
| EV-FF-0.30 | 78.00% |

### B6. Flip 7 +15 vs −15 (ExpectedValue, 1v1)

| Policy | Win rate |
| :--- | ---: |
| EV-Take15 | 47.00% |
| EV-Attack15 | 53.00% |

### Brutal conclusions

- **ExpectedValue** is the strongest mixed-table name (35.50% at 5p) and wins 1v1 vs Heuristic-22 (57%) and Adaptive (53%).
- **Adaptive** is the fastest solo (11.28) and the 2p leader (24%).
- **Heuristic-22** replaces Heuristic-26 as the Brutal stop; it is still the simple number-sum choice and close to EV.
- **Taking +15 lost to attacking the leader** in a Brutal EV 1v1 (47% vs 53%). Subtracting 15 from the other total is the stronger Flip 7 default when two EV players meet.
- **Aggressive** is still punished (solo 21.05 rounds; 38.75% vs Cautious).
- **Cautious** is still weak in mixed games (often <10%) but can beat Aggressive because Vengeance lines are attacked after Stay — pushing for Flip 7 is less free than in original Flip 7.
- Flip Four risk 0.30–0.70 barely moved EV vs Aggressive (~76–78.5%). Hit/Stay and modifier dumping matter more than that threshold in this matchup.

Take-that changes the original ranking: there is no Freeze to bank, so Aggressive’s extra cards are gifts to Steal / Discard / −N, and EV/Heuristic that stop on the number sum stay on top.
