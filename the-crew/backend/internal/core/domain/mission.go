package domain

import "slices"

// Edition is the game a mission table belongs to. One deployment serves exactly one edition.
type Edition string

const (
	EditionPlanetNine Edition = "planetNine"
	EditionDeepSea    Edition = "deepSea"
)

// Valid reports whether e is a known edition.
func (e Edition) Valid() bool { return e == EditionPlanetNine || e == EditionDeepSea }

// Distribution is how the drawn tasks reach their owners (rules §6, §10).
type Distribution string

const (
	DistNone                  Distribution = ""                      // the mission has no tasks
	DistOpen                  Distribution = "open"                  // commander first, then clockwise
	DistCommanderDecision     Distribution = "commanderDecision"     // poll, then all tasks to one other member
	DistCommanderDistribution Distribution = "commanderDistribution" // per revealed task: poll, then any member

	// Deep Sea (deep-sea/rules.md §3, §8).
	DistDraft              Distribution = "draft"              // captain first, then clockwise; passing while there are fewer tasks than members
	DistOneMember          Distribution = "oneMember"          // 6: the captain gives all tasks to one member
	DistCaptainOrVolunteer Distribution = "captainOrVolunteer" // 10, 13: poll, then all tasks to one member
	DistVolunteers         Distribution = "volunteers"         // 14–16, 26: poll, then Volunteers members take all tasks
	DistFree               Distribution = "free"               // the captain assigns every task
)

// CommMode is a Deep Sea communication mode (deep-sea/rules.md §4).
type CommMode string

const (
	CommNormal     CommMode = ""
	CommCurrents   CommMode = "currents"   // = dead zone
	CommRapture    CommMode = "rapture"    // n−2 shared tokens, repeat use allowed
	CommUnfamiliar CommMode = "unfamiliar" // a random colour card picks normal, currents or rapture per attempt
	CommNone       CommMode = "none"       // no communication
)

// CommRule restricts communication (rules §8).
type CommRule struct {
	DeadZone  bool // the position is not stated
	FromTrick int  // disruption: nobody communicates before this trick; 0 = no restriction
	Mode      CommMode
}

// SetupAction is an extra step of a mission (rules §16).
type SetupAction string

const (
	SetupSwapTokens          SetupAction = "swapTokens"          // 23: the commander may swap two tasks' tokens
	SetupMoveToken           SetupAction = "moveToken"           // 40: the commander may move a token to a task without one
	SetupSilenceMember       SetupAction = "silenceMember"       // 11: the commander silences a human
	SetupDrawAfterFirstTrick SetupAction = "drawAfterFirstTrick" // 12: each human draws from the right neighbour
)

// PollKind is the answer pair of a mission's polls (rules §10).
type PollKind string

const (
	PollYesNo   PollKind = "yesNo"
	PollWellIll PollKind = "wellIll" // mission 5
)

// OwnerRule says who owns a special objective.
type OwnerRule string

const (
	OwnerRuleCrew         OwnerRule = "crew"
	OwnerRuleCommander    OwnerRule = "commander"
	OwnerRuleChosen       OwnerRule = "chosen"       // after a poll, the commander picks another member
	OwnerRuleRoleA        OwnerRule = "roleA"        // 50: picked by the commander together with roleB
	OwnerRuleRoleB        OwnerRule = "roleB"        //
	OwnerRuleLeftOfHolder OwnerRule = "leftOfHolder" // 46: the next seat after Holder's holder, fixed at the deal
)

// ObjectiveSpec is a special objective of a mission: a template plus how it is completed at the
// deal. The template's Owner is ignored.
type ObjectiveSpec struct {
	Objective Objective
	Owner     OwnerRule
	Holder    Card  // OwnerRuleLeftOfHolder
	TaskToken Token // when set, the template's Cards become the card of the task with this token
}

// MissionDef is one row of the mission table (rules §14; Deep Sea: deep-sea/rules.md §8).
type MissionDef struct {
	Number       int
	Tasks        int     // task cards to draw
	Tokens       []Token // in placement order: the first drawn tasks carry them
	Distribution Distribution
	Comm         CommRule
	Golden       bool // the five-player handover is allowed (rules §11)
	Setup        []SetupAction
	Poll         PollKind // the answers of this mission's polls
	Objectives   []ObjectiveSpec

	// Deep Sea.
	Difficulty     int   // total difficulty of the task cards to draw
	FixedTasks     []int // 32: these task cards instead of drawing
	Volunteers     int   // DistVolunteers: how many members take the tasks
	CaptainSkipped bool  // 25: the captain receives no task
	HardestFirst   bool  // 19: the captain's first task is one of the highest difficulty
	Uncounted      bool  // epilogue: the logbook counts no attempts
}

// DrawsTasks reports whether the mission draws Deep Sea task cards by difficulty.
func (m MissionDef) DrawsTasks() bool { return m.Difficulty > 0 }

// Has reports whether the mission has the setup action a.
func (m MissionDef) Has(a SetupAction) bool { return slices.Contains(m.Setup, a) }

// missionTable is the mission table of the edition; the one place that switches on it.
func missionTable(e Edition) []MissionDef {
	switch e {
	case EditionPlanetNine:
		return planetNine
	case EditionDeepSea:
		return deepSea
	}
	return nil
}

// Missions returns the number of missions of the edition.
func Missions(e Edition) int { return len(missionTable(e)) }

// MissionFor returns mission n of the edition.
func MissionFor(e Edition, n int) (MissionDef, bool) {
	t := missionTable(e)
	if n < 1 || n > len(t) {
		return MissionDef{}, false
	}
	return t[n-1], true
}
