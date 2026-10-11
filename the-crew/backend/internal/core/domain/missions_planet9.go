package domain

// The Planet Nine missions, transcribed row by row from rules.md §14. Data only.

var (
	arrows3   = []Token{TokenArrow1, TokenArrow2, TokenArrow3}
	colourOne = []Card{{SuitPink, 1}, {SuitBlue, 1}, {SuitGreen, 1}, {SuitYellow, 1}}
	trumps    = []Card{{SuitTrump, 1}, {SuitTrump, 2}, {SuitTrump, 3}, {SuitTrump, 4}}
)

// open is a mission with n tasks in open distribution and the given tokens.
func open(n int, tokens ...Token) MissionDef {
	return MissionDef{Tasks: n, Tokens: tokens, Distribution: DistOpen}
}

// none is a mission without tasks.
func none() MissionDef { return MissionDef{} }

func (m MissionDef) dist(d Distribution) MissionDef { m.Distribution = d; return m }
func (m MissionDef) deadZone() MissionDef           { m.Comm.DeadZone = true; return m }
func (m MissionDef) disruption(n int) MissionDef    { m.Comm.FromTrick = n; return m }
func (m MissionDef) golden() MissionDef             { m.Golden = true; return m }
func (m MissionDef) setup(a SetupAction) MissionDef {
	m.Setup = append(m.Setup, a)
	return m
}
func (m MissionDef) poll(p PollKind) MissionDef { m.Poll = p; return m }
func (m MissionDef) objective(s ObjectiveSpec) MissionDef {
	m.Objectives = append(m.Objectives, s)
	return m
}

func crew(o Objective) ObjectiveSpec               { return ObjectiveSpec{Objective: o, Owner: OwnerRuleCrew} }
func owned(r OwnerRule, o Objective) ObjectiveSpec { return ObjectiveSpec{Objective: o, Owner: r} }

var planetNine = numbered([]MissionDef{
	/* 1 */ open(1),
	/* 2 */ open(2),
	/* 3 */ open(2, Token1, Token2),
	/* 4 */ open(3),
	/* 5 */ none().poll(PollWellIll).objective(owned(OwnerRuleChosen, Objective{Kind: KindWinNoTricks})),
	/* 6 */ open(3, TokenArrow1, TokenArrow2).deadZone(),
	/* 7 */ open(3, TokenOmega),
	/* 8 */ open(3, Token1, Token2, Token3),
	/* 9 */ none().objective(crew(Objective{Kind: KindCrewWinWithAny, Cards: colourOne, Count: 1})),
	/* 10 */ open(4),
	/* 11 */ open(4, Token1).setup(SetupSilenceMember),
	/* 12 */ open(4, TokenOmega).setup(SetupDrawAfterFirstTrick),
	/* 13 */ none().objective(crew(Objective{Kind: KindCrewWinWithEach, Cards: trumps})),
	/* 14 */ open(4, arrows3...).deadZone(),
	/* 15 */ open(4, Token1, Token2, Token3, Token4),
	/* 16 */ none().objective(crew(Objective{Kind: KindCrewNeverWinWith, Rank: 9})),
	/* 17 */ open(2).objective(crew(Objective{Kind: KindCrewNeverWinWith, Rank: 9})),
	/* 18 */ open(5).disruption(2),
	/* 19 */ open(5, Token1).disruption(3),
	/* 20 */ open(2).dist(DistCommanderDecision),
	/* 21 */ open(5, Token1, Token2).deadZone(),
	/* 22 */ open(5, TokenArrow1, TokenArrow2, TokenArrow3, TokenArrow4),
	/* 23 */ open(5, Token1, Token2, Token3, Token4, Token5).setup(SetupSwapTokens),
	/* 24 */ open(6).dist(DistCommanderDistribution),
	/* 25 */ open(6, TokenArrow1, TokenArrow2).deadZone().golden(),
	/* 26 */ none().objective(crew(Objective{Kind: KindCrewWinWithAny, Cards: colourOne, Count: 2})),
	/* 27 */ open(3).dist(DistCommanderDecision).golden(),
	/* 28 */ open(6, Token1, TokenOmega).disruption(3).golden(),
	/* 29 */ none().deadZone().objective(crew(Objective{Kind: KindCrewBalance, Count: 1})),
	/* 30 */ open(6, arrows3...).disruption(2).golden(),
	/* 31 */ open(6, Token1, Token2, Token3).golden(),
	/* 32 */ open(7).dist(DistCommanderDistribution).golden(),
	/* 33 */ none().poll(PollYesNo).objective(owned(OwnerRuleChosen,
		Objective{Kind: KindWinExactlyTricks, Count: 1, NotWithTrump: true})),
	/* 34 */ none().
		objective(crew(Objective{Kind: KindCrewBalance, Count: 1})).
		objective(owned(OwnerRuleCommander, Objective{Kind: KindWinTricks, Tricks: []int{1, LastTrick}})),
	/* 35 */ open(7, arrows3...).golden(),
	/* 36 */ open(7, Token1, Token2).dist(DistCommanderDistribution).golden(),
	/* 37 */ open(4).dist(DistCommanderDecision).golden(),
	/* 38 */ open(8).disruption(3).golden(),
	/* 39 */ open(8, arrows3...).deadZone().golden(),
	/* 40 */ open(8, Token1, Token2, Token3).golden().setup(SetupMoveToken),
	/* 41 */ none().poll(PollYesNo).objective(owned(OwnerRuleChosen,
		Objective{Kind: KindWinOnlyTricks, Tricks: []int{1, LastTrick}, NotWithTrump: true})),
	/* 42 */ open(9).golden(),
	/* 43 */ open(9).dist(DistCommanderDistribution).golden(),
	/* 44 */ none().objective(crew(Objective{Kind: KindCrewWinWithEach, Cards: trumps, Ordered: true})),
	/* 45 */ open(9, arrows3...).golden(),
	/* 46 */ none().objective(ObjectiveSpec{
		Objective: Objective{Kind: KindWinAllOfSuit, Suit: SuitPink},
		Owner:     OwnerRuleLeftOfHolder,
		Holder:    Card{SuitPink, 9},
	}),
	/* 47 */ open(10).golden(),
	/* 48 */ open(3, TokenOmega).golden().objective(ObjectiveSpec{
		Objective: Objective{Kind: KindCardInLastTrick},
		Owner:     OwnerRuleCrew,
		TaskToken: TokenOmega,
	}),
	/* 49 */ open(10, arrows3...).golden(),
	/* 50 */ none().
		objective(owned(OwnerRuleRoleA, Objective{Kind: KindWinOnlyTricks, Tricks: []int{1, 2, 3, 4}})).
		objective(owned(OwnerRuleRoleB, Objective{Kind: KindWinOnlyTricks, Tricks: []int{LastTrick}})),
})

// numbered sets each mission's Number from its position and gives every mission a poll kind.
func numbered(ms []MissionDef) []MissionDef {
	for i := range ms {
		ms[i].Number = i + 1
		if ms[i].Poll == "" {
			ms[i].Poll = PollYesNo
		}
	}
	return ms
}
