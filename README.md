# Flip 7 Strategy Simulation

A robust simulation of the **Flip 7** card game, implemented in Go using **Domain-Driven Design (DDD)** principles and the **Strategy Pattern**.

## Overview

This project simulates the "Flip 7" card game, where players push their luck to accumulate points without busting. It features a complete domain model, rule enforcement (including complex action cards), and an extensible AI system.

## Features

- **Domain-Driven Design**: Clean separation of concerns with `domain`, `application`, and `infrastructure` layers.
- **Strategy Pattern**: Pluggable AI strategies.
    - **Cautious**: Plays safely, banking points early.
    - **Aggressive**: Pushes for high scores and "Flip 7" bonuses.
    - **Probabilistic**: Calculates risk based on remaining cards in the deck.
    - **Probabilistic**: Calculates risk based on remaining cards in the deck.
    - **ExpectedValue**: Calculates the expected value of the next draw based on the remaining deck.
    - **Adaptive**: Switches between ExpectedValue and Aggressive based on opponent scores.
    - **Human**: Interactive CLI mode for you to play.
- **Complex Game Rules**:
    - **Actions**: Freeze, Flip Three (with nested resolution), Second Chance (with passing logic).
    - **Bonuses**: Flip 7 (collecting 7 cards) awards extra points.
    - **Scoring**: Multipliers and modifiers are applied correctly.
- **Game Modes**:
    1. **Automatic**: Watch AI agents battle it out.
    2. **Interactive**: Play against the AI.
    3. **Counting (Monte Carlo)**: Run thousands of simulations to analyze strategy win rates.

## Getting Started

### Prerequisites
- Go 1.23 or higher

### Running the Game
To start the application, run the main entry point:

```bash
go run cmd/flip7/main.go
```

You will be presented with a menu to select the game mode:

```text
Welcome to Flip 7 Strategy!
Select Mode:
1. Automatic Play (Sample Game)
2. Participating (Interactive)
4. Optimize Heuristic Strategy
5. Single Player Optimization (Fastest to 200)
6. Multiplayer Evaluation (1-5 Players)
7. Strategy Combination Evaluation (1vs1)
```

### Modes Explained

- **Automatic Play**: Runs a single game with verbose logging. Great for understanding the game flow and debugging.
- **Participating**: You take the seat of the third player. Follow the prompts to `hit`, `stay`, or choose targets for action cards.
- **Counting**: Runs 1,000 silent games and outputs the win statistics. Use this to see which strategy is currently the strongest.
- **Optimize Heuristic Strategy**: Finds the optimal stopping threshold for the Heuristic strategy.
- **Single Player Optimization**: Calculates average and median rounds to reach 200 points for each strategy.
- **Multiplayer Evaluation**: Simulates games with 1 to 5 players to evaluate strategy performance in different group sizes.
- **Strategy Combination Evaluation**: Runs 1vs1 matchups between all unique pairs of strategies.

## Documentation

- [Strategy Evaluation Results](docs/strategy_evaluation.md): Detailed analysis of strategy performance, including single-player speed and multiplayer win rates.

## Project Structure

```
flip7_strategy/
├── cmd/
│   └── flip7/          # Main entry point
├── docs/               # Domain documentation
├── internal/
│   ├── application/    # Game orchestration (GameService, SimulationService)
│   ├── domain/         # Core business logic (Entities, Value Objects)
│   │   └── strategy/   # AI implementations
│   └── infrastructure/ # Console I/O
└── README.md
```

## Rules Implemented
- **Turn Order**: Dealer goes first.
- **Busting**: Drawing a duplicate number card ends your turn (unless you have a Second Chance).
- **Flip 7**: Collecting 7 cards grants a 15-point bonus (or more depending on house rules implemented).
- **Action Cards**:
    - **Freeze**: Target banks points immediately and stops drawing.
    - **Flip Three**: Target is forced to draw 3 cards. Nested actions (Freeze/Flip Three) are queued and resolved *after* the draws.
    - **Second Chance**: Saves you from a bust. If you draw a duplicate Second Chance, you must pass it to another player.

