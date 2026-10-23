// Display helpers: names and texts for roles, scenarios, modules, abilities, spaces, losses and
// conditions. UX only; the rules live on the server.
import type {
	AbilityId,
	Condition,
	LossReason,
	Module,
	PlayerSummary,
	PlayerView,
	Role,
	ScenarioInfo,
	SpaceId
} from '#lib/types.js';

export const ROLE_NAME: Record<Role, string> = { pilot: 'Pilot', copilot: 'Co-Pilot' };

export const otherRole = (r: Role): Role => (r === 'pilot' ? 'copilot' : 'pilot');

export const COLOUR_NAME: Record<ScenarioInfo['colour'], string> = {
	green: 'Green',
	yellow: 'Yellow',
	red: 'Red',
	black: 'Black'
};

/** Difficulty order of the scenario colours (the catalogue's own order within a colour). */
export const COLOURS: ScenarioInfo['colour'][] = ['green', 'yellow', 'red', 'black'];

export const MODULE: Record<Module, { name: string; text: string }> = {
	kerosene: {
		name: 'Kerosene',
		text: 'A die on the fuel space burns its value. A round without one burns 6. Below 0 loses.'
	},
	keroseneLeak: {
		name: 'Kerosene leak',
		text: 'Every round burns the difference of the two engine dice plus 1. Below 0 loses.'
	},
	intern: {
		name: 'Intern',
		text: 'Train the intern with any die but his next token, then place that token like a die. All six must be used.'
	},
	wind: {
		name: 'Wind',
		text: 'The wind turns with the axis and is added to the engine sum.'
	},
	iceBrakes: {
		name: 'Ice brakes',
		text: 'Pair dice on the columns 2 to 5, in order, to free the brakes. The marker must pass 5.'
	}
};

export const ABILITY: Record<AbilityId, { name: string; text: string }> = {
	adaptation: {
		name: 'Adaptation',
		text: 'Once per game, each player may turn one of their unplaced dice to its opposite face.'
	},
	anticipation: {
		name: 'Anticipation',
		text: 'Once per round, before the first die is placed, the first player may reroll one of their dice.'
	},
	control: {
		name: 'Control',
		text: 'When both axis spaces hold the same value, gain a coffee.'
	},
	mastery: {
		name: 'Mastery',
		text: 'When both engine spaces hold the same value, gain a reroll token if one is available.'
	},
	synchronisation: {
		name: 'Synchronisation',
		text: 'When a round has a die on the landing gear and one on the flaps, the black die is rolled and the Co-Pilot places it.'
	},
	workingTogether: {
		name: 'Working Together',
		text: 'Once per round, put an unplaced die on the card; your partner does the same and the two dice swap values.'
	}
};

/** Short chip labels for the ability row. */
export const ABILITY_SHORT: Record<AbilityId, string> = {
	adaptation: 'Adapt',
	anticipation: 'Anticip.',
	control: 'Control',
	mastery: 'Mastery',
	synchronisation: 'Sync',
	workingTogether: 'Together'
};

export const LOSS: Record<LossReason, { title: string; text: string }> = {
	spin: { title: 'The plane spun', text: 'The axis tilted too far.' },
	collision: { title: 'Collision', text: 'The plane flew into another plane.' },
	overshoot: { title: 'Overshot the airport', text: 'The plane flew past the runway.' },
	turn: { title: 'Missed the turn', text: 'The axis was wrong for the turn.' },
	missedAirport: {
		title: 'Missed the airport',
		text: 'The plane ran out of altitude before it reached the airport.'
	},
	mandatory: {
		title: 'Flying blind',
		text: 'An axis or engine space stayed empty at the end of a round.'
	},
	kerosene: { title: 'Out of fuel', text: 'The kerosene ran out.' },
	landing: { title: 'Bad landing', text: 'The landing conditions were not met.' },
	intern: { title: 'Intern left behind', text: 'Not every intern token was used.' },
	iceBrakes: { title: 'Icy runway', text: 'The ice brakes were not freed.' },
	gaveUp: { title: 'Gave up', text: 'A pilot abandoned the flight.' }
};

export const CONDITION: Record<Condition, string> = {
	A: 'Planes are still on the approach.',
	B: 'Not all landing gear and flaps are out.',
	C: 'The axis is not level.',
	D: 'The speed is too high for the brakes.',
	intern: 'Intern tokens are left.',
	iceBrakes: 'The ice brakes are not freed.'
};

/** "Landing: …" style list for a lost game. */
export function failedConditions(failed: Condition[]): string[] {
	return failed.map((c) => CONDITION[c]);
}

export function lossTitle(reason: LossReason | null): string {
	return reason ? LOSS[reason].title : 'Flight lost';
}

export function resultHeadline(result: NonNullable<PlayerView['result']>): string {
	return result.won ? 'Safe landing' : lossTitle(result.reason);
}

/** "TGU red" for headers. */
export const scenarioShort = (s: ScenarioInfo): string => `${s.airport} ${s.colour}`;

/** "TGU Toncontín, red". */
export const scenarioName = (s: ScenarioInfo): string => `${s.airport} ${s.city}, ${s.colour}`;

/** The scenarios grouped by colour in difficulty order, keeping the catalogue order. */
export function scenarioGroups(
	catalogue: ScenarioInfo[]
): { colour: ScenarioInfo['colour']; scenarios: ScenarioInfo[] }[] {
	return COLOURS.map((colour) => ({
		colour,
		scenarios: catalogue.filter((s) => s.colour === colour)
	})).filter((g) => g.scenarios.length > 0);
}

/** The scenario's modules as a list of names; "No modules" for the plain ones. */
export function moduleList(s: ScenarioInfo): string {
	return s.modules.length ? s.modules.map((m) => MODULE[m].name).join(', ') : 'No modules';
}

export function plural(n: number, word: string, many = `${word}s`): string {
	return `${n} ${n === 1 ? word : many}`;
}

/** The altitude printed on each round's space. */
export const ALTITUDE = ['6000', '5000', '4000', '3000', '2000', '1000', 'Landing'];

export const altitudeLabel = (round: number): string => ALTITUDE[round - 1] ?? '';

/** The first player of a round: the Pilot in odd rounds. */
export const firstPlayerOf = (round: number): Role => (round % 2 === 1 ? 'pilot' : 'copilot');

const SPACE_LABEL: Record<SpaceId, string> = {
	axisPilot: 'Axis',
	axisCopilot: 'Axis',
	enginesPilot: 'Engines',
	enginesCopilot: 'Engines',
	radioPilot: 'Radio',
	radioCopilot1: 'Radio',
	radioCopilot2: 'Radio',
	gear12: 'Gear 1–2',
	gear34: 'Gear 3–4',
	gear56: 'Gear 5–6',
	flaps12: 'Flaps 1–2',
	flaps23: 'Flaps 2–3',
	flaps45: 'Flaps 4–5',
	flaps56: 'Flaps 5–6',
	brakes2: 'Brake 2',
	brakes4: 'Brake 4',
	brakes6: 'Brake 6',
	concentration1: 'Coffee',
	concentration2: 'Coffee',
	concentration3: 'Coffee',
	kerosene: 'Fuel',
	internPilot: 'Train',
	internCopilot: 'Train',
	ice2Top: 'Ice 2',
	ice3Top: 'Ice 3',
	ice4Top: 'Ice 4',
	ice5Top: 'Ice 5',
	ice2Bottom: 'Ice 2',
	ice3Bottom: 'Ice 3',
	ice4Bottom: 'Ice 4',
	ice5Bottom: 'Ice 5'
};

export const spaceLabel = (id: SpaceId): string => SPACE_LABEL[id];

/** The value each space accepts: null = any, else the allowed values. */
export function spaceValues(id: SpaceId): number[] | null {
	const digits = /^(?:gear|flaps)(\d)(\d)$/.exec(id);
	if (digits) {
		const [a, b] = [Number(digits[1]), Number(digits[2])];
		return [a, b];
	}
	const single = /^(?:brakes|ice)(\d)/.exec(id);
	return single ? [Number(single[1])] : null;
}

/** Axis positions −2…+2 that a turn mask allows (`o` = permitted); null mask = no icon. */
export function turnAllows(mask: string | null, axis: number): boolean {
	if (!mask) return true;
	return mask[axis + 2] === 'o';
}

/** The viewer's own seat in a view. */
export const meOf = (v: PlayerView): PlayerSummary | undefined => v.players.find((p) => p.isYou);

/** The other player's seat, when there is one. */
export const partnerOf = (v: PlayerView): PlayerSummary | undefined =>
	v.players.find((p) => !p.isYou);

/** The viewer's role; the Pilot while the seats are unknown. */
export const myRole = (v: PlayerView): Role => meOf(v)?.role ?? 'pilot';
