package domain

// PlayerView is the per-player projection of a game. It never contains other players' hands.
type PlayerView struct {
	GameID                  GameID
	Revision                int64
	Phase                   Phase
	You                     PlayerID
	HostID                  PlayerID
	Options                 Options
	Rules                   RulesView
	Players                 []PlayerSummary // seat order
	Hand                    []HandCard      // ONLY the viewer's hand, sorted ascending; empty in lobby
	Piles                   []PileView
	DrawPileCount           int
	Turn                    *TurnView // nil in lobby / choosingFirstPlayer
	CommunicationRestricted bool
	Result                  *ResultView // nil unless finished
	NextGameID              GameID      // empty until a rematch exists
}

// RulesView is the public subset of the ruleset.
type RulesView struct {
	MinCardsPerTurn, MinCardsDrawPileEmpty int
	HandSize                               int // for the current player count (preview in lobby)
	MinPlayers, MaxPlayers                 int
}

// PlayerSummary is the public information about a seated player.
type PlayerSummary struct {
	ID                      PlayerID
	Name                    string
	Seat                    int
	HandCount               int
	IsHost, IsYou, IsActive bool
	RevealedHand            []Card // sorted; filled for every player only when Phase == finished, else nil
}

// HandCard is a card in the viewer's hand plus where it may legally go now.
type HandCard struct {
	Card       Card
	LegalPiles []int // empty unless viewer is active in playing
}

// PileView is the public state of a pile.
type PileView struct {
	Index      int
	Direction  Direction
	StartValue int
	TopValue   int
	TopCard    *Card // nil when only the start card shows
	CardCount  int   // played cards on the pile
}

// TurnView is the public state of the current turn.
type TurnView struct {
	Number          int
	ActivePlayerID  PlayerID
	CardsPlayed     int
	RequiredMinimum int
	CanEndTurn      bool // CanEndTurn() == nil
	Plays           []Play
}

// ResultView is the public final result.
type ResultView struct {
	Outcome             Outcome
	Score               int
	Excellent           bool // Score < 10
	CardsLeftInHands    int
	CardsLeftInDrawPile int
	StuckPlayerID       PlayerID
}

// GameSummary is public, pre-join information about a game.
type GameSummary struct {
	GameID      GameID
	Phase       Phase
	HostName    string
	PlayerCount int
	MaxPlayers  int
	Joinable    bool // Phase == lobby && PlayerCount < MaxPlayers
}

// ViewFor builds the projection for player id. Errors: ErrUnknownPlayer.
func (g *Game) ViewFor(id PlayerID) (PlayerView, error) {
	panic("not implemented: WP3")
}
