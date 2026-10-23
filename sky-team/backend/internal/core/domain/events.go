package domain

// Event is a domain event. Events carry only public information (plan §5.5): placed pieces,
// traffic and Synchronisation rolls carry values; private moves (the roll, rerolls, Adaptation,
// Anticipation, the Working Together answer, discards) never carry die values. The offered
// Working Together die is public. Events are emitted as values; their slices are never shared
// with state.
type Event interface{ isEvent() }

type PlayerJoined struct {
	PlayerID PlayerID
	Name     string
}

// PlayerLeft: HostID is the host after the leave.
type PlayerLeft struct{ PlayerID, HostID PlayerID }

// FlightConfigured: the host picked the scenario and the ability cards (§4.1, §13).
type FlightConfigured struct {
	Scenario  ScenarioID
	Abilities []AbilityID
}

// RolesSwapped: PilotSeat is the Pilot's seat after the swap.
type RolesSwapped struct{ PilotSeat int }

type GameStarted struct{ Scenario ScenarioID }

// RoundStarted: the reroll pickup and the traffic rolls of the new round (§5.1).
type RoundStarted struct {
	Round        int
	RerollPicked bool
	Traffic      []TrafficRoll
}

type PlayerReady struct{ Role Role }

// DiceRolled: both hands were rolled (values stay private).
type DiceRolled struct{}

// DiePlaced: a piece went on Space with its final Value and the Coffee spent. Role is who placed it.
type DiePlaced struct {
	Role   Role
	Piece  PieceKind
	Space  SpaceID
	Value  int
	Coffee int
}

// InternTrained: Role took the intern Token (its value) and put it on Space (§12.2).
type InternTrained struct {
	Role  Role
	Token int
	Space SpaceID
}

type AxisMoved struct{ From, To int }

// WindTurned: the wind token now points Offset spaces from the top, at Value (§12.3).
type WindTurned struct{ Offset, Value int }

// PlaneMoved: the plane advanced on the approach track.
type PlaneMoved struct{ From, To int }

// PlaneCleared: the radio removed a plane token from approach Space.
type PlaneCleared struct{ Space int }

// SwitchDeployed: the switch of Space turned green (gear, flaps, brakes).
type SwitchDeployed struct{ Space SpaceID }

type CoffeeGained struct{}

type FuelChanged struct{ From, To int }

// IceDeployed: the ice-brake marker moved to slot Marker.
type IceDeployed struct{ Marker int }

type RerollSpent struct{ Role Role }

type RerollConfirmed struct{ Role Role }

// RerollGained: Mastery brought a reroll token back to the supply.
type RerollGained struct{}

type AdaptationUsed struct{ Role Role }

type AnticipationUsed struct{ Role Role }

// SwapOffered: Role put a die of Value on the Working Together card (public, §13).
type SwapOffered struct {
	Role  Role
	Value int
}

// DiceSwapped: the Working Together dice swapped values (which ones stays private).
type DiceSwapped struct{}

// SyncRolled: Synchronisation rolled the traffic die (§13).
type SyncRolled struct{ Value int }

// SyncDiscarded: the Synchronisation die had no legal space.
type SyncDiscarded struct{}

// DiceDiscarded: Role had no legal placement; Count dice were discarded (§5.2).
type DiceDiscarded struct {
	Role  Role
	Count int
}

type RoundEnded struct{ Round int }

type GameWon struct{}

// GameLost: Failed lists the failed conditions of an end-of-game loss; PlayerID is who gave up.
type GameLost struct {
	Reason   LossReason
	Failed   []Condition
	PlayerID PlayerID
}

type GameEnded struct{ ByPlayerID PlayerID }

type RematchCreated struct{ GameID GameID }

func (PlayerJoined) isEvent()     {}
func (PlayerLeft) isEvent()       {}
func (FlightConfigured) isEvent() {}
func (RolesSwapped) isEvent()     {}
func (GameStarted) isEvent()      {}
func (RoundStarted) isEvent()     {}
func (PlayerReady) isEvent()      {}
func (DiceRolled) isEvent()       {}
func (DiePlaced) isEvent()        {}
func (InternTrained) isEvent()    {}
func (AxisMoved) isEvent()        {}
func (WindTurned) isEvent()       {}
func (PlaneMoved) isEvent()       {}
func (PlaneCleared) isEvent()     {}
func (SwitchDeployed) isEvent()   {}
func (CoffeeGained) isEvent()     {}
func (FuelChanged) isEvent()      {}
func (IceDeployed) isEvent()      {}
func (RerollSpent) isEvent()      {}
func (RerollConfirmed) isEvent()  {}
func (RerollGained) isEvent()     {}
func (AdaptationUsed) isEvent()   {}
func (AnticipationUsed) isEvent() {}
func (SwapOffered) isEvent()      {}
func (DiceSwapped) isEvent()      {}
func (SyncRolled) isEvent()       {}
func (SyncDiscarded) isEvent()    {}
func (DiceDiscarded) isEvent()    {}
func (RoundEnded) isEvent()       {}
func (GameWon) isEvent()          {}
func (GameLost) isEvent()         {}
func (GameEnded) isEvent()        {}
func (RematchCreated) isEvent()   {}
