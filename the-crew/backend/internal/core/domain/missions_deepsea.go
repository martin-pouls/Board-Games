package domain

// The Deep Sea missions, transcribed row by row from deep-sea/rules.md §8. Data only. Real-time
// missions use their untimed variant (14–16, 26). The side quests Q1 and Q2 are left out (App
// decision).

// draw is a Deep Sea mission drawing task cards of total difficulty d, taken in the draft.
func draw(d int) MissionDef { return MissionDef{Difficulty: d, Distribution: DistDraft} }

func (m MissionDef) comm(mode CommMode) MissionDef { m.Comm.Mode = mode; return m }
func (m MissionDef) volunteers(n int) MissionDef {
	m.Distribution, m.Volunteers = DistVolunteers, n
	return m
}
func (m MissionDef) free() MissionDef     { return m.dist(DistFree) }
func (m MissionDef) epilogue() MissionDef { m.Uncounted = true; return m.free() }

var deepSea = numbered([]MissionDef{
	/* 1 */ draw(1),
	/* 2 */ draw(2),
	/* 3 */ draw(3),
	/* 4 */ draw(4),
	/* 5 */ draw(5),
	/* 6 */ draw(5).dist(DistOneMember),
	/* 7 */ draw(6),
	/* 8 */ none().objective(crew(Objective{Kind: KindCrewBalance, Count: 1, Rank: 9})),
	/* 9 */ draw(7).comm(CommCurrents),
	/* 10 */ draw(4).dist(DistCaptainOrVolunteer),
	/* 11 */ draw(8).comm(CommRapture),
	/* 12 */ none().objective(crew(Objective{Kind: KindNeverLead, Suits: []Suit{SuitPink, SuitTrump}})),
	/* 13 */ draw(5).dist(DistCaptainOrVolunteer),
	/* 14 */ draw(6).volunteers(1).comm(CommCurrents),
	/* 15 */ draw(6).volunteers(1).comm(CommRapture),
	/* 16 */ draw(6).volunteers(1).comm(CommNone),
	/* 17 */ draw(9).free(),
	/* 18 */ draw(9),
	/* 19 */ withHardestFirst(draw(9)),
	/* 20 */ draw(10).comm(CommUnfamiliar),
	/* 21 */ none().comm(CommUnfamiliar).objective(crew(Objective{Kind: KindCrewBalance, Count: 1, Rank: 1})),
	/* 22 */ draw(11).comm(CommUnfamiliar),
	/* 23 */ none().comm(CommUnfamiliar).disruption(2).objective(crew(Objective{Kind: KindLeaderStaysAhead})),
	/* 24 */ draw(12).comm(CommUnfamiliar),
	/* 25 */ withCaptainSkipped(draw(12).comm(CommUnfamiliar)),
	/* 26 */ draw(12).volunteers(2).comm(CommUnfamiliar),
	/* 27 */ draw(5).comm(CommUnfamiliar).objective(crew(Objective{Kind: KindLastCardPlayed, Cards: []Card{{SuitYellow, 5}}})),
	/* 28 */ draw(14).free(),
	/* 29 */ draw(15).free(),
	/* 30 */ draw(16).free(),
	/* 31 */ draw(17).free(),
	/* 32 */ {FixedTasks: []int{74, 88, 85, 80}, Distribution: DistDraft},
	/* 33 */ draw(18).epilogue(),
	/* 34 */ draw(19).epilogue(),
	/* 35 */ draw(20).epilogue(),
	/* 36 */ draw(21).epilogue(),
	/* 37 */ draw(22).epilogue(),
	/* 38 */ draw(23).epilogue(),
	/* 39 */ draw(24).epilogue(),
	/* 40 */ draw(25).epilogue(),
	/* 41 */ draw(26).epilogue(),
})

func withHardestFirst(m MissionDef) MissionDef   { m.HardestFirst = true; return m }
func withCaptainSkipped(m MissionDef) MissionDef { m.CaptainSkipped = true; return m }
