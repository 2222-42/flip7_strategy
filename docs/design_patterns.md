# Design Patterns in Flip 7 Strategy

This document outlines the software design patterns implemented in the `flip7_strategy` repository. These patterns help organize the code, make it more maintainable, and allow for flexible extension of game logic.

## 1. Strategy Pattern

The **Strategy Pattern** is the core of this project, as the name suggests. It allows the game engine to support multiple AI behaviors (strategies) interchangeably.

### Definition
The Strategy pattern defines a family of algorithms, encapsulates each one, and makes them interchangeable. It lets the algorithm vary independently from clients that use it.

### Implementation
- **Interface**: `Strategy` (defined in `internal/domain/strategy.go`)
    - `Decide(deck *Deck, hand *PlayerHand, playerScore int, otherPlayers []*Player) TurnChoice`: Determines whether to Hit or Stay.
    - `ChooseTarget(action ActionType, candidates []*Player, self *Player) *Player`: Selects a target for action cards (Freeze, Flip Three, Second Chance).
    - `Name() string`: Returns the name of the strategy.

- **Context**: `Player` (in `internal/domain/player.go`) holds a reference to a `Strategy` and delegates decision-making to it. `GameService` calls these methods during the game loop.

### Strategies
The following strategies implement the `Strategy` interface:

| Strategy | Description |
| :--- | :--- |
| **Cautious** | Stays if the risk of busting is even slightly elevated (e.g., > 10%). Prioritizes safety. |
| **Aggressive** | Pushes luck until the risk is high (> 30%). Often targets random opponents. |
| **Probabilistic** | Uses a simplified expected value calculation to decide. Adjusts risk tolerance based on score difference. |
| **Heuristic** | Stops when the sum of number cards in hand reaches a specific threshold (default 27). |
| **Expected Value** | Calculates the mathematical expected value of drawing the next card based on the remaining deck composition. |
| **Adaptive** | Switches between other strategies (e.g., Expected Value vs. Aggressive) based on the game state (winning vs. losing). |
| **Human** | Allows a human user to input decisions via the console. |

## 2. Factory Pattern

The **Factory Pattern** is used to create objects without specifying the exact class of object that will be created, or to simplify complex object initialization.

### Implementation
- **Strategy Factories**: Functions like `NewCautiousStrategy()`, `NewAggressiveStrategy()`, `NewHeuristicStrategy(threshold int)` create instances of strategies, often setting up default dependencies (like `TargetSelector`).
- **Game Entities**:
    - `NewGameService(game *domain.Game)`: Creates the service layer.
    - `NewPlayer(name string, strategy Strategy)`: Creates a player with a specific strategy.
    - `NewDeck()`: Creates and shuffles a standard deck.
    - `NewRound(...)`: Initializes a new round with active players and a deck.

## 3. Facade Pattern

The **Facade Pattern** provides a simplified interface to a library, a framework, or any other complex set of classes.

### Implementation
- **GameService** (`internal/application/game_service.go`): Acts as a facade for the game logic. It manages the flow of rounds, turns, card drawing, and rule enforcement. Clients (like the `main` function or simulation tools) interact with `GameService` rather than managing individual `Player`, `Deck`, and `Round` objects directly.
- **SimulationService** (`internal/application/simulation_service.go`): Provides a high-level interface to run thousands of games and collect statistics, abstracting away the details of running individual game loops.

## 4. Component Pattern (Target Selector)

To separate the "Decide Hit/Stay" logic from the "Choose Target" logic, we use a component-style composition.

### Implementation
- **TargetSelector Interface**: Defines `ChooseTarget(action, candidates, self)`.
- **Composition**: Strategies like `AggressiveStrategy` and `ProbabilisticStrategy` embed a `TargetSelector` (often via `CommonTargetChooser`). This allows, for example, an Aggressive Strategy to switch from "Random Targeting" to "Leader Targeting" without changing the core Hit/Stay logic, or vice-versa.
- **Implementations**:
    - `DefaultTargetSelector`: Hits leaders or statistically high-risk players.
    - `RandomTargetSelector`: Chooses targets randomly (chaos).
    - `RiskBasedTargetSelector`: Configurable risk thresholds for decision making.

## 5. State Pattern (Implicit & Explicit)

The **State Pattern** concepts are used to manage the lifecycle of a player's hand and strategy adaptability.

### Implementation
- **HandStatus**: The `PlayerHand` struct tracks its state via `Status` (`active`, `stayed`, `busted`, `frozen`). The game logic transitions the hand between these states based on events.
- **Adaptive Strategy**: `AdaptiveStrategy` explicitly acts as a state machine for AI behavior. It holds instances of other strategies (e.g., `Aggressive` and `ExpectedValue`) and delegates to one of them depending on the global game state (e.g., if an opponent is close to winning).

## 6. Observer / Logger

To decouple game logic from reporting, we use an observer-like pattern for logging.

### Implementation
- **GameLogger Interface**: Defines a contract for recording events (`Log(...)`).
- **Implementations**:
    - `CSVLogger`: Writes events to a structured CSV file for analysis.
    - `ConsoleLogger` (Implicit): `GameService` prints to stdout if not silent.
- **Usage**: `ManualGameService` logs events without knowing the details of the storage mechanism.

## 7. Dependency Injection

**Dependency Injection** is used to decouple components.

### Implementation
- **Strategies**: Strategies are injected into `Player` objects at creation time.
- **Target Selectors**: `AggressiveStrategy` and others accept a `TargetSelector` interface, allowing for customizable targeting logic.
- **GameService**: Takes a `Game` domain object, separating the state from the logic.
- **Loggers**: `GameLogger` is injected into `ManualGameService`.
