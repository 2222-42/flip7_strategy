# Flip 7 Strategy Simulation

[![Go Version](https://img.shields.io/github/go-mod/go-version/2222-42/flip7_strategy)](https://github.com/2222-42/flip7_strategy)
[![Go Tests](https://github.com/2222-42/flip7_strategy/actions/workflows/test.yml/badge.svg)](https://github.com/2222-42/flip7_strategy/actions/workflows/test.yml)
[![Lint](https://github.com/2222-42/flip7_strategy/actions/workflows/lint.yml/badge.svg)](https://github.com/2222-42/flip7_strategy/actions/workflows/lint.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/2222-42/flip7_strategy)](https://goreportcard.com/report/github.com/2222-42/flip7_strategy)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A robust simulation of the **Flip 7** card game, implemented in Go using **Domain-Driven Design (DDD)** principles and the **Strategy Pattern**.

## Overview

This project simulates the "Flip 7" card game, where players push their luck to accumulate points without busting. It features a complete domain model, rule enforcement (including complex action cards), and an extensible AI system.

> [!NOTE]
> This is a strategy study and simulation. To play the actual game with friends, please support the creators by purchasing the physical copy: [Buy Flip 7](https://theop.games/products/flip-7)
>
> **Flip 7: With a Vengeance** is a separate ruleset (108-card take-that edition). Domain model and rules live under [`docs/vengeance/`](docs/vengeance/domain_model.md). Evaluation: [`docs/vengeance/strategy_evaluation.md`](docs/vengeance/strategy_evaluation.md) ([Epic #92](https://github.com/2222-42/flip7_strategy/issues/92)).
>
> ```bash
> go run cmd/vengeance/main.go
> go run cmd/evaluate_vengeance/main.go -n 200
> go run cmd/evaluate_vengeance/main.go -n 200 -brutal
> ```
>
> Brutal Mode is opt-in (prompt in `cmd/vengeance`, or `-brutal` for evaluation). Standard scoring stays the default.
>
> Manual mode (choice `2`) is a helper for a physical Vengeance game: type the cards as they appear and Adaptive suggests Hit/Stay plus action/modifier targets. Stay banks only at round end. Input codes: `1-13`, `0`/`Z`, `U` (Unlucky 7), `L`, `-2`…`-10`, `/2`, `J`, `F4`, `SW`, `ST`, `DI`, `S`. Undo/redo: `UNDO`/`<` and `REDO`/`R`/`>` (`U` is the Unlucky 7 card). Nested action input undoes the whole action.
>
> Official rules: [Ruleset Edition 1 PDF](https://cdn.shopify.com/s/files/1/0611/3958/3198/files/26_FLIP_7_VENGEANCE_RULES_C.pdf?v=1770853609). The original Flip 7 simulation is unchanged.

## Features

- **Domain-Driven Design**: Clean separation of concerns with `domain`, `application`, and `infrastructure` layers.
- **Strategy Pattern**: Pluggable AI strategies.
    - **Cautious**: Plays safely, banking points early.
    - **Aggressive**: Pushes for high scores and "Flip 7" bonuses.
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
8. Manual Mode (Real Game Helper)
```

### Modes Explained

- **Automatic Play**: Runs a single game with verbose logging. Great for understanding the game flow and debugging.
- **Participating**: You take the seat of the third player. Follow the prompts to `hit`, `stay`, or choose targets for action cards.
    - **Save/Resume**: A "Save Code" is displayed at the start of each turn. Copy this code. To resume later, select "Participating" mode and paste the code when prompted.
- **Counting**: Runs 1,000 silent games and outputs the win statistics. Use this to see which strategy is currently the strongest.
- **Optimize Heuristic Strategy**: Finds the optimal stopping threshold for the Heuristic strategy.
- **Single Player Optimization**: Calculates average and median rounds to reach 200 points for each strategy.
- **Multiplayer Evaluation**: Simulates games with 1 to 5 players to evaluate strategy performance in different group sizes.
- **Strategy Combination Evaluation**: Runs 1vs1 matchups between all unique pairs of strategies.
- **Manual Mode**: A helper for playing a physical game.
    - **Logging**: Game events are automatically logged to `game_logs.csv` for analysis.
    - **Resume**: Supports saving and resuming game state via a "Save Code".

### Log Analysis
To analyze the logs generated by Manual Mode, run the evaluation tool:
```bash
go run cmd/evaluate_logs/main.go game_logs.csv
```
This tool outputs statistics such as total games played, bust rates, and win counts.

## Documentation

- [Strategy Evaluation Results](docs/strategy_evaluation.md): Original Flip 7 — single-player speed and multiplayer win rates.
- [Vengeance Strategy Evaluation](docs/vengeance/strategy_evaluation.md): Flip 7: With a Vengeance — same axes plus heuristic stop and Flip Four targeting.

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
- **Turn Order**: Dealer goes first. The Dealer role rotates to the next player (round-robin) after each round.
- **Busting**: Drawing a duplicate number card ends your turn (unless you have a Second Chance).
- **Flip 7**: Collecting 7 cards grants a 15-point bonus (or more depending on house rules implemented).
- **Action Cards**:
    - **Freeze**: Target banks points immediately and stops drawing.
    - **Flip Three**: Target is forced to draw 3 cards. Nested actions (Freeze/Flip Three) are queued and resolved *after* the draws.
    - **Second Chance**: Saves you from a bust. If you draw a duplicate Second Chance, you must pass it to another player.

