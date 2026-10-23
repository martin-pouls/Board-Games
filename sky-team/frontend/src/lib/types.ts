// TypeScript mirror of the Go wire contract (backend/internal/adapters/apidto). Keep in sync.
// The view, events and commands are transcribed from implementation-plan.md §6.
import type {
	ClientErrorCode,
	CommonErrorCode,
	ErrorBody as SharedErrorBody,
	ServerMessage as SharedServerMessage
} from '@board-games/shared/wire.js';

export type Phase = 'lobby' | 'playing' | 'won' | 'lost' | 'ended';
export type Role = 'pilot' | 'copilot';
export type Piece = 'pilot' | 'copilot' | 'intern' | 'traffic';
export type Step = 'discussion' | 'place' | 'reroll' | 'swap' | 'sync';
export type SpaceId =
	| 'axisPilot'
	| 'axisCopilot'
	| 'enginesPilot'
	| 'enginesCopilot'
	| 'radioPilot'
	| 'radioCopilot1'
	| 'radioCopilot2'
	| 'gear12'
	| 'gear34'
	| 'gear56'
	| 'flaps12'
	| 'flaps23'
	| 'flaps45'
	| 'flaps56'
	| 'brakes2'
	| 'brakes4'
	| 'brakes6'
	| 'concentration1'
	| 'concentration2'
	| 'concentration3'
	| 'kerosene'
	| 'internPilot'
	| 'internCopilot'
	| 'ice2Top'
	| 'ice3Top'
	| 'ice4Top'
	| 'ice5Top'
	| 'ice2Bottom'
	| 'ice3Bottom'
	| 'ice4Bottom'
	| 'ice5Bottom';
export type Module = 'kerosene' | 'keroseneLeak' | 'intern' | 'wind' | 'iceBrakes';
export type AbilityId =
	'adaptation' | 'anticipation' | 'control' | 'mastery' | 'synchronisation' | 'workingTogether';
export type LossReason =
	| 'spin'
	| 'collision'
	| 'overshoot'
	| 'turn'
	| 'missedAirport'
	| 'mandatory'
	| 'kerosene'
	| 'landing'
	| 'intern'
	| 'iceBrakes'
	| 'gaveUp';
export type Condition = 'A' | 'B' | 'C' | 'D' | 'intern' | 'iceBrakes';

export interface ScenarioInfo {
	id: string;
	airport: string;
	city: string;
	colour: 'green' | 'yellow' | 'red' | 'black';
	modules: Module[];
	abilities: number;
	realTimeDropped: boolean;
	spaces: number;
}

export interface PlayerSummary {
	id: string;
	name: string;
	seat: number;
	role: Role;
	isHost: boolean;
	isYou: boolean;
	mustAct: boolean;
	diceLeft: number;
	ready: boolean;
	revealedDice: number[] | null;
}

export interface Placement {
	space: SpaceId;
	piece: Piece;
	value: number;
	coffee: number;
	role: Role | null;
}

export interface FlightView {
	round: number;
	firstPlayer: Role;
	step: Step;
	turn: Role;
	approach: {
		spaces: { planes: number; traffic: number; turn: string | null }[];
		position: number;
	};
	/** Altitude rounds still holding a reroll token. */
	rerollRounds: number[];
	axis: number;
	aero: { blue: number; orange: number };
	gear: [boolean, boolean, boolean];
	flaps: number;
	brakes: number;
	brakeLimit: number; // the round-7 speed must be lower than this (brake or ice-brake marker)
	coffee: number;
	rerolls: number;
	planeSupply: number;
	fuel: number | null;
	leak: boolean;
	wind: { offset: number; value: number } | null;
	ice: { marker: number } | null;
	intern: {
		tokens: { value: number; taken: boolean }[];
		nextPilot: number | null;
		nextCopilot: number | null;
	} | null;
	/** In placement order. */
	placements: Placement[];
	speed: { sum: number; wind: number; total: number; move: number | null } | null;
	roundStart: { rerollPicked: boolean; traffic: { roll: number; space: number }[] };
	/** 4 slots; null = placed or discarded. */
	yourDice: (number | null)[];
	abilityState: {
		adaptationUsed: Record<Role, boolean>;
		anticipationUsed: boolean;
		togetherUsed: boolean;
		syncUsed: boolean;
		controlTriggered: boolean;
		masteryTriggered: boolean;
	};
	can: { spendReroll: boolean; adapt: boolean; anticipate: boolean; offerSwap: boolean };
	/** Your turn in `place` only; `byValue[v - 1]`. */
	legal: { byValue: SpaceId[][]; internToken: number | null; internTargets: SpaceId[] } | null;
	reroll: { confirmed: Record<Role, boolean> } | null;
	/** `value` is public; `yourSlot` only for the initiator. */
	swap: { by: Role; value: number; yourSlot: number | null } | null;
	/** `targets` only for the Co-Pilot. */
	sync: { value: number; targets: SpaceId[] } | null;
}

export interface PlayerView {
	gameId: string;
	revision: number;
	phase: Phase;
	you: string;
	hostId: string;
	nextGameId: string;
	players: PlayerSummary[];
	scenario: ScenarioInfo;
	abilities: AbilityId[];
	/** Lobby only: the 23 scenarios in rules.md §9 order. */
	catalogue: ScenarioInfo[] | null;
	/** Null in the lobby. */
	flight: FlightView | null;
	result: { won: boolean; reason: LossReason | null; failed: Condition[] } | null;
}

/** The host's remembered scenario and abilities, sent with `POST /api/games` as `flight`. */
export interface FlightOptions {
	scenario: string;
	abilities: AbilityId[];
}

export interface GameSummary {
	gameId: string;
	phase: Phase;
	hostName: string;
	playerCount: number;
	maxPlayers: number;
	joinable: boolean;
	/** RFC3339 UTC; absent when unknown. */
	createdAt?: string;
}
export interface SessionResponse {
	gameId: string;
	playerId: string;
	token: string;
	view: PlayerView;
}

export type GameEvent =
	| { type: 'playerJoined'; playerId: string; name: string }
	| { type: 'playerLeft'; playerId: string; hostId: string }
	| { type: 'flightConfigured' }
	| { type: 'rolesSwapped' }
	| { type: 'gameStarted' }
	| {
			type: 'roundStarted';
			round: number;
			rerollPicked: boolean;
			traffic: { roll: number; space: number }[];
	  }
	| { type: 'playerReady'; role: Role }
	| { type: 'diceRolled' }
	| {
			type: 'diePlaced';
			role: Role | null;
			piece: Piece;
			space: SpaceId;
			value: number;
			coffee: number;
	  }
	| { type: 'internTrained'; role: Role; token: number; space: SpaceId }
	| { type: 'axisMoved'; from: number; to: number }
	| { type: 'windTurned'; offset: number; value: number }
	| { type: 'planeMoved'; from: number; to: number }
	| { type: 'planeCleared'; space: number }
	| { type: 'switchDeployed'; space: SpaceId }
	| { type: 'coffeeGained' }
	| { type: 'fuelChanged'; from: number; to: number }
	| { type: 'iceDeployed'; marker: number }
	| { type: 'rerollSpent'; role: Role }
	| { type: 'rerollConfirmed'; role: Role }
	| { type: 'rerollGained' }
	| { type: 'adaptationUsed'; role: Role }
	| { type: 'anticipationUsed'; role: Role }
	| { type: 'swapOffered'; role: Role; value: number }
	| { type: 'diceSwapped' }
	| { type: 'syncRolled'; value: number }
	| { type: 'syncDiscarded' }
	| { type: 'diceDiscarded'; role: Role; count: number }
	| { type: 'roundEnded'; round: number }
	| { type: 'gameWon' }
	| { type: 'gameLost'; reason: LossReason; failed: Condition[] }
	| { type: 'gameEnded'; byPlayerId: string }
	| { type: 'rematchCreated'; gameId: string };

/** Sky Team's own error codes, next to the ones every game shares (shared/frontend wire.ts). */
export type GameErrorCode =
	| 'not_host'
	| 'wrong_phase'
	| 'game_full'
	| 'name_taken'
	| 'cannot_leave'
	| 'invalid_name'
	| 'unknown_player'
	| 'unknown_scenario'
	| 'invalid_abilities'
	| 'not_enough_players'
	| 'abilities_incomplete'
	| 'wrong_step'
	| 'already_ready'
	| 'not_your_turn'
	| 'no_such_die'
	| 'not_enough_coffee'
	| 'value_out_of_range'
	| 'illegal_space'
	| 'illegal_token_space'
	| 'no_reroll'
	| 'already_confirmed'
	| 'ability_not_in_play'
	| 'ability_used'
	| 'partner_has_no_dice'
	| 'rematch_exists';
export type ApiErrorCode = CommonErrorCode | ClientErrorCode | GameErrorCode;
export type ErrorBody = SharedErrorBody<ApiErrorCode>;
export type ServerMessage = SharedServerMessage<PlayerView, GameEvent, ApiErrorCode>;

/** Sent as ClientCommand & { requestId: string }. */
export type ClientCommand =
	| { type: 'configure'; scenario: string; abilities: AbilityId[] }
	| { type: 'swapRoles' }
	| { type: 'startGame' }
	| { type: 'ready' }
	| { type: 'place'; slot: number; space: SpaceId; coffee: number; tokenSpace?: SpaceId }
	| { type: 'spendReroll' }
	| { type: 'confirmReroll'; slots: number[] }
	| { type: 'adapt'; slot: number }
	| { type: 'anticipate'; slot: number }
	| { type: 'offerSwap'; slot: number }
	| { type: 'answerSwap'; slot: number }
	| { type: 'placeSync'; space: SpaceId }
	| { type: 'giveUp' }
	| { type: 'endGame' }
	| { type: 'rematch' };
