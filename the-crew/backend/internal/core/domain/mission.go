package domain

import "slices"

// Edition is the game a mission table belongs to. Only Planet Nine exists so far; Deep Sea would
// add its own table and task source.
type Edition string

const EditionPlanetNine Edition = "planetNine"

// Distribution is how the drawn tasks reach their owners (rules §6, §10).
type Distribution string

const (
	DistNone                  Distribution = ""                      // the mission has no tasks
	DistOpen                  Distribution = "open"                  // commander first, then clockwise
	DistCommanderDecision     Distribution = "commanderDecision"     // poll, then all tasks to one other member
	DistCommanderDistribution Distribution = "commanderDistribution" // per revealed task: poll, then any member
)

// CommRule restricts communication (rules §8).
type CommRule struct {
	DeadZone  bool // the position is not stated
	FromTrick int  // disruption: nobody communicates before this trick; 0 = no restriction
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

// MissionDef is one row of the mission table (rules §14).
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
}

// Has reports whether the mission has the setup action a.
func (m MissionDef) Has(a SetupAction) bool { return slices.Contains(m.Setup, a) }

// Missions returns the number of missions of the edition.
func Missions(e Edition) int {
	if e == EditionPlanetNine {
		return len(planetNine)
	}
	return 0
}

// MissionFor returns mission n of the edition.
func MissionFor(e Edition, n int) (MissionDef, bool) {
	if e != EditionPlanetNine || n < 1 || n > len(planetNine) {
		return MissionDef{}, false
	}
	return planetNine[n-1], true
}
