package domain

// Data only: the scenarios of §9 and the approach tracks of §10, transcribed from rules.md.
// Lookups and types are in scenario.go.

// scenarios in §9 order.
var scenarios = []ScenarioDef{
	{ID: "yul-green", Airport: "YUL", City: "Montréal-Trudeau", Colour: ColourGreen, Track: "YUL-G"},
}

// tracks in §10 order. Each space: planes at setup, traffic-die icons, turn icon.
var tracks = []TrackDef{
	{ID: "YUL-G", Spaces: []TrackSpace{
		{0, 0, turn("")},
		{0, 0, turn("")},
		{1, 0, turn("")},
		{2, 0, turn("")},
		{1, 0, turn("")},
		{3, 0, turn("")},
		{2, 0, turn("")},
	}},
}
