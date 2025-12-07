package application

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"time"

	"flip7_strategy/internal/domain"
	"flip7_strategy/internal/domain/logger"
	"flip7_strategy/internal/domain/strategy"
)

// GameMemento represents a snapshot of the game state.
type GameMemento string

// GameHistory manages the history of game states for undo/redo.
type GameHistory struct {
	mementos     []GameMemento
	currentIndex int
}

// Push adds a new memento to the history, truncating any future redo states.
func (h *GameHistory) Push(memento GameMemento) {
	// If we are in the middle of the history (after undo), remove future states
	if h.currentIndex < len(h.mementos)-1 {
		h.mementos = h.mementos[:h.currentIndex+1]
	}
	h.mementos = append(h.mementos, memento)
	h.currentIndex = len(h.mementos) - 1
}

// Current returns the current memento.
func (h *GameHistory) Current() (GameMemento, bool) {
	if h.currentIndex >= 0 && h.currentIndex < len(h.mementos) {
		return h.mementos[h.currentIndex], true
	}
	return "", false
}

// Undo moves the pointer back and returns the previous memento.
func (h *GameHistory) Undo() (GameMemento, bool) {
	if h.currentIndex > 0 {
		h.currentIndex--
		return h.mementos[h.currentIndex], true
	}
	return "", false
}

// Redo moves the pointer forward and returns the next memento.
func (h *GameHistory) Redo() (GameMemento, bool) {
	if h.currentIndex < len(h.mementos)-1 {
		h.currentIndex++
		return h.mementos[h.currentIndex], true
	}
	return "", false
}

// gameStateWrapper wraps the game state with metadata for serialization.
type gameStateWrapper struct {
	Game              *domain.Game `json:"game"`
	UserControlledIDs []string     `json:"user_controlled_ids"` // IDs of players with nil strategy
	GameID            string       `json:"game_id"`             // GameID for logging continuity
}

// ManualGameService handles the manual mode where the user inputs game events.
type ManualGameService struct {
	Game                *domain.Game
	Reader              *bufio.Reader
	Logger              logger.GameLogger
	GameID              string
	secondChanceHandler *domain.SecondChanceHandler
	History             GameHistory
}

// manualFlipThreeCardSource implements FlipThreeCardSource for manual mode.
type manualFlipThreeCardSource struct {
	service *ManualGameService
}

func (ms *manualFlipThreeCardSource) GetNextCard(cardNum int, target *domain.Player) (domain.Card, error) {
	// Keep retrying until valid card is entered
	for {
		fmt.Printf("Input card %d/3 for %s: ", cardNum, target.Name)
		input, _ := ms.service.Reader.ReadString('\n')
		input = strings.TrimSpace(input)

		card, err := ms.service.parseInput(input)
		if err != nil {
			fmt.Printf("Invalid input: %v. Try again.\n", err)
			continue // Retry
		}

		if err := ms.service.removeCardFromDeck(card); err != nil {
			fmt.Printf("Error: %v. Try again.\n", err)
			continue // Retry
		}

		return card, nil
	}
}

// manualFlipThreeCardProcessor implements FlipThreeCardProcessor for manual mode.
type manualFlipThreeCardProcessor struct {
	service *ManualGameService
}

func (mp *manualFlipThreeCardProcessor) ProcessImmediateCard(target *domain.Player, card domain.Card) error {
	mp.service.processCard(target, card)
	return nil
}

func (mp *manualFlipThreeCardProcessor) ProcessQueuedAction(target *domain.Player, card domain.Card) error {
	mp.service.processCard(target, card)
	return nil
}

// SelectTarget implements domain.TargetSelector interface for manual mode.
func (s *ManualGameService) SelectTarget(actionType domain.ActionType, candidates []*domain.Player, actor *domain.Player) *domain.Player {
	return s.promptForTarget(actionType, candidates, actor)
}

// NewManualGameService creates a new ManualGameService.
func NewManualGameService(reader *bufio.Reader, logger logger.GameLogger) *ManualGameService {
	return &ManualGameService{
		Reader:              reader,
		Logger:              logger,
		GameID:              fmt.Sprintf("game_%d", time.Now().Unix()),
		secondChanceHandler: domain.NewSecondChanceHandler(),
	}
}

// Run starts the manual game loop.
func (s *ManualGameService) Run() {
	fmt.Println("\n--- Manual Mode ---")
	s.setupPlayers()
	s.PushState() // Push initial state
	s.gameLoop()
}

func (s *ManualGameService) setupPlayers() {
	fmt.Println("Do you want to resume a game? (Enter save code, file path, or press Enter to start new)")
	fmt.Print("Save Code / File Path: ")
	input, _ := s.Reader.ReadString('\n')
	input = strings.TrimSpace(input)

	saveCode := input
	// Check if input is a file (and not just a short string that happens to match a filename, though unlikely for a JWT)
	// We check if the file exists and is readable.
	if len(input) > 0 && len(input) < 255 { // Basic length check to avoid stat-ing huge tokens
		if _, err := os.Stat(input); err == nil {
			content, err := os.ReadFile(input)
			if err == nil {
				saveCode = strings.TrimSpace(string(content))
				fmt.Printf("Read save code from file: %s\n", input)
			}
		}
	}

	if saveCode != "" {
		if err := s.LoadState(saveCode); err != nil {
			fmt.Printf("Failed to load game: %v. Starting new game.\n", err)
		} else {
			fmt.Println("Game resumed successfully!")
			return
		}
	}

	fmt.Print("Enter number of players: ")
	numPlayersStr, err := s.Reader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading input. Defaulting to 2 players.")
		numPlayersStr = "2"
	}
	numPlayers, err := strconv.Atoi(strings.TrimSpace(numPlayersStr))
	if err != nil || numPlayers < 1 {
		fmt.Println("Invalid number of players. Defaulting to 2.")
		numPlayers = 2
	}

	var players []*domain.Player

	// Setup "Me" (User)
	me := domain.NewPlayer("Me", nil) // Strategy is nil because user controls it
	players = append(players, me)

	// Setup other players
	for i := 1; i < numPlayers; i++ {
		fmt.Printf("Enter name for Player %d: ", i+1)
		name, err := s.Reader.ReadString('\n')
		if err != nil {
			name = ""
		}
		name = strings.TrimSpace(name)
		if name == "" {
			name = fmt.Sprintf("Player %d", i+1)
		}
		// Assign a default strategy for others just to satisfy the struct, though we won't use it for decision making in manual mode
		// actually, we might want to use it for "Best Choice" suggestions if we were simulating them, but here we just track state.
		// Let's use ProbabilisticStrategy as a placeholder.
		players = append(players, domain.NewPlayer(name, &strategy.ProbabilisticStrategy{}))
	}

	// Set start player
	fmt.Println("Select start player:")
	for i, p := range players {
		fmt.Printf("%d. %s\n", i+1, p.Name)
	}
	fmt.Print("Enter choice: ")
	startIdxStr, err := s.Reader.ReadString('\n')
	if err != nil {
		startIdxStr = "1"
	}
	startIdx, err := strconv.Atoi(strings.TrimSpace(startIdxStr))
	if err != nil || startIdx < 1 || startIdx > numPlayers {
		fmt.Println("Invalid choice. Defaulting to Me.")
		startIdx = 1
	}

	s.Game = domain.NewGame(players)
	s.Game.DealerIndex = startIdx - 1 // Set initial dealer index

	if s.Logger != nil {
		s.Logger.Log(s.GameID, "0", "system", "GameStart", map[string]interface{}{
			"num_players": len(players),
			"players":     getPlayerNames(players),
		})
	}

	fmt.Println("Game started!")
}

func getPlayerNames(players []*domain.Player) []string {
	names := make([]string, len(players))
	for i, p := range players {
		names[i] = p.Name
	}
	return names
}

func (s *ManualGameService) gameLoop() {
	for !s.Game.IsCompleted {
		s.Game.RoundCount++
		s.playRound()
		// Collect cards from players' hands to discard pile at end of round
		if s.Game.CurrentRound != nil { // Could be nil on first iteration or error
			for _, p := range s.Game.Players {
				// Number cards
				for _, val := range p.CurrentHand.RawNumberCards {
					s.Game.DiscardPile = append(s.Game.DiscardPile, domain.Card{Type: domain.CardTypeNumber, Value: val})
				}
				// Modifier cards
				s.Game.DiscardPile = append(s.Game.DiscardPile, p.CurrentHand.ModifierCards...)
				// Action cards
				s.Game.DiscardPile = append(s.Game.DiscardPile, p.CurrentHand.ActionCards...)

				// Clear the hand after collecting to prevent double counting or memory issues
				p.CurrentHand.RawNumberCards = nil
				p.CurrentHand.ModifierCards = nil
				p.CurrentHand.ActionCards = nil
			}
		}

		// Rotate dealer for next round
		s.Game.DealerIndex = (s.Game.DealerIndex + 1) % len(s.Game.Players)

		// Check for winners
		winners := s.Game.DetermineWinners()
		if len(winners) > 0 {
			s.Game.Winners = winners
			s.Game.IsCompleted = true
		}
		// Update deck reference for the next round
		// If a reshuffle happened during PlayRound, s.Game.CurrentRound.Deck points to the new deck.
		if s.Game.CurrentRound != nil && s.Game.CurrentRound.Deck != nil {
			s.Game.Deck = s.Game.CurrentRound.Deck
		}
	}
	s.printWinner()

	if s.Logger != nil {
		s.Logger.Log(s.GameID, strconv.Itoa(s.Game.RoundCount), "system", "GameEnd", map[string]interface{}{
			"winners": getPlayerNames(s.Game.Winners),
		})
	}
}

func (s *ManualGameService) playRound() {
	// Initialize new round only if not resuming
	if s.Game.CurrentRound == nil || s.Game.CurrentRound.IsEnded {
		if s.Game.Deck == nil {
			s.Game.Deck = domain.NewDeck()
		}
		dealer := s.Game.Players[s.Game.DealerIndex]
		s.Game.CurrentRound = domain.NewRound(s.Game.Players, dealer, s.Game.Deck)
		fmt.Printf("\n--- New Round! Dealer: %s ---\n", dealer.Name)

		if s.Logger != nil {
			s.Logger.Log(s.GameID, strconv.Itoa(s.Game.RoundCount), "system", "RoundStart", map[string]interface{}{
				"dealer": dealer.Name,
			})
		}
		// Push state at start of new round (stable point)
		s.PushState()
	} else {
		fmt.Println("Resuming round...")
	}

	for !s.Game.CurrentRound.IsEnded {
		// Label for restarting turn loop if undo/redo happens
	StartOfTurn:
		if len(s.Game.CurrentRound.ActivePlayers) == 0 {
			s.Game.CurrentRound.End(domain.RoundEndReasonNoActivePlayers)
			break
		}

		// Robustness check: Ensure there is at least one player with Active status
		hasActive := false
		for _, p := range s.Game.CurrentRound.ActivePlayers {
			if p.CurrentHand.Status == domain.HandStatusActive {
				hasActive = true
				break
			}
		}
		if !hasActive {
			s.Game.CurrentRound.End(domain.RoundEndReasonNoActivePlayers)
			break
		}

		// Ensure index is valid
		if s.Game.CurrentRound.CurrentTurnIndex >= len(s.Game.CurrentRound.ActivePlayers) {
			s.Game.CurrentRound.CurrentTurnIndex = 0
		}

		currentPlayer := s.Game.CurrentRound.ActivePlayers[s.Game.CurrentRound.CurrentTurnIndex]

		// Skip players who are not active (busted, stayed, frozen)
		if currentPlayer.CurrentHand.Status != domain.HandStatusActive {
			// Fix: Remove inactive player from ActivePlayers list to prevent infinite loops
			// and ensure the round can end (RoundEndReasonNoActivePlayers).
			s.Game.CurrentRound.RemoveActivePlayer(currentPlayer)
			continue
		}

		// Show Save Code
		code, err := s.SaveState()
		if err == nil {
			fmt.Printf("\n[Save Code]: %s\n", code)
		}

		fmt.Printf("\n>>> Turn: %s (Score: %d)\n", currentPlayer.Name, currentPlayer.TotalScore)

		calc := domain.NewScoreCalculator()
		score := calc.Compute(currentPlayer.CurrentHand)

		if s.Logger != nil {
			s.Logger.Log(s.GameID, strconv.Itoa(s.Game.RoundCount), currentPlayer.ID.String(), "TurnStart", map[string]interface{}{
				"score":      currentPlayer.TotalScore,
				"hand_score": score.Total,
			})
		}

		// Show current hand score before input
		fmt.Printf("Current Hand: %s | Score: %d\n", s.formatHand(currentPlayer.CurrentHand), score.Total)

		s.analyzeState(currentPlayer)

		// Input loop for this turn (single action)
		turnEnded := false
		playerRemoved := false
		shouldRestartTurn := false

		for !turnEnded {
			fmt.Print("Input (0-12, +N, x2, F, T, C, S, U/UNDO, R/REDO): ")
			input, err := s.Reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading input. Exiting game.")
				s.Game.IsCompleted = true
				return
			}
			input = strings.TrimSpace(input)

			// Check for Undo/Redo
			if strings.EqualFold(input, "U") || strings.EqualFold(input, "UNDO") || input == "<" {
				s.Undo()
				shouldRestartTurn = true
				break
			}
			if strings.EqualFold(input, "R") || strings.EqualFold(input, "REDO") || input == ">" {
				s.Redo()
				shouldRestartTurn = true
				break
			}

			if strings.EqualFold(input, "S") {
				// Validation: Cannot stay on first turn (empty hand) unless special conditions met
				if !currentPlayer.CurrentHand.CanStay() {
					fmt.Println("Invalid move: You must hit on your first turn (unless you have points or specific actions)!")
					continue
				}

				currentPlayer.CurrentHand.Status = domain.HandStatusStayed
				score := currentPlayer.BankCurrentHand()
				fmt.Printf("%s banked %d points! Total: %d\n", currentPlayer.Name, score, currentPlayer.TotalScore)

				if s.Logger != nil {
					s.Logger.Log(s.GameID, strconv.Itoa(s.Game.RoundCount), currentPlayer.ID.String(), "Stay", map[string]interface{}{
						"banked_score": score,
						"total_score":  currentPlayer.TotalScore,
					})
				}

				s.Game.CurrentRound.RemoveActivePlayer(currentPlayer)
				playerRemoved = true
				turnEnded = true
			} else {
				// Parse card or action
				card, err := s.parseInput(input)
				if err != nil {
					fmt.Printf("Invalid input: %v. Try again.\n", err)
					continue
				}

				// Remove card from deck (tracking)
				if err := s.removeCardFromDeck(card); err != nil {
					fmt.Printf("Error: %v. Try again.\n", err)
					continue
				}

				// Process card
				s.processCard(currentPlayer, card)

				// Check if player was removed (Freeze only)
				// processCard calls RemoveActivePlayer only for Freeze actions.
				// Note: Flip7 ends the round without removing the player.
				// Note: Busted players are NOT removed from ActivePlayers; only their status changes.
				// We need to check if currentPlayer is still in ActivePlayers.
				isActive := false
				for _, p := range s.Game.CurrentRound.ActivePlayers {
					if p.ID == currentPlayer.ID {
						isActive = true
						break
					}
				}
				if !isActive {
					playerRemoved = true
				}

				// Turn always ends after one action (Hit) or Action card
				turnEnded = true
			}
		}

		if shouldRestartTurn {
			goto StartOfTurn
		}

		// Check if round ended during this loop (Flip 7 or all stayed)
		if s.Game.CurrentRound.IsEnded {
			// Do NOT push state here to avoid "loop of death" on undo.
			// Let gameLoop transition to new round and push there.
			break
		}

		// Update Turn Index
		if !playerRemoved {
			s.Game.CurrentRound.CurrentTurnIndex++
		}
		// If player removed (Freeze action), the next player slides into the current index, so we don't increment.
		// Busted players remain in ActivePlayers but are skipped via the status check at the start of the loop.

		// Push state if action successful and round not ended
		// We push AFTER updating the turn index so that the saved state points to the NEXT player's turn.
		// This ensures that when we Undo, we return to the start of the turn that was just completed (or rather,
		// we return to the state where the previous player has finished, and it is now the current player's turn).
		s.PushState()
	}
}

func (s *ManualGameService) analyzeState(p *domain.Player) {
	// Show bust rate
	risk := s.Game.CurrentRound.Deck.EstimateHitRisk(p.CurrentHand.NumberCards, p.CurrentHand.HasSecondChance())
	fmt.Printf("Bust Rate: %.2f%%\n", risk*100)

	// Suggest best choice
	adaptive := strategy.NewAdaptiveStrategy()
	choice := adaptive.Decide(s.Game.CurrentRound.Deck, p.CurrentHand, p.TotalScore, s.getOpponents(p))
	fmt.Printf("Suggested Move: %s\n", choice)
}

func (s *ManualGameService) getOpponents(p *domain.Player) []*domain.Player {
	var opponents []*domain.Player
	for _, other := range s.Game.Players {
		if other.ID != p.ID {
			opponents = append(opponents, other)
		}
	}
	return opponents
}

func (s *ManualGameService) parseInput(input string) (domain.Card, error) {
	input = strings.ToUpper(input)

	// Modifiers
	switch input {
	case "+2":
		return domain.Card{Type: domain.CardTypeModifier, ModifierType: domain.ModifierPlus2}, nil
	case "+4":
		return domain.Card{Type: domain.CardTypeModifier, ModifierType: domain.ModifierPlus4}, nil
	case "+6":
		return domain.Card{Type: domain.CardTypeModifier, ModifierType: domain.ModifierPlus6}, nil
	case "+8":
		return domain.Card{Type: domain.CardTypeModifier, ModifierType: domain.ModifierPlus8}, nil
	case "+10":
		return domain.Card{Type: domain.CardTypeModifier, ModifierType: domain.ModifierPlus10}, nil
	case "X2", "*2":
		return domain.Card{Type: domain.CardTypeModifier, ModifierType: domain.ModifierX2}, nil
	}

	// Actions
	switch input {
	case "F":
		return domain.Card{Type: domain.CardTypeAction, ActionType: domain.ActionFreeze}, nil
	case "T":
		return domain.Card{Type: domain.CardTypeAction, ActionType: domain.ActionFlipThree}, nil
	case "C":
		return domain.Card{Type: domain.CardTypeAction, ActionType: domain.ActionSecondChance}, nil
	}

	// Numbers
	if val, err := strconv.Atoi(input); err == nil {
		if val >= 0 && val <= 12 {
			return domain.Card{Type: domain.CardTypeNumber, Value: domain.NumberValue(val)}, nil
		}
		return domain.Card{}, fmt.Errorf("number out of range")
	}

	return domain.Card{}, fmt.Errorf("unknown input")
}

func (s *ManualGameService) removeCardFromDeck(card domain.Card) error {
	// Check if deck is active
	if s.Game.CurrentRound == nil || s.Game.CurrentRound.Deck == nil {
		return fmt.Errorf("no active round/deck")
	}

	deck := s.Game.CurrentRound.Deck

	// Helper to find and remove card
	findAndRemove := func(d *domain.Deck, target domain.Card) bool {
		for i, c := range d.Cards {
			match := false
			if c.Type == target.Type {
				if c.Type == domain.CardTypeNumber && c.Value == target.Value {
					match = true
				} else if c.Type == domain.CardTypeModifier && c.ModifierType == target.ModifierType {
					match = true
				} else if c.Type == domain.CardTypeAction && c.ActionType == target.ActionType {
					match = true
				}
			}

			if match {
				// Remove it
				d.Cards = append(d.Cards[:i], d.Cards[i+1:]...)
				// Update counts if number
				if target.Type == domain.CardTypeNumber {
					d.RemainingCounts[target.Value]--
				}
				return true
			}
		}
		return false
	}

	// Try removing from current deck
	if findAndRemove(deck, card) {
		return nil
	}

	// If not found, it might be because the deck is empty (or the card is simply not there).
	// We should try to replenish the deck from the discard pile IF the deck is effectively empty
	// relative to the user's perception (i.e., they are trying to draw a card that should be there).
	// However, we don't know "what" they are trying to draw until they input it.
	// If the card IS valid but just not in the current small deck remnant, we reshuffle.

	if len(s.Game.DiscardPile) > 0 {
		fmt.Printf("Card not found in current deck. Attempting to reshuffle %d cards from discard pile...\n", len(s.Game.DiscardPile))
		if s.Logger != nil {
			s.Logger.Log(s.GameID, strconv.Itoa(s.Game.RoundCount), "system", "Reshuffle", map[string]interface{}{
				"discard_count": len(s.Game.DiscardPile),
			})
		}

		// Create new deck from discards
		newDeck := domain.NewDeckFromCards(s.Game.DiscardPile)
		// Append existing deck cards to new deck (in case there were a few left)
		if len(deck.Cards) > 0 {
			newDeck.Cards = append(newDeck.Cards, deck.Cards...)
			// Re-calculate counts or merge counts (simpler to just rebuild counts from all cards)
			// But NewDeckFromCards already builds counts from input. We need to add existing.
			for _, c := range deck.Cards {
				if c.Type == domain.CardTypeNumber {
					newDeck.RemainingCounts[c.Value]++
				}
			}
			// Only shuffle again if we added existing cards, otherwise NewDeckFromCards shuffle is sufficient.
			// Actually, if we append cards to the end, we SHOULD shuffle everything again to mix them.
			// Wait, the PR comment said: "NewDeckFromCards already shuffles... but then you only shuffle again if there were remaining cards... This creates inconsistent behavior".
			// Suggestion: Move shuffle outside.
		}
		newDeck.Shuffle()

		// Update references
		s.Game.CurrentRound.Deck = newDeck
		s.Game.Deck = newDeck
		s.Game.DiscardPile = []domain.Card{} // Clear discard pile

		// Try removing again from the new deck
		if findAndRemove(s.Game.CurrentRound.Deck, card) {
			return nil
		}
	}

	return fmt.Errorf("card not found in deck (already drawn?)")
}

// processCard handles the logic of adding a card to a player's hand and resolving its effects.
//
// Action Card Processing Order (based on domain model - docs/domain_model.md):
//
//   - Flip Three & Freeze: The player who DRAWS the card chooses a target player.
//     The action effect is applied to the TARGET player, then the card is added to the DRAWER's hand.
//     This order is important because:
//     1. The drawer must choose a target before knowing the full outcome
//     2. The target processes the effect (e.g., draws 3 cards for Flip Three)
//     3. Only after resolution does the drawer add the action card to their own hand
//
//   - Second Chance: Added to drawer's hand immediately. Per domain model (lines 173-175),
//     if the drawer already has one, they must pass it to another active player (or discard
//     if no valid target).
//     NOTE: In Manual Mode, this passing logic is NOT automated. When a player draws a Second
//     Chance and already has one, the user should track this situation and manually decide which
//     active player receives it, then input the Second Chance card during that target player's
//     turn instead of the original drawer's turn.
//
//   - During Flip Three: Per domain model (lines 169-172), if the target draws another Flip Three
//     or Freeze while drawing their 3 cards, those action cards are queued and resolved AFTER all
//     3 cards are drawn (if the target hasn't busted). See resolveFlipThreeManual for details.
//
// Number/Modifier Cards: Added to the player's hand immediately, checked for bust/flip7.
func (s *ManualGameService) processCard(p *domain.Player, card domain.Card) {
	fmt.Printf("Played: %v\n", card)

	if s.Logger != nil {
		s.Logger.Log(s.GameID, strconv.Itoa(s.Game.RoundCount), p.ID.String(), "CardPlayed", map[string]interface{}{
			"card": card.String(),
		})
	}

	// Special handling for Second Chance BEFORE adding to hand
	if card.Type == domain.CardTypeAction && card.ActionType == domain.ActionSecondChance {
		result := s.secondChanceHandler.HandleSecondChance(p, s.Game.CurrentRound.ActivePlayers, s)

		if result.ShouldDiscard {
			fmt.Println("All other active players already have a Second Chance. Discarding card.")
			fmt.Println("(Remove the Second Chance card from play)")
			return
		} else if result.PassToPlayer != nil {
			fmt.Printf("%s already has a Second Chance! Giving it to %s\n", p.Name, result.PassToPlayer.Name)
			fmt.Printf("(Give the Second Chance card to %s)\n", result.PassToPlayer.Name)
			// Add the card to the target player's hand for tracking
			result.PassToPlayer.CurrentHand.ActionCards = append(result.PassToPlayer.CurrentHand.ActionCards, card)
			return
		}
		// Otherwise, fall through to add to player's hand
	}

	// Special handling for Actions (Freeze and Flip Three)
	if card.Type == domain.CardTypeAction {
		if card.ActionType == domain.ActionFlipThree || card.ActionType == domain.ActionFreeze {
			// Step 1: Prompt the drawer (p) to choose a target player for the action
			target := s.promptForTarget(card.ActionType, s.Game.CurrentRound.ActivePlayers, p)
			if target == nil {
				fmt.Println("No target selected (or invalid). Action cancelled (card still played).")
			} else {
				// Step 2: Apply the action effect to the TARGET player
				switch card.ActionType {
				case domain.ActionFreeze:
					fmt.Printf("Freezing %s!\n", target.Name)
					target.CurrentHand.Status = domain.HandStatusFrozen
					score := target.BankCurrentHand()
					fmt.Printf("%s banked %d points! Total: %d\n", target.Name, score, target.TotalScore)
					s.Game.CurrentRound.RemoveActivePlayer(target)
				case domain.ActionFlipThree:
					fmt.Printf("Flip Three on %s! They must draw 3 cards.\n", target.Name)
					s.resolveFlipThreeManual(target)
				}
			}
		}
		// Step 3: Add the action card to the DRAWER's (p) hand after effect resolution
		// Note: Per issue #17, action cards (Flip Three, Freeze) end the turn after resolution.
		p.CurrentHand.AddCard(card)

		// Show current hand score
		calc := domain.NewScoreCalculator()
		score := calc.Compute(p.CurrentHand)
		fmt.Printf("Current Hand: %s | Score: %d\n", s.formatHand(p.CurrentHand), score.Total)

		return
	}

	// Add card to hand logic (for Number and Modifier cards)
	busted, flip7, discarded := p.CurrentHand.AddCard(card)

	// Handle discarded cards (e.g., from Second Chance usage)
	// In manual mode, inform the user to physically remove these cards
	if len(discarded) > 0 {
		fmt.Printf("Second Chance used! Remove %d card(s) from play: ", len(discarded))
		for i, c := range discarded {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Print(c.String())
		}
		fmt.Println()
		// Add to discard pile
		s.Game.DiscardPile = append(s.Game.DiscardPile, discarded...)
	}

	if busted {
		fmt.Println("BUSTED!")
		p.CurrentHand.Status = domain.HandStatusBusted
		s.Game.CurrentRound.RemoveActivePlayer(p)

		if s.Logger != nil {
			s.Logger.Log(s.GameID, strconv.Itoa(s.Game.RoundCount), p.ID.String(), "Bust", map[string]interface{}{
				"hand": s.formatHand(p.CurrentHand),
			})
		}

		return
	} else if flip7 {
		fmt.Println("FLIP 7!")
		p.CurrentHand.Status = domain.HandStatusStayed
		score := p.BankCurrentHand()
		fmt.Printf("%s banked %d points! Total: %d\n", p.Name, score, p.TotalScore)

		// Flip 7 ends the round immediately AND removes the player from active players
		s.Game.CurrentRound.RemoveActivePlayer(p)
		s.Game.CurrentRound.End(domain.RoundEndReasonFlip7)

		if s.Logger != nil {
			s.Logger.Log(s.GameID, strconv.Itoa(s.Game.RoundCount), p.ID.String(), "Flip7", map[string]interface{}{
				"banked_score": score,
				"total_score":  p.TotalScore,
			})
		}

		return
	}
	// Show current hand score
	calc := domain.NewScoreCalculator()
	score := calc.Compute(p.CurrentHand)
	fmt.Printf("Current Hand: %s | Score: %d\n", s.formatHand(p.CurrentHand), score.Total)
}

// promptForTarget prompts the player to select a target for an action card.
// Valid targets include all active players, including the player themselves.
// Strategic reasons for self-targeting:
// - Freeze: Bank current points immediately (defensive play)
// - Flip Three: Force yourself to draw 3 cards (aggressive play when needing points)
// - GiveSecondChance: Pass Second Chance to a specific player
func (s *ManualGameService) promptForTarget(actionType domain.ActionType, candidates []*domain.Player, actor *domain.Player) *domain.Player {
	if len(candidates) == 0 {
		// Provide action-specific error messages
		switch actionType {
		case domain.ActionGiveSecondChance:
			fmt.Println("No valid targets available: All other players already have a Second Chance card.")
		default:
			fmt.Println("No active players available to target.")
		}
		return nil
	}

	// Display appropriate message based on action type
	switch actionType {
	case domain.ActionFreeze:
		fmt.Println("Select target to Freeze:")
	case domain.ActionFlipThree:
		fmt.Println("Select target for Flip Three:")
	case domain.ActionGiveSecondChance:
		fmt.Println("Select player to give Second Chance to:")
	default:
		fmt.Println("Select Target:")
	}

	// Suggestion Logic using AdaptiveStrategy
	adaptive := strategy.NewAdaptiveStrategy()
	if s.Game.CurrentRound != nil {
		adaptive.SetDeck(s.Game.CurrentRound.Deck)
	}
	suggested := adaptive.ChooseTarget(actionType, candidates, actor)

	for i, c := range candidates {
		fmt.Printf("%d. %s\n", i+1, s.FormatCandidateOption(c, suggested))
	}

	fmt.Print("Enter choice: ")
	input, _ := s.Reader.ReadString('\n')
	idx, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil || idx < 1 || idx > len(candidates) {
		return nil
	}
	return candidates[idx-1]
}

// FormatCandidateOption formats a candidate player for display in the selection list.
// It includes the player's name, score, hand contents, and marks the suggested candidate.
// Note: Returns "[]" for nil CurrentHand. In practice, this method is called during active
// gameplay when all candidates have initialized hands, but the nil check provides defensive
// programming against edge cases.
func (s *ManualGameService) FormatCandidateOption(candidate *domain.Player, suggested *domain.Player) string {
	marker := ""
	if suggested != nil && candidate.ID == suggested.ID {
		marker = " [Suggested]"
	}
	handStr := "[]"
	if candidate.CurrentHand != nil {
		handStr = s.formatHand(candidate.CurrentHand)
	}
	return fmt.Sprintf("%s (Score: %d) Hand: %s%s", candidate.Name, candidate.TotalScore, handStr, marker)
}

// resolveFlipThreeManual handles the Flip Three action effect on the target player.
// Per domain model (docs/domain_model.md lines 169-172), Flip Three forces the target to:
//  1. Draw 3 cards one by one
//  2. If Second Chance is drawn: Process normally (set aside/use if duplicate drawn)
//  3. If Flip Three or Freeze is drawn: Queue it and resolve AFTER all 3 cards are drawn
//     (only if the target hasn't busted)
//
// This function prompts the user to input 3 cards for the target player and processes
// each card according to these rules. The loop exits early if the target busts.
func (s *ManualGameService) resolveFlipThreeManual(target *domain.Player) {
	// Create FlipThree executor with manual mode implementations
	source := &manualFlipThreeCardSource{service: s}
	processor := &manualFlipThreeCardProcessor{service: s}

	// Create logger function that prints to console
	logger := func(message string) {
		fmt.Println(message)
	}

	executor := domain.NewFlipThreeExecutor(source, processor, logger)
	executor.Execute(target, s.Game.CurrentRound)
}

func (s *ManualGameService) formatHand(h *domain.PlayerHand) string {
	var parts []string
	for _, val := range h.RawNumberCards {
		parts = append(parts, strconv.Itoa(int(val)))
	}
	for _, mod := range h.ModifierCards {
		parts = append(parts, string(mod.ModifierType))
	}
	for _, act := range h.ActionCards {
		parts = append(parts, string(act.ActionType))
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func (s *ManualGameService) printWinner() {
	if len(s.Game.Winners) == 0 {
		fmt.Println("Game Over. No winner determined.")
		return
	}
	fmt.Println("Game Over. Winner(s):")
	for _, winner := range s.Game.Winners {
		fmt.Printf(" - %s with %d points\n", winner.Name, winner.TotalScore)
	}
}

// SaveState serializes the current game state to a base64 string.
func (s *ManualGameService) SaveState() (string, error) {
	// Collect IDs of user-controlled players (those with nil strategy)
	var userControlledIDs []string
	for _, p := range s.Game.Players {
		if p.Strategy == nil {
			userControlledIDs = append(userControlledIDs, p.ID.String())
		}
	}

	wrapper := gameStateWrapper{
		Game:              s.Game,
		UserControlledIDs: userControlledIDs,
		GameID:            s.GameID,
	}

	data, err := json.Marshal(wrapper)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

// LoadState deserializes the game state from a base64 string.
func (s *ManualGameService) LoadState(encoded string) error {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return fmt.Errorf("invalid code: %v", err)
	}

	var wrapper gameStateWrapper
	if err := json.Unmarshal(decoded, &wrapper); err != nil {
		return fmt.Errorf("failed to parse game state: %v", err)
	}

	// Validate that the game state is not nil
	if wrapper.Game == nil {
		return fmt.Errorf("invalid save code: game state is missing")
	}

	// Validate that the loaded game is resumable
	if wrapper.Game.IsCompleted {
		return fmt.Errorf("cannot resume: the loaded game is already completed")
	}
	if wrapper.Game.CurrentRound != nil && wrapper.Game.CurrentRound.IsEnded {
		return fmt.Errorf("cannot resume: the current round in the loaded game is already ended")
	}

	s.RelinkPointers(wrapper.Game, wrapper.UserControlledIDs)
	s.Game = wrapper.Game
	s.GameID = wrapper.GameID // Restore GameID for logging continuity
	return nil
}

// PushState captures the current game state and pushes it to history.
func (s *ManualGameService) PushState() {
	if s.Game == nil {
		return
	}
	state, err := s.SaveState()
	if err != nil {
		fmt.Printf("Warning: Failed to save state for history: %v\n", err)
		return
	}
	s.History.Push(GameMemento(state))
}

// Undo reverts the game state to the previous memento.
func (s *ManualGameService) Undo() {
	memento, ok := s.History.Undo()
	if !ok {
		fmt.Println("Cannot undo: No previous state.")
		return
	}
	if err := s.LoadState(string(memento)); err != nil {
		fmt.Printf("Error undoing state: %v\n", err)
		// Try to recover? At least we are at some state (likely the one we failed to leave or a broken one)
		// But LoadState overwrites s.Game. if it fails mid-way...
	} else {
		fmt.Println("Undid last action.")
	}
}

// Redo advances the game state to the next memento.
func (s *ManualGameService) Redo() {
	memento, ok := s.History.Redo()
	if !ok {
		fmt.Println("Cannot redo: No future state.")
		return
	}
	if err := s.LoadState(string(memento)); err != nil {
		fmt.Printf("Error redoing state: %v\n", err)
	} else {
		fmt.Println("Redid action.")
	}
}

// RelinkPointers restores pointer relationships after deserialization.
// It ensures that all references to players point to the same instances and restores strategies.
func (s *ManualGameService) RelinkPointers(g *domain.Game, userControlledIDs []string) {
	// Create a set of user-controlled player IDs for quick lookup
	userControlledSet := make(map[string]bool)
	for _, id := range userControlledIDs {
		userControlledSet[id] = true
	}

	playerMap := make(map[string]*domain.Player)
	for _, p := range g.Players {
		playerMap[p.ID.String()] = p
		// Restore strategies based on user control
		if userControlledSet[p.ID.String()] {
			p.Strategy = nil
		} else {
			// Default to ProbabilisticStrategy for AI players in manual mode
			p.Strategy = &strategy.ProbabilisticStrategy{}
		}
	}

	if g.CurrentRound != nil {
		// Relink Round Players
		for i, p := range g.CurrentRound.Players {
			if existing, ok := playerMap[p.ID.String()]; ok {
				g.CurrentRound.Players[i] = existing
			}
		}
		// Relink Active Players
		for i, p := range g.CurrentRound.ActivePlayers {
			if existing, ok := playerMap[p.ID.String()]; ok {
				g.CurrentRound.ActivePlayers[i] = existing
			}
		}
		// Relink Dealer
		if g.CurrentRound.Dealer != nil {
			if existing, ok := playerMap[g.CurrentRound.Dealer.ID.String()]; ok {
				g.CurrentRound.Dealer = existing
			}
		}
		// Relink Deck
		if g.Deck != nil {
			g.CurrentRound.Deck = g.Deck
		}
	}

	// Relink Winners
	for i, p := range g.Winners {
		if existing, ok := playerMap[p.ID.String()]; ok {
			g.Winners[i] = existing
		}
	}
}
