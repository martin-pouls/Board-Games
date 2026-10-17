// The two editions share one app: the edition is fixed per build (vite.config.ts defines
// __EDITION__ from EDITION or the Vite mode). The domain says trump, dummy and commander; the UI
// says rocket, JARVIS and commander (Planet Nine) or submarine, Tonoja and captain (Deep Sea).
// Every UI string that names one of these goes through `terms`.

export type Edition = 'planetNine' | 'deepSea';

export const edition: Edition = __EDITION__;

export interface Terms {
	/** The game's name: page titles, the home page, the invite. */
	title: string;
	/** What the home page says about the game. */
	tagline: string;
	/** Trump card, lower case: "rocket 4". */
	trump: string;
	trumps: string;
	/** The dummy crew member of a two-player game. */
	dummy: string;
	/** "JARVIS's". */
	dummyPossessive: string;
	/** The member who leads and makes the "decide together" calls, lower case. */
	commander: string;
	/** The communication token. */
	commToken: string;
	/** The browser's theme-color (address bar) in light and dark mode; keep in sync with app.html. */
	themeColor: { light: string; dark: string };
}

export const editionTerms: Record<Edition, Terms> = {
	planetNine: {
		title: 'The Crew',
		tagline: 'Reach Planet Nine together: win every trick you promised, in silence.',
		trump: 'rocket',
		trumps: 'rockets',
		dummy: 'JARVIS',
		dummyPossessive: "JARVIS's",
		commander: 'commander',
		commToken: 'radio token',
		themeColor: { light: '#f2f0ea', dark: '#0b1022' }
	},
	deepSea: {
		title: 'The Crew: Mission Deep Sea',
		tagline: 'Dive into the deep sea together: win every task you took on, in silence.',
		trump: 'submarine',
		trumps: 'submarines',
		dummy: 'Tonoja',
		dummyPossessive: "Tonoja's",
		commander: 'captain',
		commToken: 'sonar token',
		themeColor: { light: '#e7f3f1', dark: '#04131d' }
	}
};

/** The terms of this build's edition. */
export const terms: Terms = editionTerms[edition];

/** "captain" to "Captain", for the start of a sentence. */
export function cap(s: string): string {
	return s.charAt(0).toUpperCase() + s.slice(1);
}

/** The number of missions of this edition (Deep Sea: 41, then the side quests Q1 = 42, Q2 = 43). */
export const missionCount = edition === 'deepSea' ? 43 : 50;

/** The name of mission n in this edition: its number, or "Q1" / "Q2" for Deep Sea's side quests. */
export function missionName(n: number): string {
	if (edition === 'deepSea' && n > 41) return `Q${n - 41}`;
	return String(n);
}
