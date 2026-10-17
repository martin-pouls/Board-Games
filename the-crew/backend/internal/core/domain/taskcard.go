package domain

// Deep Sea task cards (deep-sea/rules.md §9): each card is an objective template with a
// difficulty per crew size. The cards themselves are data (tasks_deepsea.go).

// TaskCard is one of the 96 Deep Sea task cards.
type TaskCard struct {
	ID         int       // 1..96, the numbering of deep-sea/rules.md §9
	Difficulty [3]int    // for 3, 4 and 5 crew members (2 players with the dummy: the 3 column)
	Objective  Objective // the template; its Owner is ignored
	CountBy    [3]int    // when set, the objective's Count by crew size (48, 49)
	Redeal     [][]Card  // re-deal if one hand holds all cards of one entry (submarine tasks)
}

// TaskCards is the number of Deep Sea task cards.
const TaskCards = 96

// TaskCardFor returns task card id.
func TaskCardFor(id int) (TaskCard, bool) {
	if id < 1 || id > len(deepSeaTasks) {
		return TaskCard{}, false
	}
	return deepSeaTasks[id-1], true
}

// crewColumn is the index of the difficulty column for the given number of crew members (the
// dummy included): 3 → 0, 4 → 1, 5 → 2.
func crewColumn(members int) int {
	return min(max(members, 3), 5) - 3
}

// DifficultyFor is the card's difficulty with the given number of crew members.
func (t TaskCard) DifficultyFor(members int) int { return t.Difficulty[crewColumn(members)] }

// NotForCaptain reports whether the captain may never take the card (cards 4–6).
func (t TaskCard) NotForCaptain() bool { return t.Objective.Kind == KindCmpCaptain }

// objectiveFor returns the card's objective for the given number of crew members, unowned.
func (t TaskCard) objectiveFor(members int) Objective {
	o := t.Objective.clone()
	o.Owner = OwnerUnassigned
	if t.CountBy != [3]int{} {
		o.Count = t.CountBy[crewColumn(members)]
	}
	if o.Kind == KindPredict {
		o.Count = -1
	}
	return o
}
