package domain

// InternTokens is the number of intern tokens, values 1..6 (§12.2).
const InternTokens = 6

// newInternRow returns the intern tokens in random order (§12.2). RNG use: Shuffle(6).
func newInternRow(rng RNG) []int {
	row := make([]int, InternTokens)
	for i := range row {
		row[i] = i + 1
	}
	rng.Shuffle(len(row), func(i, j int) { row[i], row[j] = row[j], row[i] })
	return row
}

// internNext returns the index in f.Intern of role's next token: the Pilot takes from the left
// end, the Co-Pilot from the right end (§12.2). False without the module or when the row is used up.
func (f *Flight) internNext(role Role) (int, bool) {
	if f.Intern == nil || f.InternLeft > f.InternRight {
		return 0, false
	}
	if role == RolePilot {
		return f.InternLeft, true
	}
	return f.InternRight, true
}

// internTaken reports whether the token at index i of the row has been used.
func (f *Flight) internTaken(i int) bool {
	return i < f.InternLeft || i > f.InternRight
}

// internLeft returns the number of tokens still on the intern board.
func (f *Flight) internLeft() int {
	if f.Intern == nil {
		return 0
	}
	return max(0, f.InternRight-f.InternLeft+1)
}
