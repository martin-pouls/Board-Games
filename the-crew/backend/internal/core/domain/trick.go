package domain

// Trick rules that every edition shares (rules §5). They know nothing about seats, the dummy or
// missions: only cards, the led card and which cards are trump. Keep it that way, so the file can
// later move to a shared trick engine unchanged.

// legalCards returns the cards of hand that may be played to a trick whose first card is lead
// (nil when leading). Whoever has a card of the led suit must play one; trump is a suit too. A
// player who cannot follow may play any card. The result is a new slice in hand order, never nil.
func legalCards(hand []Card, lead *Card) []Card {
	if lead != nil {
		var follow []Card
		for _, c := range hand {
			if c.Suit == lead.Suit {
				follow = append(follow, c)
			}
		}
		if len(follow) > 0 {
			return follow
		}
	}
	return cloneCards(hand)
}

// trickWinner returns the index into cards (in play order, cards[0] led) of the winning card:
// the highest trump, or without trump the highest card of the led suit. cards must not be empty.
func trickWinner(cards []Card) int {
	best := 0
	for i := 1; i < len(cards); i++ {
		if beats(cards[i], cards[best]) {
			best = i
		}
	}
	return best
}

// beats reports whether c beats w, the best card so far. w is always of the led suit or trump, so
// a card of any other colour never wins.
func beats(c, w Card) bool {
	switch {
	case c.IsTrump() && !w.IsTrump():
		return true
	case c.Suit == w.Suit:
		return c.Rank > w.Rank
	}
	return false
}
