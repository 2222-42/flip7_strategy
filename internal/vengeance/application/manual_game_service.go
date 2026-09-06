package application

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"flip7_strategy/internal/vengeance/domain"
	"flip7_strategy/internal/vengeance/domain/strategy"
)

const inputHelp = `Cards: 1-13, 0/Z (Zero), U (Unlucky 7), L (Lucky 13), -2/-4/-6/-8/-10, /2
Actions: J (Just One More), F4 (Flip Four), SW (Swap), ST (Steal), DI (Discard)
Commands: S stay, UNDO/<, REDO/>, HANDS, HELP`

var (
	errRewound = errors.New("rewound")
	errEOF     = errors.New("end of input")
)

type promptResult int

const (
	promptOK promptResult = iota
	promptRewind
	promptEOF
)

// ManualGameService is a helper for a physical Vengeance game: the user types
// the cards that were flipped, and Adaptive suggests Hit/Stay and targets.
type ManualGameService struct {
	Game    *domain.Game
	Reader  *bufio.Reader
	Advisor domain.Strategy
	History gameHistory
	phase   string
	cursor  int
	rewound bool
	eof     bool
}

func NewManualGameService(reader *bufio.Reader) *ManualGameService {
	return &ManualGameService{
		Reader:  reader,
		Advisor: strategy.NewAdaptiveStrategy(),
	}
}

func (s *ManualGameService) Run() {
	fmt.Println("\n--- Vengeance Manual Mode (physical game helper) ---")
	fmt.Println(inputHelp)
	s.setupPlayers()
	s.gameLoop()
}

func (s *ManualGameService) setupPlayers() {
	fmt.Print("Enter number of players: ")
	n := s.readInt(0)
	if n < 1 {
		n = 3
		fmt.Println("Defaulting to 3 players.")
	}

	players := make([]*domain.Player, 0, n)
	me := domain.NewPlayer("Me", s.Advisor)
	players = append(players, me)
	for i := 1; i < n; i++ {
		fmt.Printf("Enter name for Player %d: ", i+1)
		name := s.readLine()
		if name == "" {
			name = fmt.Sprintf("Player %d", i+1)
		}
		players = append(players, domain.NewPlayer(name, s.Advisor))
	}

	fmt.Println("Select dealer (start player is the one on their left):")
	for i, p := range players {
		fmt.Printf("%d. %s\n", i+1, p.Name)
	}
	fmt.Print("Enter choice: ")
	idx := s.readInt(1)
	if idx < 1 || idx > n {
		idx = 1
	}

	s.Game = domain.NewGame(players)
	s.Game.DealerIndex = idx - 1
	s.Game.Deck = domain.NewUnshuffledDeck()
	fmt.Println("Game started. Type the cards as they appear.")
}

func (s *ManualGameService) gameLoop() {
	for !s.Game.IsCompleted {
		s.Game.RoundCount++
		s.playRound()
		if s.eof {
			fmt.Println("End of input. Exiting.")
			return
		}
		s.collectTableCards()
		winners := s.Game.DetermineWinners()
		if len(winners) > 0 {
			s.Game.Winners = winners
			s.Game.IsCompleted = true
			break
		}
		s.Game.DealerIndex = (s.Game.DealerIndex + 1) % len(s.Game.Players)
		if s.Game.CurrentRound != nil {
			s.Game.Deck = s.Game.CurrentRound.Deck
		}
	}
	s.printWinner()
}

func (s *ManualGameService) collectTableCards() {
	if s.Game.CurrentRound == nil {
		return
	}
	for _, p := range s.Game.Players {
		if p.CurrentHand == nil {
			continue
		}
		s.Game.DiscardPile = append(s.Game.DiscardPile, p.CurrentHand.NumberLine...)
		s.Game.DiscardPile = append(s.Game.DiscardPile, p.CurrentHand.ModifierLine...)
	}
}

func (s *ManualGameService) playRound() {
	if s.Game.Deck == nil {
		s.Game.Deck = domain.NewUnshuffledDeck()
	}
	if s.Game.CurrentRound == nil || s.Game.CurrentRound.IsEnded {
		dealer := s.Game.Players[s.Game.DealerIndex]
		s.Game.CurrentRound = domain.NewRound(s.Game.Players, dealer, s.Game.Deck)
		s.phase = phaseDeal
		s.cursor = 0
		fmt.Printf("\n--- Round %d  Dealer: %s ---\n", s.Game.RoundCount, dealer.Name)
		fmt.Println("Initial deal (clockwise from dealer's left).")
		s.PushState()
	}

	for !s.Game.CurrentRound.IsEnded {
		if s.eof {
			return
		}
		order := s.Game.CurrentRound.OfferOrder
		if len(order) == 0 {
			s.Game.CurrentRound.End(domain.RoundEndReasonNoActivePlayers)
			break
		}
		if s.phase == phaseDeal {
			if s.cursor >= len(order) {
				s.phase = phaseTurns
				s.cursor = 0
				continue
			}
			p := order[s.cursor]
			if p.CurrentHand == nil || p.CurrentHand.Status != domain.HandStatusActive {
				s.cursor++
				continue
			}
			s.printTable()
			fmt.Printf("Deal to %s.\n", p.Name)
			switch s.promptAndProcess(p, false) {
			case promptEOF:
				return
			case promptRewind:
				s.rewound = false
				continue
			}
			if s.Game.CurrentRound.IsEnded {
				break
			}
			s.cursor++
			if s.cursor >= len(order) {
				s.phase = phaseTurns
				s.cursor = 0
			}
			s.PushState()
			continue
		}

		if len(s.Game.CurrentRound.ActivePlayers()) == 0 {
			s.Game.CurrentRound.End(domain.RoundEndReasonNoActivePlayers)
			break
		}
		if s.cursor < 0 || s.cursor >= len(order) {
			s.cursor = 0
		}
		start := s.cursor
		acted := false
		for {
			p := order[s.cursor]
			if p.CurrentHand != nil && p.CurrentHand.Status == domain.HandStatusActive {
				s.printTable()
				s.analyzeState(p)
				switch s.promptAndProcess(p, true) {
				case promptEOF:
					return
				case promptRewind:
					s.rewound = false
				default:
					if !s.Game.CurrentRound.IsEnded {
						s.cursor = (s.cursor + 1) % len(order)
						s.PushState()
					}
				}
				acted = true
				break
			}
			s.cursor = (s.cursor + 1) % len(order)
			if s.cursor == start {
				break
			}
		}
		if s.Game.CurrentRound.IsEnded {
			break
		}
		if !acted {
			s.Game.CurrentRound.End(domain.RoundEndReasonNoActivePlayers)
		}
	}

	if s.eof {
		return
	}
	s.scoreNonBusted()
}

func (s *ManualGameService) scoreNonBusted() {
	fmt.Println("\n--- End of round ---")
	for _, p := range s.Game.Players {
		if p.CurrentHand == nil || p.CurrentHand.Status == domain.HandStatusBusted {
			fmt.Printf("%s scores 0 (busted). Total: %d\n", p.Name, p.TotalScore)
			continue
		}
		score := p.BankCurrentHand()
		fmt.Printf("%s scores %d. Total: %d\n", p.Name, score, p.TotalScore)
	}
}

func (s *ManualGameService) analyzeState(p *domain.Player) {
	fmt.Printf("\n>>> Turn: %s (Total: %d)\n", p.Name, p.TotalScore)
	calc := domain.NewScoreCalculator()
	fmt.Printf("Current Hand: %s | Round: %d\n", formatHand(p.CurrentHand), calc.Compute(p.CurrentHand).Total)

	deck := s.riskDeck()
	risk := 0.0
	if deck != nil {
		risk = deck.EstimateHitRisk(p.CurrentHand)
	}
	fmt.Printf("Bust Rate: %.2f%%\n", risk*100)

	ctx := domain.DecisionContext{
		Deck:         s.Game.CurrentRound.Deck,
		DiscardPile:  s.Game.DiscardPile,
		Hand:         p.CurrentHand,
		PlayerScore:  p.TotalScore,
		OtherPlayers: s.others(p),
	}
	if d, ok := s.Advisor.(interface{ SetDeck(*domain.Deck) }); ok {
		d.SetDeck(deck)
	}
	choice := s.Advisor.Decide(ctx)
	if p.CurrentHand.MustHit() {
		fmt.Println("Suggested Move: hit (The Zero: must hit)")
		return
	}
	fmt.Printf("Suggested Move: %s (Adaptive)\n", choice)
}

func (s *ManualGameService) promptAndProcess(p *domain.Player, allowStay bool) promptResult {
	for {
		fmt.Print("Input: ")
		input := s.readLine()
		if s.eof {
			return promptEOF
		}
		if s.handleHistoryCommand(input, false) {
			if s.rewound {
				return promptRewind
			}
			continue
		}
		switch strings.ToUpper(input) {
		case "HELP", "H", "?":
			fmt.Println(inputHelp)
			continue
		case "HANDS":
			s.printTable()
			continue
		case "S", "STAY":
			if !allowStay {
				fmt.Println("Initial deal: enter the card that was flipped.")
				continue
			}
			if !p.CurrentHand.CanStay() {
				fmt.Println("Cannot stay (The Zero forces Hit, or already inactive).")
				continue
			}
			p.CurrentHand.Stay()
			fmt.Printf("%s stays. Score banks at round end.\n", p.Name)
			return promptOK
		}

		spec, err := ParseManualInput(input)
		if err != nil {
			fmt.Printf("Invalid input: %v. Type HELP for codes.\n", err)
			continue
		}
		card, err := s.takeFromDeck(spec)
		if err != nil {
			fmt.Printf("Error: %v. Try again.\n", err)
			continue
		}
		s.processCard(p, card)
		if s.eof {
			return promptEOF
		}
		if s.rewound {
			return promptRewind
		}
		return promptOK
	}
}

func (s *ManualGameService) handleHistoryCommand(input string, nested bool) bool {
	if isUndoCommand(input) {
		if nested {
			s.tryAbortCurrent()
		} else {
			s.tryUndo()
		}
		return true
	}
	if isRedoCommand(input) {
		s.tryRedo()
		return true
	}
	return false
}

func isUndoCommand(input string) bool {
	u := strings.ToUpper(strings.TrimSpace(input))
	return u == "UNDO" || input == "<"
}

func isRedoCommand(input string) bool {
	u := strings.ToUpper(strings.TrimSpace(input))
	return u == "REDO" || u == "R" || input == ">"
}

func (s *ManualGameService) aborted() bool {
	return s.rewound || s.eof
}

func (s *ManualGameService) takeFromDeck(spec domain.CardSpec) (domain.TableCard, error) {
	round := s.Game.CurrentRound
	if round == nil || round.Deck == nil {
		return domain.TableCard{}, fmt.Errorf("no active deck")
	}
	if card, ok := round.Deck.RemoveMatching(spec); ok {
		return card, nil
	}
	if len(s.Game.DiscardPile) > 0 {
		fmt.Printf("Not in draw pile. Reshuffling %d discarded cards...\n", len(s.Game.DiscardPile))
		left := round.Deck.Cards
		combined := append(append([]domain.TableCard{}, s.Game.DiscardPile...), left...)
		round.Deck = domain.NewDeckFromCards(combined)
		s.Game.Deck = round.Deck
		s.Game.DiscardPile = nil
		if card, ok := round.Deck.RemoveMatching(spec); ok {
			return card, nil
		}
	}
	return domain.TableCard{}, fmt.Errorf("card not found in deck (already drawn?)")
}

func (s *ManualGameService) processCard(actor *domain.Player, card domain.TableCard) {
	fmt.Printf("Played: %s\n", card)
	if s.Game.CurrentRound.IsEnded {
		s.discard(card)
		return
	}
	switch card.Spec.Type {
	case domain.CardTypeNumber, domain.CardTypeSpecialNumber:
		s.receiveNumber(actor, card)
	case domain.CardTypeModifier:
		s.assignModifier(actor, card)
	case domain.CardTypeAction:
		s.resolveAction(actor, card)
	}
	if actor.CurrentHand != nil {
		calc := domain.NewScoreCalculator()
		fmt.Printf("Current Hand: %s | Round: %d\n", formatHand(actor.CurrentHand), calc.Compute(actor.CurrentHand).Total)
	}
}

func (s *ManualGameService) receiveNumber(target *domain.Player, card domain.TableCard) {
	res := target.CurrentHand.ReceiveNumberLike(card)
	if len(res.Wiped) > 0 {
		s.Game.DiscardPile = append(s.Game.DiscardPile, res.Wiped...)
		fmt.Printf("%s Unlucky 7: wipe %d cards, keep the 7.\n", target.Name, len(res.Wiped))
	}
	if res.Busted {
		fmt.Printf("%s BUSTED!\n", target.Name)
		return
	}
	if res.Flip7 {
		fmt.Printf("%s FLIP 7! Round over.\n", target.Name)
		s.Game.CurrentRound.End(domain.RoundEndReasonFlip7)
	}
}

func (s *ManualGameService) assignModifier(actor *domain.Player, card domain.TableCard) {
	candidates := s.Game.CurrentRound.NonBustedPlayers()
	if len(candidates) == 0 {
		s.discard(card)
		return
	}
	s.prepareAdvisor()
	suggested := s.Advisor.ChooseModifierTarget(card.Spec.ModifierType, candidates, actor)
	fmt.Printf("Assign %s. Suggested: %s\n", card, nameOf(suggested))
	target := s.promptPlayer("Modifier target", candidates, suggested)
	if s.aborted() {
		return
	}
	if target == nil {
		target = suggested
	}
	if target == nil {
		target = actor
	}
	fmt.Printf("%s gets %s\n", target.Name, card)
	target.CurrentHand.ReceiveModifier(card)
}

func (s *ManualGameService) resolveAction(actor *domain.Player, card domain.TableCard) {
	switch card.Spec.ActionType {
	case domain.ActionJustOneMore:
		s.discard(card)
		target := s.promptActionPlayer(actor, domain.ActionJustOneMore, "Just One More")
		if s.aborted() {
			return
		}
		fmt.Printf("Just One More on %s. Enter the forced card.\n", target.Name)
		forced, err := s.readCardFromTable()
		if err != nil {
			if !s.aborted() {
				fmt.Printf("%v\n", err)
			}
			return
		}
		s.processCard(target, forced)
		if s.aborted() || s.Game.CurrentRound.IsEnded {
			return
		}
		if target.CurrentHand.Status == domain.HandStatusActive {
			target.CurrentHand.Stay()
			fmt.Printf("%s is forced to stay.\n", target.Name)
		}
	case domain.ActionFlipFour:
		s.discard(card)
		target := s.promptActionPlayer(actor, domain.ActionFlipFour, "Flip Four")
		if s.aborted() {
			return
		}
		s.runFlipFour(target)
	case domain.ActionSteal:
		s.executeSteal(actor, card)
	case domain.ActionDiscard:
		s.executeDiscard(actor, card)
	case domain.ActionSwap:
		s.executeSwap(actor, card)
	default:
		s.discard(card)
	}
}

func (s *ManualGameService) runFlipFour(target *domain.Player) {
	delayed := make([]domain.TableCard, 0, domain.FlipFourCardCount)
	for i := 0; i < domain.FlipFourCardCount; i++ {
		if s.Game.CurrentRound.IsEnded || target.CurrentHand.Status == domain.HandStatusBusted {
			break
		}
		fmt.Printf("Flip Four %d/4 for %s.\n", i+1, target.Name)
		card, err := s.readCardFromTable()
		if err != nil {
			if s.aborted() {
				return
			}
			fmt.Printf("%v\n", err)
			s.discardAll(delayed)
			return
		}
		if card.Spec.IsNumberLike() {
			s.receiveNumber(target, card)
			continue
		}
		delayed = append(delayed, card)
	}
	if s.Game.CurrentRound.IsEnded || target.CurrentHand.Status == domain.HandStatusBusted {
		s.discardAll(delayed)
		return
	}
	for _, card := range delayed {
		if s.aborted() {
			return
		}
		if s.Game.CurrentRound.IsEnded {
			s.discard(card)
			continue
		}
		s.processCard(target, card)
	}
}

func (s *ManualGameService) executeSteal(actor *domain.Player, action domain.TableCard) {
	refs := s.Game.FaceUpRefs()
	s.discard(action)
	if len(refs) == 0 {
		fmt.Println("Steal: no face-up cards. Discarded.")
		return
	}
	s.prepareAdvisor()
	suggested := s.Advisor.ChooseCardTarget(domain.ActionSteal, refs, actor)
	ref := s.promptCardRef("Steal", refs, suggested)
	if s.aborted() {
		return
	}
	if ref == nil {
		ref = suggested
	}
	if ref == nil {
		ref = fallbackCardRef(refs, actor)
	}
	owner := s.Game.PlayerByID(ref.OwnerID)
	if owner == nil || owner.CurrentHand == nil {
		fmt.Println("Steal failed (owner gone).")
		return
	}
	stolen, ok := owner.CurrentHand.RemoveCard(ref.ID)
	if !ok {
		fmt.Println("Steal failed (card gone).")
		return
	}
	fmt.Printf("%s steals %s from %s\n", actor.Name, stolen, owner.Name)
	if stolen.Spec.IsNumberLike() {
		s.receiveNumber(actor, stolen)
		return
	}
	if stolen.Spec.Type == domain.CardTypeModifier {
		actor.CurrentHand.ReceiveModifier(stolen)
	}
}

func (s *ManualGameService) executeDiscard(actor *domain.Player, action domain.TableCard) {
	refs := s.Game.FaceUpRefs()
	s.discard(action)
	if len(refs) == 0 {
		fmt.Println("Discard: no face-up cards. Discarded.")
		return
	}
	s.prepareAdvisor()
	suggested := s.Advisor.ChooseCardTarget(domain.ActionDiscard, refs, actor)
	ref := s.promptCardRef("Discard", refs, suggested)
	if s.aborted() {
		return
	}
	if ref == nil {
		ref = suggested
	}
	if ref == nil {
		ref = fallbackCardRef(refs, actor)
	}
	owner := s.Game.PlayerByID(ref.OwnerID)
	if owner == nil || owner.CurrentHand == nil {
		fmt.Println("Discard failed (owner gone).")
		return
	}
	removed, ok := owner.CurrentHand.RemoveCard(ref.ID)
	if !ok {
		fmt.Println("Discard failed (card gone).")
		return
	}
	fmt.Printf("Discard %s from %s\n", removed, owner.Name)
	s.discard(removed)
}

func (s *ManualGameService) executeSwap(actor *domain.Player, action domain.TableCard) {
	refs := s.Game.FaceUpRefs()
	s.discard(action)
	s.prepareAdvisor()
	suggested := s.Advisor.ChooseSwapPair(refs, actor)
	if domain.FirstLegalSwapPair(refs) == nil {
		fmt.Println("Swap: no legal pair. Discarded.")
		return
	}
	pair := s.promptSwap(refs, suggested)
	if s.aborted() {
		return
	}
	for pair == nil || pair.A.OwnerID == pair.B.OwnerID || pair.A.ID == pair.B.ID {
		fmt.Println("Swap needs two face-up cards from different players. Try again.")
		pair = s.promptSwap(refs, suggested)
		if s.aborted() {
			return
		}
	}
	ownerA := s.Game.PlayerByID(pair.A.OwnerID)
	ownerB := s.Game.PlayerByID(pair.B.OwnerID)
	cardA, okA := ownerA.CurrentHand.RemoveCard(pair.A.ID)
	cardB, okB := ownerB.CurrentHand.RemoveCard(pair.B.ID)
	if !okA || !okB {
		if okA {
			s.giveCard(ownerA, cardA)
		}
		if okB {
			s.giveCard(ownerB, cardB)
		}
		fmt.Println("Swap failed.")
		return
	}
	fmt.Printf("Swap %s (%s) with %s (%s)\n", cardA, ownerA.Name, cardB, ownerB.Name)
	s.giveCard(ownerA, cardB)
	s.giveCard(ownerB, cardA)
}

func (s *ManualGameService) giveCard(p *domain.Player, card domain.TableCard) {
	if p.CurrentHand.Status == domain.HandStatusBusted {
		s.discard(card)
		return
	}
	if card.Spec.IsNumberLike() {
		s.receiveNumber(p, card)
		return
	}
	if card.Spec.Type == domain.CardTypeModifier {
		p.CurrentHand.ReceiveModifier(card)
	}
}

func (s *ManualGameService) promptActionPlayer(actor *domain.Player, action domain.ActionType, label string) *domain.Player {
	candidates := s.Game.CurrentRound.NonBustedPlayers()
	s.prepareAdvisor()
	suggested := s.Advisor.ChoosePlayerTarget(action, candidates, actor)
	fmt.Printf("%s. Suggested: %s\n", label, nameOf(suggested))
	target := s.promptPlayer(label+" target", candidates, suggested)
	if s.aborted() {
		return nil
	}
	if target == nil {
		target = suggested
	}
	if target == nil {
		target = actor
	}
	return target
}

func (s *ManualGameService) promptPlayer(title string, candidates []*domain.Player, suggested *domain.Player) *domain.Player {
	if len(candidates) == 0 {
		return nil
	}
	fmt.Println(title + ":")
	for i, c := range candidates {
		mark := ""
		if suggested != nil && c.ID == suggested.ID {
			mark = " [Suggested]"
		}
		fmt.Printf("%d. %s (Total %d) %s%s\n", i+1, c.Name, c.TotalScore, formatHand(c.CurrentHand), mark)
	}
	fmt.Print("Enter choice: ")
	input := s.readLine()
	if s.eof {
		return nil
	}
	if s.handleHistoryCommand(input, true) {
		return nil
	}
	idx, err := strconv.Atoi(input)
	if err != nil || idx < 1 || idx > len(candidates) {
		return suggested
	}
	return candidates[idx-1]
}

func (s *ManualGameService) promptCardRef(title string, refs []domain.CardRef, suggested *domain.CardRef) *domain.CardRef {
	fmt.Println(title + ":")
	for i, r := range refs {
		owner := s.Game.PlayerByID(r.OwnerID)
		mark := ""
		if suggested != nil && r.ID == suggested.ID {
			mark = " [Suggested]"
		}
		fmt.Printf("%d. %s: %s%s\n", i+1, owner.Name, r.Spec.String(), mark)
	}
	fmt.Print("Enter choice: ")
	input := s.readLine()
	if s.eof {
		return nil
	}
	if s.handleHistoryCommand(input, true) {
		return nil
	}
	idx, err := strconv.Atoi(input)
	if err != nil || idx < 1 || idx > len(refs) {
		return suggested
	}
	return &refs[idx-1]
}

func (s *ManualGameService) promptSwap(refs []domain.CardRef, suggested *domain.SwapPair) *domain.SwapPair {
	if suggested != nil {
		a := s.Game.PlayerByID(suggested.A.OwnerID)
		b := s.Game.PlayerByID(suggested.B.OwnerID)
		fmt.Printf("Suggested swap: %s's %s <-> %s's %s\n", a.Name, suggested.A.Spec, b.Name, suggested.B.Spec)
	}
	fmt.Println("Pick first card, then second (different players).")
	first := s.promptCardRef("Swap card A", refs, swapEnd(suggested, true))
	second := s.promptCardRef("Swap card B", refs, swapEnd(suggested, false))
	if first == nil || second == nil {
		return suggested
	}
	return &domain.SwapPair{A: *first, B: *second}
}

func swapEnd(pair *domain.SwapPair, first bool) *domain.CardRef {
	if pair == nil {
		return nil
	}
	if first {
		return &pair.A
	}
	return &pair.B
}

func (s *ManualGameService) readCardFromTable() (domain.TableCard, error) {
	for {
		fmt.Print("Input card: ")
		input := s.readLine()
		if s.eof {
			return domain.TableCard{}, errEOF
		}
		if s.handleHistoryCommand(input, true) {
			if s.rewound {
				return domain.TableCard{}, errRewound
			}
			continue
		}
		switch strings.ToUpper(input) {
		case "HELP", "H", "?":
			fmt.Println(inputHelp)
			continue
		case "HANDS":
			s.printTable()
			continue
		}
		spec, err := ParseManualInput(input)
		if err != nil {
			fmt.Printf("Invalid: %v\n", err)
			continue
		}
		card, err := s.takeFromDeck(spec)
		if err != nil {
			if s.Game.CurrentRound == nil || s.Game.CurrentRound.Deck == nil {
				return domain.TableCard{}, err
			}
			fmt.Printf("Error: %v. Try again.\n", err)
			continue
		}
		return card, nil
	}
}

func (s *ManualGameService) riskDeck() *domain.Deck {
	if s.Game.CurrentRound == nil {
		return nil
	}
	d := s.Game.CurrentRound.Deck
	if d != nil && len(d.Cards) > 0 {
		return d
	}
	if len(s.Game.DiscardPile) > 0 {
		return domain.DeckWithCards(s.Game.DiscardPile)
	}
	return d
}

func (s *ManualGameService) prepareAdvisor() {
	if s.Game.CurrentRound == nil {
		return
	}
	if d, ok := s.Advisor.(interface{ SetDeck(*domain.Deck) }); ok {
		d.SetDeck(s.Game.CurrentRound.Deck)
	}
}

func (s *ManualGameService) others(self *domain.Player) []*domain.Player {
	var out []*domain.Player
	for _, p := range s.Game.Players {
		if p.ID != self.ID {
			out = append(out, p)
		}
	}
	return out
}

func (s *ManualGameService) printTable() {
	fmt.Println("Table:")
	for _, p := range s.Game.Players {
		st := ""
		if p.CurrentHand != nil {
			st = string(p.CurrentHand.Status)
		}
		fmt.Printf("  %s [%s] Total %d | %s\n", p.Name, st, p.TotalScore, formatHand(p.CurrentHand))
	}
}

func (s *ManualGameService) printWinner() {
	fmt.Println("\nGame over.")
	for _, p := range s.Game.Players {
		fmt.Printf("- %s: %d\n", p.Name, p.TotalScore)
	}
}

func (s *ManualGameService) discard(card domain.TableCard) {
	s.Game.DiscardPile = append(s.Game.DiscardPile, card)
}

func (s *ManualGameService) discardAll(cards []domain.TableCard) {
	s.Game.DiscardPile = append(s.Game.DiscardPile, cards...)
}

func (s *ManualGameService) readLine() string {
	line, err := s.Reader.ReadString('\n')
	if err != nil {
		s.eof = true
		if err != io.EOF {
			fmt.Printf("Error reading input: %v\n", err)
		}
	}
	return strings.TrimSpace(line)
}

func (s *ManualGameService) readInt(def int) int {
	line := s.readLine()
	n, err := strconv.Atoi(line)
	if err != nil {
		return def
	}
	return n
}

func formatHand(h *domain.PlayerHand) string {
	if h == nil {
		return "[]"
	}
	parts := make([]string, 0, len(h.NumberLine)+len(h.ModifierLine))
	for _, c := range h.NumberLine {
		parts = append(parts, c.String())
	}
	for _, c := range h.ModifierLine {
		parts = append(parts, c.String())
	}
	return "[" + strings.Join(parts, ", ") + "]"
}

func nameOf(p *domain.Player) string {
	if p == nil {
		return "(none)"
	}
	return p.Name
}

// ParseManualInput maps typed codes to a card spec (identity is assigned when taken from the deck).
func ParseManualInput(input string) (domain.CardSpec, error) {
	in := strings.ToUpper(strings.TrimSpace(input))
	switch in {
	case "0", "Z", "ZERO":
		return domain.NewSpecialCard(domain.SpecialZero).Spec, nil
	case "U", "U7", "UNLUCKY":
		return domain.NewSpecialCard(domain.SpecialUnlucky7).Spec, nil
	case "L", "L13", "LUCKY":
		return domain.NewSpecialCard(domain.SpecialLucky13).Spec, nil
	case "-2":
		return domain.NewModifierCard(domain.ModifierMinus2).Spec, nil
	case "-4":
		return domain.NewModifierCard(domain.ModifierMinus4).Spec, nil
	case "-6":
		return domain.NewModifierCard(domain.ModifierMinus6).Spec, nil
	case "-8":
		return domain.NewModifierCard(domain.ModifierMinus8).Spec, nil
	case "-10":
		return domain.NewModifierCard(domain.ModifierMinus10).Spec, nil
	case "/2", "DIV2", "D2":
		return domain.NewModifierCard(domain.ModifierDivide2).Spec, nil
	case "J", "JOM", "JUST":
		return domain.NewActionCard(domain.ActionJustOneMore).Spec, nil
	case "F4", "F", "FOUR":
		return domain.NewActionCard(domain.ActionFlipFour).Spec, nil
	case "SW", "SWAP":
		return domain.NewActionCard(domain.ActionSwap).Spec, nil
	case "ST", "STEAL":
		return domain.NewActionCard(domain.ActionSteal).Spec, nil
	case "DI", "DISC", "DISCARD":
		return domain.NewActionCard(domain.ActionDiscard).Spec, nil
	}
	n, err := strconv.Atoi(in)
	if err != nil {
		return domain.CardSpec{}, fmt.Errorf("unknown input %q", input)
	}
	if n < 1 || n > 13 {
		return domain.CardSpec{}, fmt.Errorf("number out of range")
	}
	return domain.NewNumberCard(domain.NumberValue(n)).Spec, nil
}
