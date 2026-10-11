package domain

import (
	"fmt"
	"slices"
	"time"
)

// PlayerView is the projection of a game for one player. Before the game ends it never exposes
// the teammate's hand or gift choice, nor the deck.
type PlayerView struct {
	GameID      GameID
	Revision    int64
	Phase       Phase
	You, HostID PlayerID
	Level       int             // 0 in the lobby
	Players     []PlayerSummary // seat order
	Hand        []Card          // only the viewer's hand, sorted; [] in the lobby
	Playable    []Card          // the viewer's legal cards when they lead or follow now, else []
	Board       *BoardView      // nil in the lobby
	Round       int
	Decree      *Card
	Trick       *TrickView       // the trick in progress; nil outside a trick
	LastTrick   *TrickRecordView // the last completed trick until the next card is led
	Pending     *PendingView     // nil unless playing
	LastGift    *GiftView        // the viewer's side of the last gift exchange until the next lead
	Result      *ResultView
	NextGameID  GameID
}

// PlayerSummary is the public information about a seated player.
type PlayerSummary struct {
	ID            PlayerID
	Name          string
	Seat          int
	End           End // the end of the path the player faces
	HandCount     int
	IsHost, IsYou bool
	IsDealer      bool   // this round's dealer (false in the lobby)
	MustAct       bool   // the game waits for this player
	RevealedHand  []Card // the sorted hand once the game is over; nil before
}

// BoardView is the forest board.
type BoardView struct {
	Side              BoardSide
	Start             int
	Spaces            []SpaceView // from the left end to the right end
	Traveller         int
	ForestTokens      int // left to place
	ForestTokensTotal int
	GemsLeft          int
}

// SpaceView is one path space with its location.
type SpaceView struct {
	Gems    int
	Plus    bool
	Covered bool
}

// TrickView is the trick in progress.
type TrickView struct {
	Number   int
	LeaderID PlayerID
	Lead     *Card
	Follow   *Card
	WinnerID PlayerID // once both cards are on the table
}

// TrickRecordView is the last completed trick.
type TrickRecordView struct {
	Number       int
	LeaderID     PlayerID
	Lead, Follow Card
	WinnerID     PlayerID
	From, To     int
	Delta        int
	Beyond       End
	Collected    bool
}

// PendingView is the decision the game waits for.
type PendingView struct {
	Step        Step
	PlayerIDs   []PlayerID   // who must act
	ChosenID    PlayerID     // foxesSwap: the player who may swap
	MoveOptions []MoveOption // move: the distinct results, the normal move first
	Ends        []End        // forestEnd: the ends the token may cover
	YourGift    *Card        // gift: the card the viewer has chosen
}

// GiftView is the viewer's side of the last gift exchange.
type GiftView struct {
	Gave, Received Card
}

// ResultView describes a finished game.
type ResultView struct {
	Outcome  Phase // won | lost
	Reason   LossReason
	PlayerID PlayerID
	Round    int
	GemsLeft int
	Score    *Score // only on a win
}

// GameSummary is the public, pre-join information about a game.
type GameSummary struct {
	GameID      GameID
	Phase       Phase
	HostName    string
	PlayerCount int
	MaxPlayers  int
	Joinable    bool // lobby && PlayerCount < MaxPlayers
	CreatedAt   time.Time
}

// ViewFor builds the projection for player id. Errors: ErrUnknownPlayer.
func (g *Game) ViewFor(id PlayerID) (PlayerView, error) {
	me, seat, ok := g.PlayerByID(id)
	if !ok {
		return PlayerView{}, fmt.Errorf("%w: %s", ErrUnknownPlayer, id)
	}
	over := g.Phase == PhaseWon || g.Phase == PhaseLost
	started := g.Phase != PhaseLobby && g.Phase != PhaseEnded
	actors := g.Actors()
	v := PlayerView{
		GameID:     g.ID,
		Revision:   g.Revision,
		Phase:      g.Phase,
		You:        id,
		HostID:     g.HostID,
		Level:      g.Level,
		Players:    make([]PlayerSummary, len(g.Players)),
		Hand:       cloneCards(me.Hand),
		Playable:   g.LegalCards(seat),
		Round:      g.Round,
		Decree:     cloneCard(g.Decree),
		NextGameID: g.NextGameID,
	}
	for i, p := range g.Players {
		ps := PlayerSummary{
			ID:        p.ID,
			Name:      p.Name,
			Seat:      i,
			End:       endOf(i),
			HandCount: len(p.Hand),
			IsHost:    p.ID == g.HostID,
			IsYou:     p.ID == id,
			IsDealer:  started && i == g.Dealer,
			MustAct:   slices.Contains(actors, i),
		}
		if over {
			ps.RevealedHand = cloneCards(p.Hand)
		}
		v.Players[i] = ps
	}
	if started {
		v.Board = g.boardView()
	}
	if g.Phase == PhasePlaying && g.Trick.Number > 0 {
		t := &TrickView{
			Number:   g.Trick.Number,
			LeaderID: g.Players[g.Trick.Leader].ID,
			Lead:     cloneCard(g.Trick.Lead),
			Follow:   cloneCard(g.Trick.Follow),
		}
		if t.Lead != nil && t.Follow != nil {
			t.WinnerID = g.Players[g.TrickWinner()].ID
		}
		v.Trick = t
	}
	if r := g.LastTrick; r != nil {
		v.LastTrick = &TrickRecordView{
			Number: r.Number, LeaderID: g.Players[r.Leader].ID, Lead: r.Lead, Follow: r.Follow,
			WinnerID: g.Players[r.Winner].ID, From: r.From, To: r.To, Delta: r.Delta,
			Beyond: r.Beyond, Collected: r.Collected,
		}
	}
	if g.LastGift != nil {
		v.LastGift = &GiftView{Gave: g.LastGift.Given[seat], Received: g.LastGift.Given[1-seat]}
	}
	if g.Phase == PhasePlaying {
		v.Pending = g.pendingView(seat, actors)
	}
	if over && g.Result != nil {
		rv := &ResultView{
			Outcome:  g.Phase,
			Reason:   g.Result.Reason,
			PlayerID: g.Result.PlayerID,
			Round:    g.Round,
			GemsLeft: g.Board.GemsLeft(),
		}
		if g.Phase == PhaseWon {
			s := g.Score()
			rv.Score = &s
		}
		v.Result = rv
	}
	return v, nil
}

// boardView builds the board projection.
func (g *Game) boardView() *BoardView {
	l := g.LevelDef()
	b := &g.Board
	bv := &BoardView{
		Side:              l.Side,
		Start:             l.Start(),
		Spaces:            make([]SpaceView, len(b.Gems)),
		Traveller:         b.Traveller,
		ForestTokens:      b.ForestTokens,
		ForestTokensTotal: l.ForestTokens,
		GemsLeft:          b.GemsLeft(),
	}
	for i, s := range l.Spaces() {
		bv.Spaces[i] = SpaceView{Gems: b.Gems[i], Plus: s.Plus, Covered: b.Covered(i)}
	}
	return bv
}

// pendingView builds the pending decision for the viewer in seat.
func (g *Game) pendingView(seat int, actors []int) *PendingView {
	pv := &PendingView{Step: g.Step, PlayerIDs: make([]PlayerID, len(actors))}
	for i, a := range actors {
		pv.PlayerIDs[i] = g.Players[a].ID
	}
	switch g.Step {
	case StepFoxesSwap:
		pv.ChosenID = g.Players[g.Trick.Chosen].ID
	case StepMove:
		pv.MoveOptions = g.MoveOptions()
	case StepForestEnd:
		pv.Ends = g.CoverableEnds()
	case StepGift:
		pv.YourGift = cloneCard(g.Trick.Gift[seat])
	}
	return pv
}
