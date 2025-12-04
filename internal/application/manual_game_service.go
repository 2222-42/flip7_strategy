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
	"flip7_strategy/internal/domain/rules"
	"flip7_strategy/internal/domain/strategy"
)

const FlipThreeCardCount = 3

// gameStateWrapper wraps the game state with metadata for serialization.
type gameStateWrapper struct {
	Game              *domain.Game `json:"game"`
	UserControlledIDs []string     `json:"user_controlled_ids"` // IDs of players with nil strategy
	GameID            string       `json:"game_id"`             // GameID for logging continuity
}

// ManualGameService handles the manual mode where the user inputs game events.
type ManualGameService struct {
	Game   *domain.Game
	Reader *bufio.Reader
	Logger logger.GameLogger
	GameID string
}

// NewManualGameService creates a new ManualGameService.
func NewManualGameService(reader *bufio.Reader, logger logger.GameLogger) *ManualGameService {
	return &ManualGameService{
		Reader: reader,
		Logger: logger,
		GameID: fmt.Sprintf("game_%d", time.Now().Unix()),
	}
}

// Run starts the manual game loop.
func (s *ManualGameService) Run() {
	fmt.Println("\n--- Manual Mode ---")
	s.setupPlayers()
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
		// Rotate dealer for next round
		s.Game.DealerIndex = (s.Game.DealerIndex + 1) % len(s.Game.Players)

		// Check for winners
		winners := s.Game.DetermineWinners()
		if len(winners) > 0 {
			s.Game.Winners = winners
			s.Game.IsCompleted = true
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
		deck := domain.NewDeck()
		dealer := s.Game.Players[s.Game.DealerIndex]
		s.Game.CurrentRound = domain.NewRound(s.Game.Players, dealer, deck)
		fmt.Printf("\n--- New Round! Dealer: %s ---\n", dealer.Name)

		if s.Logger != nil {
			s.Logger.Log(s.GameID, strconv.Itoa(s.Game.RoundCount), "system", "RoundStart", map[string]interface{}{
				"dealer": dealer.Name,
			})
		}
	} else {
		fmt.Println("Resuming round...")
	}

	for !s.Game.CurrentRound.IsEnded {
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

		for !turnEnded {
			fmt.Print("Input (0-12, +N, x2, F, T, C, S): ")
			input, err := s.Reader.ReadString('\n')
			if err != nil {
				fmt.Println("Error reading input. Exiting game.")
				return
			}
			input = strings.TrimSpace(input)

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

		// Check if round ended during this loop (Flip 7 or all stayed)
		if s.Game.CurrentRound.IsEnded {
			break
		}

		// Update Turn Index
		if !playerRemoved {
			s.Game.CurrentRound.CurrentTurnIndex++
		}
		// If player removed (Freeze action), the next player slides into the current index, so we don't increment.
		// Busted players remain in ActivePlayers but are skipped via the status check at the start of the loop.
	}
}

func (s *ManualGameService) analyzeState(p *domain.Player) {
	// Show bust rate
	risk := s.Game.CurrentRound.Deck.EstimateHitRisk(p.CurrentHand.NumberCards)
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
	if s.Game.CurrentRound == nil || s.Game.CurrentRound.Deck == nil {
		return fmt.Errorf("no active round/deck")
	}
	deck := s.Game.CurrentRound.Deck
	// Find and remove card from deck.Cards
	for i, c := range deck.Cards {
		match := false
		if c.Type == card.Type {
			if c.Type == domain.CardTypeNumber && c.Value == card.Value {
				match = true
			} else if c.Type == domain.CardTypeModifier && c.ModifierType == card.ModifierType {
				match = true
			} else if c.Type == domain.CardTypeAction && c.ActionType == card.ActionType {
				match = true
			}
		}

		if match {
			// Remove it
			deck.Cards = append(deck.Cards[:i], deck.Cards[i+1:]...)
			// Update counts if number
			if card.Type == domain.CardTypeNumber {
				deck.RemainingCounts[card.Value]--
			}
			return nil
		}
	}
	return fmt.Errorf("card not found in deck (already drawn?)")
}

// GetCard implements rules.CardSource for Manual Mode (used during Flip Three).
func (s *ManualGameService) GetCard() (domain.Card, error) {
	fmt.Print("Input card (0-12, +N, x2, F, T, C): ")
	input, err := s.Reader.ReadString('\n')
	if err != nil {
		return domain.Card{}, err
	}
	input = strings.TrimSpace(input)
	card, err := s.parseInput(input)
	if err != nil {
		return domain.Card{}, err
	}
	if err := s.removeCardFromDeck(card); err != nil {
		return domain.Card{}, err
	}
	return card, nil
}

// SelectTarget implements rules.TargetSelector for Manual Mode.
func (s *ManualGameService) SelectTarget(actionType domain.ActionType, candidates []*domain.Player, source *domain.Player) *domain.Player {
	return s.promptForTarget(source)
}

// processCard handles the logic of adding a card to a player's hand and resolving its effects.
func (s *ManualGameService) processCard(p *domain.Player, card domain.Card) {
	fmt.Printf("Played: %v\n", card)

	if s.Logger != nil {
		s.Logger.Log(s.GameID, strconv.Itoa(s.Game.RoundCount), p.ID.String(), "CardPlayed", map[string]interface{}{
			"card": card.String(),
		})
	}

	engine := rules.NewGameEngine()
	result, err := engine.ApplyCard(s.Game.CurrentRound, p, card, s)
	if err != nil {
		fmt.Printf("Error applying card: %v\n", err)
		return
	}

	// Log Discards
	if len(result.Discarded) > 0 {
		for _, d := range result.Discarded {
			fmt.Printf("Discarded: %v\n", d)
		}
	}

	// Log Outcomes
	if result.Busted {
		fmt.Println("BUSTED!")
		if s.Logger != nil {
			s.Logger.Log(s.GameID, strconv.Itoa(s.Game.RoundCount), p.ID.String(), "Bust", map[string]interface{}{
				"hand": s.formatHand(p.CurrentHand),
			})
		}
	} else if result.Flip7 {
		fmt.Println("FLIP 7!")
		fmt.Printf("%s banked %d points! Total: %d\n", p.Name, result.BankedScore, p.TotalScore)
		if s.Logger != nil {
			s.Logger.Log(s.GameID, strconv.Itoa(s.Game.RoundCount), p.ID.String(), "Flip7", map[string]interface{}{
				"banked_score": result.BankedScore,
				"total_score":  p.TotalScore,
			})
		}
	} else {
		// Show current hand score (if not busted/flip7)
		calc := domain.NewScoreCalculator()
		score := calc.Compute(p.CurrentHand)
		fmt.Printf("Current Hand: %s | Score: %d\n", s.formatHand(p.CurrentHand), score.Total)
	}

	// Handle Actions
	if result.ActionType != "" {
		switch result.ActionType {
		case domain.ActionFreeze:
			fmt.Printf("Freezing %s!\n", result.Target.Name)
			fmt.Printf("%s banked %d points! Total: %d\n", result.Target.Name, result.Target.TotalScore, result.Target.TotalScore) // Score updated in engine
		case domain.ActionFlipThree:
			fmt.Printf("Flip Three on %s! They must draw 3 cards.\n", result.Target.Name)
			s.resolveFlipThreeManual(result.Target)
		}
	}
}

// promptForTarget prompts the player to select a target for an action card.
// Valid targets include all active players, including the player themselves.
// Strategic reasons for self-targeting:
// - Freeze: Bank current points immediately (defensive play)
// - Flip Three: Force yourself to draw 3 cards (aggressive play when needing points)
func (s *ManualGameService) promptForTarget(p *domain.Player) *domain.Player {
	candidates := s.Game.CurrentRound.ActivePlayers
	if len(candidates) == 0 {
		fmt.Println("No active players to target.")
		return nil
	}

	fmt.Println("Select Target:")
	for i, c := range candidates {
		fmt.Printf("%d. %s (Score: %d)\n", i+1, c.Name, c.TotalScore)
	}

	fmt.Print("Enter choice: ")
	input, _ := s.Reader.ReadString('\n')
	idx, err := strconv.Atoi(strings.TrimSpace(input))
	if err != nil || idx < 1 || idx > len(candidates) {
		return nil
	}
	return candidates[idx-1]
}

// resolveFlipThreeManual handles the Flip Three action effect on the target player.
func (s *ManualGameService) resolveFlipThreeManual(target *domain.Player) {
	engine := rules.NewGameEngine()
	results, err := engine.ExecuteFlipThree(s.Game.CurrentRound, target, s, s)
	if err != nil {
		fmt.Printf("Error executing Flip Three: %v\n", err)
		return
	}

	for i, res := range results {
		fmt.Printf("%s forced draw (%d/3) result: Busted=%v, Action=%s\n", target.Name, i+1, res.Busted, res.ActionType)
		if res.Busted {
			fmt.Printf("%s BUSTED in Flip Three!\n", target.Name)
			if s.Logger != nil {
				s.Logger.Log(s.GameID, strconv.Itoa(s.Game.RoundCount), target.ID.String(), "Bust", map[string]interface{}{
					"hand": s.formatHand(target.CurrentHand),
				})
			}
		}
		if res.Flip7 {
			fmt.Printf("%s FLIP 7 in Flip Three!\n", target.Name)
			fmt.Printf("%s banked %d points! Total: %d\n", target.Name, res.BankedScore, target.TotalScore)
			if s.Logger != nil {
				s.Logger.Log(s.GameID, strconv.Itoa(s.Game.RoundCount), target.ID.String(), "Flip7", map[string]interface{}{
					"banked_score": res.BankedScore,
					"total_score":  target.TotalScore,
				})
			}
		}
		if res.ActionType == domain.ActionFreeze && res.Target != nil {
			fmt.Printf("Freeze resolved on %s\n", res.Target.Name)
			fmt.Printf("%s banked %d points! Total: %d\n", res.Target.Name, res.Target.TotalScore, res.Target.TotalScore)
		}
		if res.ActionType == domain.ActionFlipThree && res.Target != nil {
			fmt.Printf("Nested Flip Three on %s\n", res.Target.Name)
			s.resolveFlipThreeManual(res.Target) // Recursive
		}
	}
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
	}

	// Relink Winners
	for i, p := range g.Winners {
		if existing, ok := playerMap[p.ID.String()]; ok {
			g.Winners[i] = existing
		}
	}
}
