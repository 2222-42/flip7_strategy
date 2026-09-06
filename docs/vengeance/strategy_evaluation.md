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
- **Cautious** is still weak in mixed games (often <10%) but can beat Aggressive because Vengeance lines are attacked after Stay — pushing for Flip 7 is less free than in original Flip 7.
- Flip Four risk 0.30–0.70 barely moved EV vs Aggressive (~76–78.5%). Hit/Stay and modifier dumping matter more than that threshold in this matchup.

Take-that changes the original ranking: there is no Freeze to bank, so Aggressive’s extra cards are gifts to Steal / Discard / −N, and EV/Heuristic that stop on the number sum stay on top.
