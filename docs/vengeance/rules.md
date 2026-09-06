# Flip 7: With a Vengeance Rules

Authoritative rules for the `Flip7Vengeance` bounded context. Implementation tests in issue #96 must encode these invariants, not re-interpret the PDF.

**Primary source:** [Flip 7 With a Vengeance — Ruleset Edition 1](https://cdn.shopify.com/s/files/1/0611/3958/3198/files/26_FLIP_7_VENGEANCE_RULES_C.pdf?v=1770853609) (The Op Games / USAopoly, 2026).

**Design:** [`domain_model.md`](domain_model.md). Where this file and the model disagree, **this file wins**; update the model.

Standard rules only. Brutal Mode is summarized at the end and is **not** in the engine until a follow-up issue.

## Terms

| Term | Meaning |
| :--- | :--- |
| **Hit** | Take a card from the deck. |
| **Stay** | Do not take a card. Turn the leftmost card in your number row sideways. Your cards stay in the round (face up, targetable). |
| **Bust** | Receive a second card of a Number you already have (Lucky 13 / Unlucky 7 exceptions below). Flip your cards **face down**; they are out of play. Score 0 this round. |
| **Active** | Has not stayed and has not busted. |
| **Round** | Dealing until every player has busted or stayed, or someone Flips 7. |
| **Flip 7** | Seven **different** Number ranks in your line. Ends the round immediately. +15 bonus. |
| **Line** | Number cards in a row; Modifier cards above them. Actions never stay in the line. |

Cards are not safe until the round ends. Stay does not bank.

## Deck (108 cards)

This composition matches the printed sheet, the line "thirteen 13's … one 1", and the box "CONTENTS: 108 CARDS".

| Card | Copies | Notes |
| :--- | ---: | :--- |
| 13 regular | 12 | + Lucky 13 = thirteen 13's |
| 12 | 12 | |
| 11 | 11 | |
| 10 | 10 | |
| 9 | 9 | |
| 8 | 8 | |
| 7 regular | 6 | + Unlucky 7 = seven 7's |
| 6 | 6 | |
| 5 | 5 | |
| 4 | 4 | |
| 3 | 3 | |
| 2 | 2 | |
| 1 | 1 | |
| The Zero | 1 | Special Number, rank 0. No regular 0. |
| Unlucky 7 | 1 | Special Number, rank 7 |
| Lucky 13 | 1 | Special Number, rank 13 |
| −2, −4, −6, −8, −10, ÷2 | 1 each | 6 modifiers |
| Just One More, Flip Four, Swap, Steal, Discard | 2 each | 10 actions |

89 regular numbers + 3 specials + 6 modifiers + 10 actions = **108**.

More than 18 players: use a second deck (216 cards). Simulation default is one deck.

## Playing a round

1. Shuffle thoroughly. Choose a Dealer (round 1: chosen / seat 0 in sim).
2. **Initial deal:** starting with the player on the Dealer's **left**, clockwise, each player receives one face-up card. Action and Modifier pause and resolve immediately. Some players may have several cards and others none, depending on those resolves.
3. The Dealer then offers Hit or Stay. Offer order is the same clockwise ring, starting left of the Dealer, skipping non-active players. (This differs from original Flip 7, where the Dealer is first.)
4. On Hit, draw one card and resolve it immediately (except Flip Four's delayed Action/Modifier, below).
5. Repeat until a round-end condition.

A player with The Zero, on their own Hit/Stay decision, **must Hit** (unless a Just One More force-Stay already applied this round; see interpretations).

Stay with an empty line (only discarded actions so far) is legal: set status to stayed; there is no card to turn sideways.

### End of a round

Either:

1. Every player has busted or stayed, or
2. One player has Flip 7 (seven different Number ranks) — round ends **immediately**.

Then score every non-busted line. Busted players score 0.

### Starting the next round

Set all cards from the round aside. **Do not** shuffle them back. Pass the remaining deck left; that player is the new Dealer.

When the deck runs out, shuffle the discarded cards into a new deck. Mid-round reshuffle: leave every card in front of players where it is, including busted (face-down) cards.

### End of the game

After a round, if at least one player has 200 or more, the player with the **most** points wins. Ties for that highest score: all tied players are winners (PDF is silent; same as original `DetermineWinners`).

## Scoring (standard)

Order is mandatory. Official example: 3+11+5+7+10+8+4 = 48, ÷2 → 24, −4 → 20, Flip 7 +15 → **35**.

1. Sum Number ranks in the line (The Zero contributes 0; Lucky 13 and Unlucky 7 contribute 13 and 7).
2. If ÷2 is in the modifier row: `floor(sum / 2)`.
3. Subtract −2…−10. Result cannot go below 0.
4. If Flip 7, add 15.

Busted: 0. The Zero without Flip 7: total **0** (overrides steps 1–3). The Zero **with** Flip 7: do steps 1–4 as usual (no zero override).

Only one ÷2 exists in a single deck.

## Special Number cards

They are Number cards: they occupy the number row, count toward Flip 7, score their rank, and can be stolen or swapped. Effects start when you get the card and **stop when it leaves** your line.

### The Zero (rank 0)

- Round total becomes 0 unless you Flip 7.
- Counts as one of the seven Number cards (rank 0 is distinct from 1–13).
- While you possess it, you must Hit on your turn.
- Just One More on a Zero holder is explicit in the PDF ("it might not be possible to flip to 7 Number cards"): the forced Stay wins; they take the one card then Stay even if Zero remains.

### Unlucky 7 (rank 7)

- On receipt: **do not bust**, even if a 7 is already in the line.
- Discard all Number and Modifier cards in front of you; keep Unlucky 7 only.
- During Flip Four: discard all previous cards first, then finish the remaining flips. Unlucky 7 counts as one of the four.
- Protection is **on receipt of Unlucky 7 only**. A later regular 7 while Unlucky 7 is in the line busts.

### Lucky 13 (rank 13)

- You may hold a second 13 without busting (Lucky 13 plus one regular 13, either order).
- Both 13s score (13+13). Rank 13 still counts as **one** distinct rank for Flip 7.
- A third 13 busts.
- There is only one Lucky 13. The second 13 is always a regular 13.

## Action cards

Common rules:

- Resolve immediately when revealed (Flip Four defers Action/Modifier until its four draws finish).
- Play on any player who has **not busted**, including yourself and anyone who stayed (or was forced to stay).
- If you are the only non-busted player, play it on yourself.
- A stayed recipient must resolve it and does **not** re-enter Active.
- Single-use: discard after play.
- The player who **receives** the Action (the one it was played on, or who flipped it onto themselves) is the **resolver** and makes its targeting choices.
- Beginning of a round, Swap / Steal / Discard with **no face-up card** on the table: discard the Action with no effect. Same if a resolver is forced to play one of these and no legal card exists (e.g. empty lines).

### Just One More (×2)

Force any non-busted player to accept the **next** deck card.

- If that card is an Action, **that player** (the forced one) resolves it.
- If it is a Modifier, that player is the default recipient unless the modifier assigner is the forced player choosing another non-busted target — treat it as the forced player receiving a normal flip: they assign the Modifier (they are the actor).
- Then the forced player **must Stay** (if they did not bust and the round did not already end on Flip 7).
- This Stay overrides The Zero's Must Hit.

### Swap (×2)

Swap any two **face-up** cards, either:

- one of yours with another player's, or
- two cards belonging to two other players.

Not allowed: two cards in the same line; any face-down (busted) card; swapping with the deck or discard.

Number ↔ Modifier is allowed (both are face-up cards). After the swap, each new owner **receives** the incoming card (special effects, duplicate 13, Unlucky 7 wipe, possible bust / Flip 7).

### Steal (×2)

Take any one face-up card on the table and add it to **your** (the resolver's) line. The victim loses it; their special effects stop. You `Receive` it immediately (Zero / Unlucky 7 / Lucky 13 / duplicate / Flip 7 all apply to you).

### Discard (×2)

Choose any non-busted player and one of their face-up cards; that card goes to the discard pile. Special effects stop for the owner. Discarding Zero ends Must Hit. Discarding a Number can break a would-be Flip 7 (if the round has not already ended).

### Flip Four (×2)

Force any non-busted player to accept the next **four** cards, one at a time.

- Stop early on Flip 7 or bust.
- Number, Action, and Modifier **all** count toward the four.
- If the target has not busted and did not Flip 7: Action and Modifier flipped during the four are resolved **in draw order after** the four cards.
- If the target busts: do not resolve those delayed cards; discard them.
- If the target Flips 7: the round ends immediately. Do not draw the rest of the four. Do not resolve delayed Action/Modifier from this Flip Four (Flip 7 already ended the round; later Steal/Discard must not unmake it).
- Nested Action (another Flip Four, Just One More, …) is resolved by the Flip Four **target** after the four, if they survived without Flip 7.
- Unlucky 7: wipe previous cards first, then continue remaining flips.

## Modifier cards

- Play on any non-busted player, including stayed. If you are the only non-busted player, keep it.
- Not Number cards: no Flip 7 credit, no bust.
- May be swapped or discarded.
- ÷2: divide the Number sum by two (round down), **then** subtract −2…−10.
- −2…−10: subtract from the (possibly halved) Number sum.
- Standard mode: round score cannot go below 0.
- At initial deal there are always non-busted players, so a Modifier is never "no target". Unlike Swap/Steal/Discard, do not discard a Modifier for lack of cards.

## Original Flip 7 vs Vengeance

| Topic | Original | Vengeance |
| :--- | :--- | :--- |
| Deck size | 94 | 108 |
| Numbers | 0–12 (n×n, 0×1) | 1–13 (7×6, 13×12) + Zero, Unlucky 7, Lucky 13 |
| Modifiers | +2…+10, ×2 | −2…−10, ÷2 |
| Actions | Freeze, Flip Three, Second Chance ×3 | Just One More, Flip Four, Swap, Steal, Discard ×2 |
| Score formula | `(sum × ×2) + adds + 15` | `max(0, floor(sum/2?) − subs) + 15` |
| Banking | Stay / Freeze / Flip 7 | Round end only |
| Stay | Leaves the round; line still scores | Line stays in the round and can still be attacked |
| Bust | Duplicate number unless Second Chance | Duplicate rank; Lucky 13 allows a second 13; Unlucky 7 does not bust on receipt |
| Action target | Player | Player (JOM, Flip Four, Modifier) and **card** (Swap, Steal, Discard) |
| Initial deal order | From Dealer | From Dealer's **left**, clockwise |
| Second Chance | Stays in hand | No such card; actions are discarded after play |
| Freeze | Banks and stops | Absent |

## Adopted interpretations (frozen)

These were the open questions on #93 / #95. Each line is now an invariant.

| # | Question | Decision | Why |
| :--- | :--- | :--- | :--- |
| 1 | Zero + Flip 7 scoring | Full formula (sum, ÷2, subtract, +15). Zero override **only** when Flip 7 is not achieved. | "Total score becomes zero, **unless** you can Flip 7." |
| 2 | Two 13s toward Flip 7 | Seven **distinct ranks**. Two 13s are one rank; both still add 13 to the sum. | Flip 7 is defined as "seven **different** Number cards." "Count toward the Flip 7 bonus" means they are Number cards, not two ranks. |
| 3 | Lucky 13 then regular 13, or the reverse | Either order. Second 13 does not bust while Lucky 13 is in the line. | "Collect a second 13 without busting." |
| 4 | Unlucky 7 onto an existing 7 | No bust; wipe; keep Unlucky 7 only. | "You cannot bust on an Unlucky 7 when you get it." |
| 5 | Regular 7 while holding Unlucky 7 | Bust. | Protection is receipt of Unlucky 7 only. |
| 6 | Unlucky 7 during Flip Four | Wipe first, then remaining flips. Counts as one of the four. | PDF Flip 4 clause on Unlucky 7. |
| 7 | Who resolves an Action drawn under JOM / Flip Four | The **recipient** (forced player / Flip Four target). | JOM: "If an Action card is flipped, **that player** can resolve it." |
| 8 | JOM Stay vs Zero Must Hit | JOM Stay wins. | PDF calls out JOM on a Zero holder making Flip 7 possibly impossible. |
| 9 | Swap pairs | Exactly two face-up cards on **two different** players: self↔other or other↔other. Same line illegal. Number↔Modifier legal. | "Either one of your own with another player, or cards between two other players." |
| 10 | Steal/Swap/Discard of specials | Legal if face up. Receiver `Receive`s immediately; giver's effect stops. | "Can be stolen or swapped"; effects last only while possessed. |
| 11 | Discard Zero | Legal. Must Hit ends. | Effect stops when it leaves. |
| 12 | Modifier with no other non-busted player | Keep it. Never discarded for "no cards." | "If you are the only player who hasn't busted, you must keep it." |
| 13 | Swap/Steal/Discard, no face-up card | Discard the Action, no effect. Applies at round start **and** whenever no legal card exists. | PDF beginning-of-round clause, extended to the empty-target case so the engine cannot stall. |
| 14 | Flip 7 mid Flip Four | Stop remaining draws. Skip delayed Action/Modifier from that Flip Four. Round over. | Flip 7 "ends the round immediately"; delayed Steal must not unmake it. |
| 15 | Game-end tie | All players tied for the highest score are winners. | PDF silent; matches original `DetermineWinners`. |
| 16 | Hit/Stay offer order | Clockwise from Dealer's left, same as the initial deal. | PDF deal is from the left; offers are not specified as Dealer-first. |

## Invariants for tests (#96)

Each item should be a unit or round test.

1. `NewDeck` has 108 cards and the per-rank counts in the deck table.
2. Score example: numbers {3,11,5,7,10,8,4} + ÷2 + −4 + Flip 7 → 35.
3. Score without Flip 7, same cards except one number missing: 48 → 24 → 20, no +15.
4. Standard score never negative: numbers {1} + −10 → 0.
5. Duplicate regular rank busts; line face down; round score 0.
6. Lucky 13 + one regular 13: no bust; `DistinctRanks` counts 13 once; sum includes 26.
7. Lucky 13 + two regular 13s: bust on the third 13.
8. Unlucky 7 onto a line that already has 7: no bust; only Unlucky 7 remains.
9. Unlucky 7 then regular 7: bust.
10. Zero without Flip 7: `Compute` total 0 even with other numbers/modifiers; `CanStay() == false`.
11. Zero with seven distinct ranks including 0: Flip 7, full formula +15, no zero override.
12. Stay: status stayed, leftmost number-like sideways, cards still `FaceUp`.
13. Bust: all cards `FaceUp == false`; not in `FaceUpCards()` for Steal/Swap/Discard.
14. Stayed player may receive Modifier and Action; stays stayed.
15. Only non-busted player must assign Modifier/Action to self.
16. Initial Swap/Steal/Discard with empty table: action discarded, deal continues.
17. Just One More: target draws one, then stayed (if alive); Zero holder included.
18. Flip Four: four draws or stop on bust / Flip 7; delayed Action/Modifier only if survived without Flip 7.
19. Flip Four + Unlucky 7 on draw 2 of 4: wipe, then draws 3 and 4.
20. Steal Number into a duplicate rank: thief busts (unless Lucky 13 exception).
21. Reshuffle uses discard only; table cards including busted stay put.
22. After a round with someone ≥200, `DetermineWinners` is the highest total(s).
23. Original Flip 7 package tests still pass (`go test ./...` from repo root).

## Brutal Mode (out of scope)

Not implemented. For later:

- Round score may go below 0.
- Modifier may be given to a **busted** player.
- On Flip 7, the player may take +15 **or** subtract 15 from another player.

Do not put these flags in `ScoreCalculator` / `ModifierAssigner` until that issue exists.
