// What the server would put into `legal` and `sync.targets`, computed for the fixtures so their
// data always agrees with the placements. Fixtures only: the real answer comes from the server
// (rules.md §3.1, §6, §12), and the UI never recomputes legality from this.
import type { Module, Placement, Role, SpaceId } from '#lib/types.js';

type SpaceColour = 'blue' | 'orange' | 'both';

interface SpaceDef {
	colour: SpaceColour;
	/** Allowed values; null = any. */
	values: number[] | null;
}

const ANY = null;

export const SPACES: Record<SpaceId, SpaceDef> = {
	axisPilot: { colour: 'blue', values: ANY },
	axisCopilot: { colour: 'orange', values: ANY },
	enginesPilot: { colour: 'blue', values: ANY },
	enginesCopilot: { colour: 'orange', values: ANY },
	radioPilot: { colour: 'blue', values: ANY },
	radioCopilot1: { colour: 'orange', values: ANY },
	radioCopilot2: { colour: 'orange', values: ANY },
	gear12: { colour: 'blue', values: [1, 2] },
	gear34: { colour: 'blue', values: [3, 4] },
	gear56: { colour: 'blue', values: [5, 6] },
	flaps12: { colour: 'orange', values: [1, 2] },
	flaps23: { colour: 'orange', values: [2, 3] },
	flaps45: { colour: 'orange', values: [4, 5] },
	flaps56: { colour: 'orange', values: [5, 6] },
	brakes2: { colour: 'blue', values: [2] },
	brakes4: { colour: 'blue', values: [4] },
	brakes6: { colour: 'blue', values: [6] },
	concentration1: { colour: 'both', values: ANY },
	concentration2: { colour: 'both', values: ANY },
	concentration3: { colour: 'both', values: ANY },
	kerosene: { colour: 'both', values: ANY },
	internPilot: { colour: 'blue', values: ANY },
	internCopilot: { colour: 'orange', values: ANY },
	ice2Top: { colour: 'blue', values: [2] },
	ice3Top: { colour: 'blue', values: [3] },
	ice4Top: { colour: 'blue', values: [4] },
	ice5Top: { colour: 'blue', values: [5] },
	ice2Bottom: { colour: 'both', values: [2] },
	ice3Bottom: { colour: 'both', values: [3] },
	ice4Bottom: { colour: 'both', values: [4] },
	ice5Bottom: { colour: 'both', values: [5] }
};

const ALL_SPACES = Object.keys(SPACES) as SpaceId[];
const FLAPS: SpaceId[] = ['flaps12', 'flaps23', 'flaps45', 'flaps56'];
const BRAKES: SpaceId[] = ['brakes2', 'brakes4', 'brakes6'];
const ICE_COLUMNS = [2, 3, 4, 5];

export interface PanelState {
	modules: Module[];
	placements: Placement[];
	flaps: number;
	brakes: number;
	/** Ice marker slot, 0…4; null without the module. */
	ice: number | null;
	/** The value of the next intern token per role; null = none left or no module. */
	internNext: Record<Role, number | null>;
}

const isIntern = (id: SpaceId) => id === 'internPilot' || id === 'internCopilot';

function active(id: SpaceId, st: PanelState): boolean {
	if (id.startsWith('brakes')) return !st.modules.includes('iceBrakes');
	if (id.startsWith('ice')) return st.modules.includes('iceBrakes');
	if (id === 'kerosene') return st.modules.includes('kerosene');
	if (isIntern(id)) return st.modules.includes('intern');
	return true;
}

/** The wind ring value for an offset from the top space (rules.md §12.3). */
export function windValue(offset: number): number {
	const p = ((offset % 20) + 20) % 20;
	const m = Math.min(p, 20 - p);
	return [3, 3, 2, 2, 1, 0, -1, -2, -2, -3, -3][m];
}

interface Piece {
	/** Whose colour the space must have; null = any colour (the Synchronisation die). */
	role: Role | null;
	value: number;
	/** Intern training is not possible (the piece is the token or the black die). */
	noTraining?: boolean;
	/** Concentration is not possible (the intern token). */
	noConcentration?: boolean;
}

function fits(id: SpaceId, piece: Piece, st: PanelState): boolean {
	const def = SPACES[id];
	if (!active(id, st)) return false;
	if (st.placements.some((p) => p.space === id)) return false;
	if (piece.role && def.colour !== 'both') {
		if ((def.colour === 'blue') !== (piece.role === 'pilot')) return false;
	}
	if (def.values && !def.values.includes(piece.value)) return false;
	if (id.startsWith('concentration') && piece.noConcentration) return false;
	if (isIntern(id)) {
		if (piece.noTraining || !piece.role) return false;
		const next = st.internNext[piece.role];
		return next !== null && next !== piece.value;
	}
	const flap = FLAPS.indexOf(id);
	if (flap >= 0) return flap <= st.flaps;
	const brake = BRAKES.indexOf(id);
	if (brake >= 0) return brake === st.brakes;
	const column = /^ice(\d)/.exec(id);
	if (column && st.ice !== null) return ICE_COLUMNS.indexOf(Number(column[1])) >= st.ice;
	return true;
}

/** The spaces `role`'s die of the final value `value` may go on, in table order. */
export function legalSpaces(role: Role, value: number, st: PanelState): SpaceId[] {
	return ALL_SPACES.filter((id) => fits(id, { role, value }, st));
}

/** `legal.byValue` for one role: index `v - 1` lists the spaces for a final value `v`. */
export function legalByValue(role: Role, st: PanelState): SpaceId[][] {
	return [1, 2, 3, 4, 5, 6].map((v) => legalSpaces(role, v, st));
}

/** Where an intern token may go: anything the trainer could use, but not concentration. */
export function internTargets(role: Role, token: number, st: PanelState): SpaceId[] {
	return ALL_SPACES.filter((id) =>
		fits(id, { role, value: token, noTraining: true, noConcentration: true }, st)
	);
}

/** Where the Synchronisation die may go: any colour, no coffee, no intern space. */
export function syncTargets(value: number, st: PanelState): SpaceId[] {
	return ALL_SPACES.filter((id) => fits(id, { role: null, value, noTraining: true }, st));
}
