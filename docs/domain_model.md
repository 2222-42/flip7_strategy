# Flip 7 Domain Model (DDD in Go)

This document outlines the Domain-Driven Design (DDD) model for the Flip 7 card game, tailored for implementation in Go. The model captures the core business rules from the provided PDF ruleset, emphasizing entities with identity, value objects for immutability, aggregates for consistency boundaries, and domain services for complex logic. 

Key DDD principles applied:
- **Bounded Context**: `Flip7Game` - Encompasses the entire game lifecycle, including rounds, players, and deck management.
- **Ubiquitous Language**: Terms like "Hit", "Stay", "Bust", "Flip 7", "Active Player" are used consistently.
- **Aggregates**: `Game` (root for overall consistency), `Round` (root for round-specific invariants like end conditions).
- **Repositories**: Outlined for persistence (e.g., `GameRepository` for loading/saving game state).
- **Focus on Counting/Strategy**: Deck tracks card counts for probability calculations; services advise on risks.

The model supports 3+ players, up to 18 (or more with multiple decks), ages 8+, ~20 min playtime. Goal: First to 200 points wins, with press-your-luck mechanics.

## Core Domain Concepts

### Value Objects

Value objects are immutable structs without identity (no ID field). They represent concepts like cards and scores.

**Card**  
Description: Represents a single card: Number (0-12 pts), Modifier (+2/+4/etc., x2), or Action (Freeze, FlipThree, SecondChance). Immutable.  
Go Struct Preview:  
```go
type Card struct {
    Type         CardType      `json:"type"`
    Value        NumberValue   // For Number cards
    ModifierType ModifierType  // For Modifier cards
    ActionType   ActionType    // For Action cards
}
```

**PointValue**  
Description: Immutable score calculation result.  
Go Struct Preview:  
```go
type PointValue struct {
    BaseSum    int    `json:"base_sum"`
    Modifiers  []int  `json:"modifiers"` // e.g., [+4, 2] for +4 and x2
    Bonus      int    `json:"bonus"`     // 15 for Flip 7
    Total      int    `json:"total"`
}
```

**NumberValue**  
Description: Enum for card numbers (0-12). Higher values have more copies (e.g., 12: 12 copies).  
Go Struct Preview:  
```go
type NumberValue int // 0 to 12
const (
    Zero NumberValue = 0
    // ...
    Twelve NumberValue = 12
)
```

**CardType, ModifierType, ActionType**  
Description: Enums for card categorization.  
Go Struct Preview:  
```go
type CardType string // "number", "modifier", "action"
type ModifierType string // "add_2", "multiply_2", etc.
type ActionType string // "freeze", "flip_three", "second_chance"
```

**HandStatus**  
Description: Player hand state.  
Go Struct Preview:  
```go
type HandStatus string // "active", "stayed", "busted", "frozen"
```

**TurnChoice**  
Description: Player decision.  
Go Struct Preview:  
```go
type TurnChoice string // "hit", "stay"
```

**RoundEndReason**  
Description: Why round ended.  
Go Struct Preview:  
```go
type RoundEndReason string // "no_active_players", "flip7_achieved"
```

### Entities

Entities have identity (UUID) and mutable state. They track lifecycle changes.

**Player**  
Description: A game participant with score history. Tracks hands across rounds for counting analysis.  
Key Methods: `StartNewRound()`, `BankScore(score int)`, `IsWinner() bool`  
Fields (Go Struct Preview):  
```go
type Player struct {
    ID         uuid.UUID `json:"id"`
    Name       string    `json:"name"`
    TotalScore int       `json:"total_score"`
    Hands      []PlayerHand // History for strategy
    CurrentHand *PlayerHand
}
```

**PlayerHand**  
Description: A player's cards and status in one round. Enforces uniqueness (bust on duplicates).  
Key Methods: `AddCard(card Card) (busted bool, flip7 bool)`, `CalculateScore() PointValue`, `ResolveAction(card Card)`  
Fields (Go Struct Preview):  
```go
type PlayerHand struct {
    ID                uuid.UUID `json:"id"`
    NumberCards       map[NumberValue]struct{} // Set for uniqueness
    ModifierCards     []Card `json:"modifier_cards"`
    ActionCards       []Card `json:"action_cards"`
    SecondChanceUsed  bool `json:"second_chance_used"`
    Status            HandStatus `json:"status"`
}
```

**Round**  
Description: Manages one game round: Dealing, turns, end conditions.  
Key Methods: `DealInitialCards()`, `PlayerTurn(player *Player, choice TurnChoice)`, `CheckEndConditions() bool`, `ResolveAction(card Card, target *Player)`  
Fields (Go Struct Preview):  
```go
type Round struct {
    ID             uuid.UUID `json:"id"`
    Dealer         *Player `json:"dealer"`
    Players        []*Player `json:"players"`
    Deck           *Deck `json:"deck"`
    DiscardPile    []Card `json:"discard_pile"` // For reshuffle
    ActivePlayers  []*Player `json:"active_players"`
    IsEnded        bool `json:"is_ended"`
    EndReason      RoundEndReason `json:"end_reason"`
}
```

**Game**  
Description: Aggregate root for the full game. Oversees rounds until 200 points.  
Key Methods: `StartRound(dealer *Player)`, `EndRound()`, `DetermineWinner() *Player`  
Fields (Go Struct Preview):  
```go
type Game struct {
    ID            uuid.UUID `json:"id"`
    Players       []*Player `json:"players"`
    CurrentRound  int `json:"current_round"`
    Rounds        []Round `json:"rounds"`
    IsCompleted   bool `json:"is_completed"`
    Winners       []*Player `json:"winners"`
}
```

### Aggregates

Aggregates ensure consistency within boundaries. Invariants are enforced via the root.

**Game Aggregate**  
- Contained Entities/Value Objects: `Players`, `Rounds` (each with `PlayerHand`, `Deck`)  
- Invariants:  
  - Game ends when any player >=200 points.  
  - Dealer rotates left each round.  
  - Scores banked only on Stay/Freeze/Flip7.

**Round Aggregate**  
- Contained Entities/Value Objects: `Players` (with `PlayerHand`), `Deck`, `DiscardPile`  
- Invariants:  
  - Initial deal: Sequential starting from Dealer. Actions resolved immediately.
  - Bust on duplicate Number (unless SecondChance).  
  - Ends on no active players or Flip7.  
  - Deck passes left; reshuffle discards if empty (keep player cards).
  - **Flip Three Rules**:
    - Draw 3 cards one by one.
    - If Second Chance drawn: Set aside/use.
    - If Freeze/FlipThree drawn: Resolve AFTER the 3 draws (if not busted).
  - **Second Chance Rules**:
    - If drawn and player already has one: Must give to another active player (Strategy choice).
    - If no eligible target: Discard.

### Domain Services

Stateless services for cross-entity logic, especially counting/strategy.

**ScoreCalculator**  
Description: Computes final scores per hand.  
Interface (Go Preview):  
```go
type ScoreCalculator interface {
    Compute(hand *PlayerHand) PointValue
}
// Impl: Sum numbers, apply modifiers (x2 first), add bonus.
```

**ActionResolver**  
Description: Handles Action cards (e.g., Freeze banks points; FlipThree forces 3 draws).  
Interface (Go Preview):  
```go
type ActionResolver interface {
    Resolve(round *Round, card Card, target *Player) error
}
// E.g., For SecondChance: Track and discard on duplicate.
```

**StrategyAdvisor**  
Description: Analyzes deck for hit risk (counting: remaining duplicates / total cards). Recommends Hit/Stay.  
Interface (Go Preview):  
```go
type StrategyAdvisor interface {
    Recommend(deck *Deck, hand *PlayerHand, playerScore int) TurnChoice
}
// E.g., Risk = sum(remaining[existing] / total); if <0.1 or near Flip7, Hit.
```

StrategyAdvisor should be designed using the Strategy Pattern, which is a pattern of Software Design.

**DeckFactory**  
Description: Builds initial deck (94 cards: Numbers per PDF counts, 2x each Modifier, 3x each Action). Supports multiple decks (>18 players).  
Interface (Go Preview):  
```go
type DeckFactory interface {
    NewDeck(multiDeck bool) *Deck
}
```

### Deck (Value Object with Mutable Counts)

Special value object for the deck, tracking remaining cards for strategy.

Go Struct Preview:  
```go
type Deck struct {
    Cards          []Card `json:"cards"` // Shuffled sequence
    RemainingCounts map[NumberValue]int `json:"remaining_counts"` // For counting
}

func (d *Deck) Draw() (Card, error) {
    // Pop card, update counts if Number.
}

func (d *Deck) EstimateHitRisk(hand *PlayerHand) float64 {
    total := len(d.Cards)
    if total == 0 { return 0 }
    risk := 0.0
    for val := range hand.NumberCards {
        risk += float64(d.RemainingCounts[val]) / float64(total)
    }
    return risk
}

func (d *Deck) Shuffle() { /* random.Shuffle */ }
```

**Initialization**:  
- Numbers: 12:x12, 11:x11, ..., 1:x1, 0:x1.  
- Modifiers: +2/+4/+6/+8/+10/x2: x2 each.  
- Actions: Freeze/FlipThree/SecondChance: x3 each.

## Repositories (Infrastructure Layer)

Interfaces for persistence (e.g., in-memory or DB).

Go Preview:  
```go
type GameRepository interface {
    Save(game *Game) error
    Load(id uuid.UUID) (*Game, error)
    // For strategy: Query past hands for full deck tracking.
}

type PlayerRepository interface {
    FindByName(name string) (*Player, error)
}
```

## Implementation Notes for Go

- **Packages**: `domain/` for entities/services; `infrastructure/` for repos; `application/` for use cases (e.g., `StartGameHandler`).  
- **UUID**: Use `github.com/google/uuid`.  
- **Immutability**: Value objects as structs; entities with mutex for concurrency (multi-player sim).  
- **Events**: Consider domain events (e.g., `PlayerBustedEvent`, `RoundEndedEvent`) for loose coupling.  
- **Testing**: Unit tests for invariants (e.g., bust on duplicate); integration for full rounds.  
- **Extensions**: For 2-player solo mode: Self-challenge to 200 in <5 rounds.

This Markdown model serves as a blueprint. Next steps: Flesh out full Go structs/interfaces in code, or refine specific parts (e.g., Action resolution).