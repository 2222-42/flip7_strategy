package application

import (
	"fmt"

	"flip7_strategy/internal/vengeance/domain"
)

const maxNestedActions = 20
const maxRounds = 500

// GameService orchestrates a Vengeance game.
type GameService struct {
	Game   *domain.Game
	Silent bool
	depth  int
}

func NewGameService(game *domain.Game) *GameService {
	return &GameService{Game: game}
}

func (s *GameService) log(format string, a ...interface{}) {
	if !s.Silent {
		fmt.Printf(format, a...)
	}
}

func (s *GameService) RunGame() {
	if s.Game.Deck == nil {
		s.Game.Deck = domain.NewDeck()
	}

	for !s.Game.IsCompleted {
		s.Game.RoundCount++
		if s.Game.RoundCount > maxRounds {
			s.log("Aborting after %d rounds without a winner.\n", maxRounds)
			s.Game.IsCompleted = true
			break
		}

		dealer := s.Game.Players[s.Game.DealerIndex]
		s.Game.CurrentRound = domain.NewRound(s.Game.Players, dealer, s.Game.Deck)
		s.PlayRound()

		if s.Game.CurrentRound.EndReason == domain.RoundEndReasonAborted {
			s.log("Game aborted (empty deck and discard).\n")
			s.Game.IsCompleted = true
			break
		}

		s.collectTableCards()

		winners := s.Game.DetermineWinners()
		if len(winners) > 0 {
			s.Game.IsCompleted = true
			s.Game.Winners = winners
			break
		}

		s.Game.DealerIndex = (s.Game.DealerIndex + 1) % len(s.Game.Players)
		s.Game.Deck = s.Game.CurrentRound.Deck
	}
}

func (s *GameService) collectTableCards() {
	for _, p := range s.Game.Players {
		if p.CurrentHand == nil {
			continue
		}
		s.Game.DiscardPile = append(s.Game.DiscardPile, p.CurrentHand.NumberLine...)
		s.Game.DiscardPile = append(s.Game.DiscardPile, p.CurrentHand.ModifierLine...)
	}
}

func (s *GameService) ensureDeck() bool {
	round := s.Game.CurrentRound
	if round.Deck != nil && len(round.Deck.Cards) > 0 {
		return true
	}
	if len(s.Game.DiscardPile) == 0 {
		return false
	}
	s.log("Deck empty. Reshuffling %d cards from discard pile...\n", len(s.Game.DiscardPile))
	round.Deck = domain.NewDeckFromCards(s.Game.DiscardPile)
	s.Game.DiscardPile = nil
	return len(round.Deck.Cards) > 0
}

func (s *GameService) DrawCard() (domain.TableCard, error) {
	if !s.ensureDeck() {
		return domain.TableCard{}, fmt.Errorf("deck is empty and discard pile is empty")
	}
	return s.Game.CurrentRound.Deck.Draw()
}

func (s *GameService) PlayRound() {
	round := s.Game.CurrentRound
	s.log("--- New Round %d! Dealer: %s ---\n", s.Game.RoundCount, round.Dealer.Name)

	for _, p := range round.OfferOrder {
		if round.IsEnded {
			s.scoreRound()
			return
		}
		if p.CurrentHand.Status != domain.HandStatusActive {
			continue
		}
		card, err := s.DrawCard()
		if err != nil {
			s.log("%s\n", err.Error())
			round.End(domain.RoundEndReasonAborted)
			return
		}
		s.log("%s dealt: %v\n", p.Name, card)
		s.ProcessFlip(p, card)
	}

	if round.IsEnded {
		s.scoreRound()
		return
	}

	for !round.IsEnded {
		active := round.ActivePlayers()
		if len(active) == 0 {
			round.End(domain.RoundEndReasonNoActivePlayers)
			break
		}

		progress := false
		for _, p := range round.OfferOrder {
			if round.IsEnded {
				break
			}
			if p.CurrentHand.Status != domain.HandStatusActive {
				continue
			}
			progress = true
			s.ensureDeck()

			ctx := domain.DecisionContext{
				Deck:         round.Deck,
				DiscardPile:  s.Game.DiscardPile,
				Hand:         p.CurrentHand,
				PlayerScore:  p.TotalScore,
				OtherPlayers: s.others(p),
				Rules:        s.Game.Rules,
			}
			choice := domain.TurnChoiceStay
			if p.Strategy != nil {
				s.prepareStrategy(p)
				choice = p.Strategy.Decide(ctx)
			}
			if p.CurrentHand.MustHit() {
				choice = domain.TurnChoiceHit
			}
			s.log("%s decides to %s\n", p.Name, choice)

			if choice == domain.TurnChoiceStay && p.CurrentHand.CanStay() {
				p.CurrentHand.Stay()
				s.log("%s stays.\n", p.Name)
				continue
			}

			card, err := s.DrawCard()
			if err != nil {
				s.log("%s\n", err.Error())
				round.End(domain.RoundEndReasonAborted)
				break
			}
			s.log("%s drew: %v\n", p.Name, card)
			s.ProcessFlip(p, card)
		}
		if !progress && !round.IsEnded {
			round.End(domain.RoundEndReasonNoActivePlayers)
		}
	}

	s.scoreRound()
}

func (s *GameService) scoreRound() {
	if s.Game.CurrentRound.EndReason == domain.RoundEndReasonAborted {
		return
	}
	calc := domain.NewScoreCalculatorFor(s.Game.Rules)
	var flip7 *domain.Player
	for _, p := range s.Game.Players {
		if p.CurrentHand == nil {
			continue
		}
		if p.CurrentHand.Status == domain.HandStatusBusted && !s.Game.Rules.ModifiersTargetBusted {
			s.log("%s scores 0 (busted). Total: %d\n", p.Name, p.TotalScore)
			continue
		}
		pv := calc.Compute(p.CurrentHand)
		amt := pv.Total
		if p.CurrentHand.HasFlip7() && p.CurrentHand.Status != domain.HandStatusBusted {
			flip7 = p
			if s.Game.Rules.Flip7AsAttack {
				amt -= pv.Bonus
			}
		}
		p.BankScore(amt)
		s.log("%s scores %d. Total: %d\n", p.Name, amt, p.TotalScore)
	}
	if flip7 == nil || !s.Game.Rules.Flip7AsAttack {
		return
	}
	choice := domain.Flip7BonusChoice{}
	if flip7.Strategy != nil {
		s.prepareStrategy(flip7)
		choice = flip7.Strategy.ChooseFlip7Bonus(flip7, s.others(flip7))
	}
	if choice.SubtractFrom != nil && choice.SubtractFrom.ID != flip7.ID {
		choice.SubtractFrom.BankScore(-domain.Flip7Bonus)
		s.log("%s Flip 7: −%d to %s. Totals %d / %d\n", flip7.Name, domain.Flip7Bonus, choice.SubtractFrom.Name, flip7.TotalScore, choice.SubtractFrom.TotalScore)
		return
	}
	flip7.BankScore(domain.Flip7Bonus)
	s.log("%s Flip 7: takes +%d. Total: %d\n", flip7.Name, domain.Flip7Bonus, flip7.TotalScore)
}

func (s *GameService) others(self *domain.Player) []*domain.Player {
	var out []*domain.Player
	for _, p := range s.Game.Players {
		if p.ID != self.ID {
			out = append(out, p)
		}
	}
	return out
}

// ProcessFlip handles a card revealed in front of actor (deal, hit, JOM, Flip Four number).
func (s *GameService) ProcessFlip(actor *domain.Player, card domain.TableCard) {
	round := s.Game.CurrentRound
	if round.IsEnded {
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
}

func (s *GameService) receiveNumber(target *domain.Player, card domain.TableCard) {
	res := target.CurrentHand.ReceiveNumberLike(card)
	if len(res.Wiped) > 0 {
		s.Game.DiscardPile = append(s.Game.DiscardPile, res.Wiped...)
		s.log("%s received Unlucky 7 and wiped %d cards.\n", target.Name, len(res.Wiped))
	}
	if res.Busted {
		s.log("%s BUSTED!\n", target.Name)
		return
	}
	if res.Flip7 {
		s.log("%s FLIP 7!\n", target.Name)
		s.Game.CurrentRound.End(domain.RoundEndReasonFlip7)
	}
}

func (s *GameService) assignModifier(actor *domain.Player, card domain.TableCard) {
	candidates := s.modifierCandidates()
	if len(candidates) == 0 {
		s.discard(card)
		return
	}
	target := actor
	if actor.Strategy != nil {
		s.prepareStrategy(actor)
		if t := actor.Strategy.ChooseModifierTarget(card.Spec.ModifierType, candidates, actor); t != nil {
			target = t
		}
	}
	if !s.validModifierTarget(target) {
		target = actor
		if !s.validModifierTarget(target) && len(candidates) > 0 {
			target = candidates[0]
		}
	}
	s.log("%s assigns %s to %s\n", actor.Name, card, target.Name)
	target.CurrentHand.ReceiveModifier(card)
}

func (s *GameService) modifierCandidates() []*domain.Player {
	if s.Game.Rules.ModifiersTargetBusted {
		var out []*domain.Player
		for _, p := range s.Game.Players {
			if p.CurrentHand != nil {
				out = append(out, p)
			}
		}
		return out
	}
	if s.Game.CurrentRound == nil {
		return nil
	}
	return s.Game.CurrentRound.NonBustedPlayers()
}

func (s *GameService) validModifierTarget(p *domain.Player) bool {
	if p == nil || p.CurrentHand == nil {
		return false
	}
	if s.Game.Rules.ModifiersTargetBusted {
		return true
	}
	return p.CurrentHand.Status != domain.HandStatusBusted
}

func (s *GameService) resolveAction(actor *domain.Player, card domain.TableCard) {
	if s.depth >= maxNestedActions {
		s.log("Nested action limit reached; discarding %s\n", card)
		s.discard(card)
		return
	}
	s.depth++
	defer func() { s.depth-- }()

	switch card.Spec.ActionType {
	case domain.ActionJustOneMore:
		target := s.choosePlayer(actor, domain.ActionJustOneMore)
		s.discard(card)
		s.log("%s plays Just One More on %s\n", actor.Name, target.Name)
		s.executeJustOneMore(target)
	case domain.ActionFlipFour:
		target := s.choosePlayer(actor, domain.ActionFlipFour)
		s.discard(card)
		s.log("%s plays Flip Four on %s\n", actor.Name, target.Name)
		s.executeFlipFour(target)
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

func (s *GameService) executeJustOneMore(target *domain.Player) {
	if s.Game.CurrentRound.IsEnded || !s.isNonBusted(target) {
		return
	}
	card, err := s.DrawCard()
	if err != nil {
		s.Game.CurrentRound.End(domain.RoundEndReasonAborted)
		return
	}
	s.log("%s (Just One More) flips: %v\n", target.Name, card)
	s.ProcessFlip(target, card)
	if s.Game.CurrentRound.IsEnded {
		return
	}
	if target.CurrentHand.Status == domain.HandStatusActive {
		target.CurrentHand.Stay()
		s.log("%s is forced to stay.\n", target.Name)
	}
}

func (s *GameService) executeFlipFour(target *domain.Player) {
	round := s.Game.CurrentRound
	if round.IsEnded || !s.isNonBusted(target) {
		return
	}

	delayed := make([]domain.TableCard, 0, domain.FlipFourCardCount)
	for i := 0; i < domain.FlipFourCardCount; i++ {
		if round.IsEnded || target.CurrentHand.Status == domain.HandStatusBusted {
			break
		}
		card, err := s.DrawCard()
		if err != nil {
			round.End(domain.RoundEndReasonAborted)
			s.discardAll(delayed)
			return
		}
		s.log("%s (Flip Four %d/4) flips: %v\n", target.Name, i+1, card)
		if card.Spec.IsNumberLike() {
			s.receiveNumber(target, card)
			continue
		}
		delayed = append(delayed, card)
	}

	if round.IsEnded || target.CurrentHand.Status == domain.HandStatusBusted {
		s.discardAll(delayed)
		return
	}
	for _, card := range delayed {
		if round.IsEnded {
			s.discard(card)
			continue
		}
		s.ProcessFlip(target, card)
	}
}

func (s *GameService) executeSteal(actor *domain.Player, action domain.TableCard) {
	refs := s.Game.FaceUpRefs()
	s.discard(action)
	if len(refs) == 0 {
		s.log("%s plays Steal with no target; discarded.\n", actor.Name)
		return
	}
	var ref *domain.CardRef
	if actor.Strategy != nil {
		s.prepareStrategy(actor)
		ref = actor.Strategy.ChooseCardTarget(domain.ActionSteal, refs, actor)
	} else {
		ref = &refs[0]
	}
	if ref == nil {
		ref = fallbackCardRef(refs, actor)
	}
	owner := s.Game.PlayerByID(ref.OwnerID)
	if owner == nil || owner.CurrentHand == nil {
		return
	}
	stolen, ok := owner.CurrentHand.RemoveCard(ref.ID)
	if !ok {
		return
	}
	s.log("%s steals %s from %s\n", actor.Name, stolen, owner.Name)
	if stolen.Spec.IsNumberLike() {
		s.receiveNumber(actor, stolen)
		return
	}
	if stolen.Spec.Type == domain.CardTypeModifier {
		actor.CurrentHand.ReceiveModifier(stolen)
	}
}

func (s *GameService) executeDiscard(actor *domain.Player, action domain.TableCard) {
	refs := s.Game.FaceUpRefs()
	s.discard(action)
	if len(refs) == 0 {
		s.log("%s plays Discard with no target; discarded.\n", actor.Name)
		return
	}
	var ref *domain.CardRef
	if actor.Strategy != nil {
		s.prepareStrategy(actor)
		ref = actor.Strategy.ChooseCardTarget(domain.ActionDiscard, refs, actor)
	} else {
		ref = &refs[0]
	}
	if ref == nil {
		ref = fallbackCardRef(refs, actor)
	}
	owner := s.Game.PlayerByID(ref.OwnerID)
	if owner == nil || owner.CurrentHand == nil {
		return
	}
	removed, ok := owner.CurrentHand.RemoveCard(ref.ID)
	if !ok {
		return
	}
	s.log("%s discards %s from %s\n", actor.Name, removed, owner.Name)
	s.discard(removed)
}

func (s *GameService) executeSwap(actor *domain.Player, action domain.TableCard) {
	refs := s.Game.FaceUpRefs()
	s.discard(action)
	var pair *domain.SwapPair
	if actor.Strategy != nil {
		s.prepareStrategy(actor)
		pair = actor.Strategy.ChooseSwapPair(refs, actor)
	}
	if pair == nil || pair.A.OwnerID == pair.B.OwnerID || pair.A.ID == pair.B.ID {
		pair = domain.FirstLegalSwapPair(refs)
	}
	if pair == nil || pair.A.OwnerID == pair.B.OwnerID || pair.A.ID == pair.B.ID {
		s.log("%s plays Swap with no legal pair; discarded.\n", actor.Name)
		return
	}
	ownerA := s.Game.PlayerByID(pair.A.OwnerID)
	ownerB := s.Game.PlayerByID(pair.B.OwnerID)
	if ownerA == nil || ownerB == nil {
		return
	}
	cardA, okA := ownerA.CurrentHand.RemoveCard(pair.A.ID)
	cardB, okB := ownerB.CurrentHand.RemoveCard(pair.B.ID)
	if !okA || !okB {
		if okA {
			s.returnCard(ownerA, cardA)
		}
		if okB {
			s.returnCard(ownerB, cardB)
		}
		return
	}
	s.log("%s swaps %s (%s) with %s (%s)\n", actor.Name, cardA, ownerA.Name, cardB, ownerB.Name)
	s.giveCard(ownerA, cardB)
	if s.Game.CurrentRound.IsEnded {
		s.giveCard(ownerB, cardA)
		return
	}
	s.giveCard(ownerB, cardA)
}

func (s *GameService) giveCard(p *domain.Player, card domain.TableCard) {
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

func (s *GameService) returnCard(p *domain.Player, card domain.TableCard) {
	s.giveCard(p, card)
}

func (s *GameService) choosePlayer(actor *domain.Player, action domain.ActionType) *domain.Player {
	candidates := s.Game.CurrentRound.NonBustedPlayers()
	if len(candidates) == 0 {
		return actor
	}
	if actor.Strategy == nil {
		return candidates[0]
	}
	s.prepareStrategy(actor)
	t := actor.Strategy.ChoosePlayerTarget(action, candidates, actor)
	if t == nil || !s.isNonBusted(t) {
		return actor
	}
	return t
}

func fallbackCardRef(refs []domain.CardRef, self *domain.Player) *domain.CardRef {
	if self != nil {
		for i := range refs {
			if refs[i].OwnerID != self.ID {
				return &refs[i]
			}
		}
	}
	return &refs[0]
}

func (s *GameService) prepareStrategy(p *domain.Player) {
	if p == nil || p.Strategy == nil || s.Game.CurrentRound == nil {
		return
	}
	if d, ok := p.Strategy.(interface{ SetDeck(*domain.Deck) }); ok {
		d.SetDeck(s.Game.CurrentRound.Deck)
	}
	if r, ok := p.Strategy.(interface{ SetRules(domain.GameRules) }); ok {
		r.SetRules(s.Game.Rules)
	}
}

func (s *GameService) isNonBusted(p *domain.Player) bool {
	return p != nil && p.CurrentHand != nil && p.CurrentHand.Status != domain.HandStatusBusted
}

func (s *GameService) discard(card domain.TableCard) {
	s.Game.DiscardPile = append(s.Game.DiscardPile, card)
}

func (s *GameService) discardAll(cards []domain.TableCard) {
	s.Game.DiscardPile = append(s.Game.DiscardPile, cards...)
}
