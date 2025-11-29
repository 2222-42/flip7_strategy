package application

import (
	"bufio"
	"fmt"
	"strconv"
	"strings"

	"flip7_strategy/internal/domain"
	"flip7_strategy/internal/domain/strategy"
)

const FlipThreeCardCount = 3

// ManualGameService handles the manual mode where the user inputs game events.
type ManualGameService struct {
	Game   *domain.Game
	Reader *bufio.Reader
}

// NewManualGameService creates a new ManualGameService.
func NewManualGameService(reader *bufio.Reader) *ManualGameService {
	return &ManualGameService{
		Reader: reader,
	}
}

// Run starts the manual game loop.
func (s *ManualGameService) Run() {
	fmt.Println("\n--- Manual Mode ---")
	s.setupPlayers()
	s.gameLoop()
}

func (s *ManualGameService) setupPlayers() {
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
	fmt.Println("Game started!")
}

func (s *ManualGameService) gameLoop() {
	for !s.Game.IsCompleted {
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
}

func (s *ManualGameService) playRound() {
	// Initialize new round with deck
	deck := domain.NewDeck()
	dealer := s.Game.Players[s.Game.DealerIndex]
	s.Game.CurrentRound = domain.NewRound(s.Game.Players, dealer, deck)

	fmt.Printf("\n--- New Round! Dealer: %s ---\n", dealer.Name)

	// Reset hands (NewRound does this, but let's be sure)
	// NewRound calls StartNewRound which resets hands.

	for !s.Game.CurrentRound.IsEnded {
		// Round-Robin: Iterate through active players once per "round" of turns
		// But wait, IsEnded checks if active players are empty.
		// We need to iterate through a snapshot of active players.

		activePlayers := make([]*domain.Player, len(s.Game.CurrentRound.ActivePlayers))
		copy(activePlayers, s.Game.CurrentRound.ActivePlayers)

		for _, currentPlayer := range activePlayers {
			// Check if player is still active (might have been removed by previous player's action)
			if currentPlayer.CurrentHand.Status != domain.HandStatusActive {
				continue
			}

			fmt.Printf("\n>>> Turn: %s (Score: %d)\n", currentPlayer.Name, currentPlayer.TotalScore)
			s.analyzeState(currentPlayer)

			// Input loop for this turn (single action)
			turnEnded := false
			for !turnEnded {
				fmt.Print("Input (0-12, +N, x2, F, T, C, S): ")
				input, err := s.Reader.ReadString('\n')
				if err != nil {
					fmt.Println("Error reading input. Exiting game.")
					return
				}
				input = strings.TrimSpace(input)

				if strings.EqualFold(input, "S") {
					// Validation: Cannot stay on first turn (empty hand)
					// Validation: Cannot stay on first turn (empty hand) unless special conditions met
					if !currentPlayer.CurrentHand.CanStay() {
						fmt.Println("Invalid move: You must hit on your first turn (unless you have points or specific actions)!")
						continue
					}

					currentPlayer.CurrentHand.Status = domain.HandStatusStayed
					score := currentPlayer.BankCurrentHand()
					fmt.Printf("%s banked %d points! Total: %d\n", currentPlayer.Name, score, currentPlayer.TotalScore)
					s.Game.CurrentRound.RemoveActivePlayer(currentPlayer)
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

					// Turn always ends after one action (Hit) or Action card
					turnEnded = true
				}
			}

			// Check if round ended during this loop (Flip 7 or all stayed)
			if s.Game.CurrentRound.IsEnded {
				break
			}
		}
		// After round-robin pass, check if all players are inactive and round is not ended
		if !s.Game.CurrentRound.IsEnded && len(s.Game.CurrentRound.ActivePlayers) == 0 {
			s.Game.CurrentRound.End(domain.RoundEndReasonNoActivePlayers)
			break
		}
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

// processCard handles the logic of adding a card to a player's hand and resolving its effects.
//
// Action Card Processing Order (per domain model - docs/domain_model.md lines 169-175):
// - Flip Three & Freeze: The player who DRAWS the card chooses a target player.
//   The action effect is applied to the TARGET player, then the card is added to the DRAWER's hand.
//   This order is important because:
//   1. The drawer must choose a target before knowing the full outcome
//   2. The target processes the effect (e.g., draws 3 cards for Flip Three)
//   3. Only after resolution does the drawer add the action card to their own hand
//
// - Second Chance: Added to drawer's hand immediately. If the drawer already has one,
//   they must pass it to another active player (or discard if no valid target).
//   NOTE: In Manual Mode, this passing logic is NOT automated. The user must manually
//   handle Second Chance passing by not inputting the card for the original drawer
//   and instead inputting it for the target player on their turn.
//
// - During Flip Three: If the target draws another Flip Three or Freeze while drawing
//   their 3 cards, those action cards are queued and resolved AFTER all 3 cards are drawn
//   (if the target hasn't busted). See resolveFlipThreeManual for details.
//
// Number/Modifier Cards: Added to the player's hand immediately, checked for bust/flip7.
func (s *ManualGameService) processCard(p *domain.Player, card domain.Card) {
	fmt.Printf("Played: %v\n", card)

	// Special handling for Actions
	if card.Type == domain.CardTypeAction {
		if card.ActionType == domain.ActionFlipThree || card.ActionType == domain.ActionFreeze {
			// Step 1: Prompt the drawer (p) to choose a target player for the action
			target := s.promptForTarget(p)
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

	// Add card to hand logic
	busted, flip7, _ := p.CurrentHand.AddCard(card)

	if busted {
		fmt.Println("BUSTED!")
		p.CurrentHand.Status = domain.HandStatusBusted
		return
	} else if flip7 {
		fmt.Println("FLIP 7!")
		p.CurrentHand.Status = domain.HandStatusStayed
		score := p.BankCurrentHand()
		fmt.Printf("%s banked %d points! Total: %d\n", p.Name, score, p.TotalScore)

		// Flip 7 ends the round immediately
		s.Game.CurrentRound.IsEnded = true
		s.Game.CurrentRound.EndReason = domain.RoundEndReasonFlip7

		return
	} else {
		// Show current hand score
		calc := domain.NewScoreCalculator()
		score := calc.Compute(p.CurrentHand)
		fmt.Printf("Current Hand: %s | Score: %d\n", s.formatHand(p.CurrentHand), score.Total)
		return
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
// Per domain model (docs/domain_model.md lines 169-172), Flip Three forces the target to:
// 1. Draw 3 cards one by one
// 2. If Second Chance is drawn: Process normally (set aside/use if duplicate drawn)
// 3. If Flip Three or Freeze is drawn: Queue it and resolve AFTER all 3 cards are drawn
//    (only if the target hasn't busted)
//
// This function prompts the user to input 3 cards for the target player and processes
// each card according to these rules. The loop exits early if the target busts.
func (s *ManualGameService) resolveFlipThreeManual(target *domain.Player) {
	for i := 0; i < FlipThreeCardCount; i++ {
		// Exit early if target is no longer active (busted or other status change)
		if target.CurrentHand.Status != domain.HandStatusActive {
			break
		}
		
		fmt.Printf("Input card %d/%d for %s: ", i+1, FlipThreeCardCount, target.Name)
		input, _ := s.Reader.ReadString('\n')
		input = strings.TrimSpace(input)

		card, err := s.parseInput(input)
		if err != nil {
			fmt.Printf("Invalid input: %v. Try again.\n", err)
			i-- // Retry this card
			continue
		}

		if err := s.removeCardFromDeck(card); err != nil {
			fmt.Printf("Error: %v. Try again.\n", err)
			i-- // Retry this card
			continue
		}

		// Handle cards drawn during Flip Three according to game rules:
		// - Number/Modifier/Second Chance: Process immediately via processCard
		// - Flip Three/Freeze: Queue for later resolution (added to hand, resolved manually after)
		//
		// Note: In Manual Mode, we simplify by adding queued actions to the hand with a warning.
		// The user is expected to track and resolve these actions manually after the 3 draws.
		// A full implementation would maintain a separate action queue and auto-resolve.
		if card.Type == domain.CardTypeAction && (card.ActionType == domain.ActionFlipThree || card.ActionType == domain.ActionFreeze) {
			fmt.Println("Action card drawn during Flip Three! Per game rules, this should be queued.")
			fmt.Println("(Manual Mode: Card added to hand. Resolve manually after completing all 3 draws.)")
			target.CurrentHand.ActionCards = append(target.CurrentHand.ActionCards, card)
		} else {
			// Process number cards, modifiers, and Second Chance immediately
			s.processCard(target, card)
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
