package domain

import (
	"fmt"
	"slices"
	"time"
)

// PlayerView is the projection of a game for one player (plan §5.5). Before the game ends it
// never exposes the partner's unplaced dice, reroll selection or Working Together slot; it
// never exposes future rolls.
type PlayerView struct {
	GameID      GameID
	Revision    int64
	Phase       Phase
	You, HostID PlayerID
	NextGameID  GameID
	Players     []PlayerSummary // seat order
	Scenario    ScenarioInfo
	Abilities   []AbilityID    // the chosen cards; never nil
	Catalogue   []ScenarioInfo // lobby only: every scenario in §9 order; nil otherwise
	Flight      *FlightView    // nil in the lobby (and after a lobby was ended)
	Result      *ResultView    // nil unless won or lost
}

// PlayerSummary is the public information about a seated player.
type PlayerSummary struct {
	ID            PlayerID
	Name          string
	Seat          int
	Role          Role
	IsHost, IsYou bool
	MustAct       bool  // the game waits for this player
	DiceLeft      int   // unplaced dice this round
	Ready         bool  // pressed Ready in this round's discussion
	RevealedDice  []int // unplaced dice in slot order once the game is over; nil before
}

// ScenarioInfo is the public description of a scenario. Spaces is the approach track's length.
type ScenarioInfo struct {
	ID              ScenarioID
	Airport, City   string
	Colour          ScenarioColour
	Modules         []Module // never nil
	Abilities       int
	RealTimeDropped bool
	Spaces          int
}

// FlightView is the cockpit as one player sees it.
type FlightView struct {
	Round        int
	FirstPlayer  Role
	Step         Step
	Turn         Role
	Approach     ApproachView
	RerollRounds []int // altitude rounds still holding a reroll token
	Axis         int
	AeroBlue     int
	AeroOrange   int
	Gear         [3]bool
	Flaps        int
	Brakes       int
	BrakeLimit   int
	Coffee       int
	Rerolls      int
	PlaneSupply  int
	Fuel         *int        // nil without a kerosene track
	Leak         bool        // kerosene leak (§12.5)
	Wind         *WindView   // nil without wind
	Ice          *IceView    // nil without ice brakes
	Intern       *InternView // nil without the intern
	Placements   []PlacementView
	Speed        *SpeedView // once the second engine piece is placed this round
	RoundStart   RoundStart
	YourDice     [HandSize]int // the viewer's hand by slot; 0 = empty
	AbilityState AbilityStateView
	Can          CanView
	Legal        *LegalView  // the viewer's turn in StepPlace only
	Reroll       *RerollView // StepReroll only
	Swap         *SwapView   // StepSwap only
	Sync         *SyncView   // StepSync only
}

// ApproachView is the approach track with its tokens. Position is the plane's space (1..N).
type ApproachView struct {
	Spaces   []ApproachSpaceView
	Position int
}

// ApproachSpaceView is one approach space; Turn is the mask ("" = no icon).
type ApproachSpaceView struct {
	Planes, Traffic int
	Turn            string
}

// WindView is the wind ring: the token's offset from the top and the value it points at.
type WindView struct{ Offset, Value int }

// IceView is the ice-brake marker slot (0..4).
type IceView struct{ Marker int }

// InternView is the intern token row; Next* is the index of each role's next token, nil when
// the row is used up.
type InternView struct {
	Tokens                 []InternTokenView
	NextPilot, NextCopilot *int
}

// InternTokenView is one intern token.
type InternTokenView struct {
	Value int
	Taken bool
}

// PlacementView is a placed piece. Role is who placed it (the Co-Pilot for the traffic die).
type PlacementView struct {
	Space  SpaceID
	Piece  PieceKind
	Value  int
	Coffee int
	Role   Role
}

// SpeedView is the speed of this round. Move is nil in the final round (no movement, §6.2).
type SpeedView struct {
	Sum, Wind, Total int
	Move             *int
}

// AbilityStateView says which abilities were used; the results stay private.
type AbilityStateView struct {
	AdaptationUsed                     [2]bool // by role
	AnticipationUsed, TogetherUsed     bool
	SyncUsed                           bool
	ControlTriggered, MasteryTriggered bool
}

// CanView lists the side intents the viewer may send now (§13, §5.3).
type CanView struct {
	SpendReroll, Adapt, Anticipate, OfferSwap bool
}

// LegalView is the acting viewer's options: ByValue[v-1] are the spaces a die of final value v
// may go on; InternToken is the value of the viewer's next intern token and InternTargets the
// spaces it may go on.
type LegalView struct {
	ByValue       [6][]SpaceID
	InternToken   *int
	InternTargets []SpaceID
}

// RerollView is the reroll step: who has confirmed (the selection stays private).
type RerollView struct{ Confirmed [2]bool }

// SwapView is the Working Together card: the offered value is public, the initiator's slot is
// shown to the initiator only.
type SwapView struct {
	By       Role
	Value    int
	YourSlot *int
}

// SyncView is the Synchronisation die; Targets are shown to the Co-Pilot only.
type SyncView struct {
	Value   int
	Targets []SpaceID
}

// ResultView describes a finished game.
type ResultView struct {
	Won      bool
	Reason   LossReason
	Failed   []Condition // never nil
	PlayerID PlayerID    // who gave up
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

// ScenarioInfoFor builds the public description of a scenario.
func ScenarioInfoFor(s ScenarioDef) ScenarioInfo {
	t, _ := TrackFor(s.Track)
	return ScenarioInfo{
		ID:              s.ID,
		Airport:         s.Airport,
		City:            s.City,
		Colour:          s.Colour,
		Modules:         append([]Module{}, s.Modules...),
		Abilities:       s.Abilities,
		RealTimeDropped: s.RealTimeDropped,
		Spaces:          t.Len(),
	}
}

// ViewFor builds the projection for player id. Errors: ErrUnknownPlayer.
func (g *Game) ViewFor(id PlayerID) (PlayerView, error) {
	_, seat, ok := g.PlayerByID(id)
	if !ok {
		return PlayerView{}, fmt.Errorf("%w: %s", ErrUnknownPlayer, id)
	}
	over := g.Phase == PhaseWon || g.Phase == PhaseLost
	actors := g.Actors()
	v := PlayerView{
		GameID:     g.ID,
		Revision:   g.Revision,
		Phase:      g.Phase,
		You:        id,
		HostID:     g.HostID,
		NextGameID: g.NextGameID,
		Players:    make([]PlayerSummary, len(g.Players)),
		Scenario:   ScenarioInfoFor(g.ScenarioDef()),
		Abilities:  append([]AbilityID{}, g.Abilities...),
	}
	for i, p := range g.Players {
		role := g.RoleOf(i)
		ps := PlayerSummary{
			ID:      p.ID,
			Name:    p.Name,
			Seat:    i,
			Role:    role,
			IsHost:  p.ID == g.HostID,
			IsYou:   p.ID == id,
			MustAct: slices.Contains(actors, role),
		}
		if g.Flight != nil {
			ps.DiceLeft = g.Flight.R.diceLeft(role)
			ps.Ready = g.Flight.R.Ready[role]
			if over {
				ps.RevealedDice = g.Flight.R.handValues(role)
			}
		}
		v.Players[i] = ps
	}
	if g.Phase == PhaseLobby {
		for _, s := range scenarios {
			v.Catalogue = append(v.Catalogue, ScenarioInfoFor(s))
		}
	}
	if g.Flight != nil {
		v.Flight = g.flightView(g.RoleOf(seat))
	}
	if over && g.Result != nil {
		v.Result = &ResultView{
			Won:      g.Phase == PhaseWon,
			Reason:   g.Result.Reason,
			Failed:   append([]Condition{}, g.Result.Failed...),
			PlayerID: g.Result.PlayerID,
		}
	}
	return v, nil
}

// flightView builds the cockpit for the viewer in role me.
func (g *Game) flightView(me Role) *FlightView {
	f := g.Flight
	r := &f.R
	sc := g.ScenarioDef()
	t := g.track()
	fv := &FlightView{
		Round:        f.Round,
		FirstPlayer:  g.FirstPlayer(),
		Step:         r.Step,
		Turn:         r.Turn,
		Approach:     ApproachView{Spaces: make([]ApproachSpaceView, len(f.Planes)), Position: f.Position},
		RerollRounds: g.RerollRounds(),
		Axis:         f.Axis,
		AeroBlue:     f.AeroBlue(),
		AeroOrange:   f.AeroOrange(),
		Gear:         f.Gear,
		Flaps:        f.Flaps,
		Brakes:       f.Brakes,
		BrakeLimit:   g.BrakeLimit(),
		Coffee:       f.Coffee,
		Rerolls:      f.Rerolls,
		PlaneSupply:  f.PlaneSupply(),
		Leak:         sc.Has(ModuleKeroseneLeak),
		Placements:   make([]PlacementView, len(r.Placements)),
		RoundStart:   RoundStart{RerollPicked: r.Start.RerollPicked, Traffic: append([]TrafficRoll{}, r.Start.Traffic...)},
		YourDice:     r.Dice[me],
		AbilityState: AbilityStateView{
			AdaptationUsed:   f.AdaptationUsed,
			AnticipationUsed: r.AnticipationUsed,
			TogetherUsed:     r.TogetherUsed,
			SyncUsed:         r.SyncUsed,
		},
	}
	for i, n := range f.Planes {
		s := ApproachSpaceView{Planes: n}
		if i < t.Len() {
			s.Traffic = t.Spaces[i].Traffic
			s.Turn = t.Spaces[i].Turn.String()
		}
		fv.Approach.Spaces[i] = s
	}
	if sc.hasKerosene() {
		fuel := f.Fuel
		fv.Fuel = &fuel
	}
	if sc.Has(ModuleWind) {
		fv.Wind = &WindView{Offset: f.Wind, Value: WindValue(f.Wind)}
	}
	if sc.Has(ModuleIceBrakes) {
		fv.Ice = &IceView{Marker: f.Ice}
	}
	if f.Intern != nil {
		iv := &InternView{Tokens: make([]InternTokenView, len(f.Intern))}
		for i, v := range f.Intern {
			iv.Tokens[i] = InternTokenView{Value: v, Taken: f.internTaken(i)}
		}
		if i, ok := f.internNext(RolePilot); ok {
			iv.NextPilot = &i
		}
		if i, ok := f.internNext(RoleCopilot); ok {
			iv.NextCopilot = &i
		}
		fv.Intern = iv
	}
	for i, p := range r.Placements {
		fv.Placements[i] = PlacementView{Space: p.Space, Piece: p.Piece, Value: p.Value, Coffee: p.Coffee, Role: p.Role}
	}
	if s := r.Speed; s != nil {
		sv := &SpeedView{Sum: s.Sum, Wind: s.Wind, Total: s.Total()}
		if f.Round < Rounds {
			move := s.Move
			sv.Move = &move
		}
		fv.Speed = sv
	}
	if g.Phase != PhasePlaying {
		return fv
	}
	switch r.Step {
	case StepReroll:
		fv.Reroll = &RerollView{Confirmed: r.RerollConfirmed}
	case StepSwap:
		if s := r.Swap; s != nil {
			sv := &SwapView{By: s.By, Value: r.Dice[s.By][s.Slot]}
			if s.By == me {
				slot := s.Slot
				sv.YourSlot = &slot
			}
			fv.Swap = sv
		}
	case StepSync:
		sv := &SyncView{Value: r.SyncValue}
		if me == RoleCopilot {
			sv.Targets = g.legalSpaces(RoleCopilot, PieceTraffic, r.SyncValue)
		}
		fv.Sync = sv
	}
	// Legal (the place step), Can and the Control/Mastery triggers come with the base round and
	// the abilities.
	return fv
}
