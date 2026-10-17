// TypeScript mirror of the Go wire contract (backend/internal/adapters/apidto). Keep in sync.
import type {
	ClientErrorCode,
	CommonErrorCode,
	ErrorBody as SharedErrorBody,
	ServerMessage as SharedServerMessage
} from '@board-games/shared/wire.js';

export type Phase = 'lobby' | 'playing' | 'debrief' | 'ended';

/** The four colours and trump ("rockets" in the UI). */
export type Suit = 'pink' | 'blue' | 'green' | 'yellow' | 'trump';
/** A playing card: colours 1–9, trump 1–4. */
export interface Card {
	suit: Suit;
	rank: number;
}
/** Task tokens: 1–5 (fixed position), omega (last), arrow1–arrow4 (relative order). */
export type Token =
	'' | '1' | '2' | '3' | '4' | '5' | 'omega' | 'arrow1' | 'arrow2' | 'arrow3' | 'arrow4';

export interface PlayerSummary {
	id: string;
	name: string;
	/** Join order = seat order. */
	seat: number;
	isHost: boolean;
	isYou: boolean;
}

/** The logbook line of one mission. */
export interface LogEntry {
	mission: number;
	attempts: number;
	distress: boolean;
	completed: boolean;
	/** Attempts, plus one with the distress signal. */
	total: number;
}

export type ObjectiveKind =
	| 'winCards'
	| 'winNoTricks'
	| 'winExactlyTricks'
	| 'winOnlyTricks'
	| 'winTricks'
	| 'winAllOfSuit'
	| 'crewWinWithAny'
	| 'crewWinWithEach'
	| 'crewNeverWinWith'
	| 'crewBalance'
	| 'cardInLastTrick'
	// Deep Sea
	| 'cmpTricks'
	| 'cmpCaptain'
	| 'predict'
	| 'avoidTricks'
	| 'noConsecutive'
	| 'consecutive'
	| 'winWith'
	| 'winTrickWhere'
	| 'winCount'
	| 'winEachColour'
	| 'winAllOfAnyColour'
	| 'neverLead'
	| 'equalSuits'
	| 'moreSuit'
	| 'all'
	| 'leaderStaysAhead'
	| 'lastCardPlayed';
/** The relation of a Deep Sea objective; "" when the kind has none. */
export type Comparison = '' | 'more' | 'fewer' | 'equal' | 'moreThanAll' | 'exactly' | 'atLeast';
/** What every card of a trick must meet (`winTrickWhere`); "" otherwise. */
export type TrickCondition =
	'' | 'below' | 'above' | 'even' | 'odd' | 'sumAtLeast' | 'sumBelow' | 'sumBetween';
/** Selects cards: a suit ("" = any colour, trump = submarines) and/or a rank (0 = any). */
export interface CardMatch {
	suit: Suit | '';
	rank: number;
}
/** The whole crew owns an objective. */
export const OWNER_CREW = -1;
/** An objective's owner is not chosen yet. */
export const OWNER_UNASSIGNED = -2;
/** In `Objective.tricks`: the last trick. */
export const LAST_TRICK = -1;

/** A task or special objective. Only the fields its kind uses are meaningful. */
export interface Objective {
	kind: ObjectiveKind;
	/** A seat, OWNER_CREW or OWNER_UNASSIGNED. */
	owner: number;
	cards: Card[];
	count: number;
	/** 1-based trick numbers; LAST_TRICK = the last trick. */
	tricks: number[];
	suit: Suit | '';
	rank: number;
	notWithTrump: boolean;
	ordered: boolean;

	// Deep Sea. Empty, zero or null when the kind does not use the field.
	cmp: Comparison;
	/** winCount, neverLead, equalSuits, moreSuit. */
	suits: Suit[];
	/** winCount: colour cards of these ranks. */
	ranks: number[];
	/** winWith: the winning card. */
	with: CardMatch | null;
	/** winWith: another card in the trick. */
	target: CardMatch | null;
	/** winTrickWhere, with `count` (and `max` for sumBetween). */
	where: TrickCondition;
	max: number;
	/** consecutive: exactly `count` tricks in total, all in a row. */
	exact: boolean;
	/** equalSuits: within one trick. */
	inOneTrick: boolean;
	/** predict: the prediction is secret until the debrief. */
	hidden: boolean;
	/** all: every part (same owner). */
	parts: Objective[];
}
export type OwnerRule = 'crew' | 'commander' | 'chosen' | 'roleA' | 'roleB' | 'leftOfHolder';
/** A special objective of a mission: a template plus the rule for its owner. */
export interface ObjectiveSpec extends Objective {
	ownerRule: OwnerRule;
	/** leftOfHolder: the card whose holder's left neighbour owns it. */
	holder: Card | null;
	/** The objective's card is the card of the task with this token. */
	taskToken?: Token;
}
export type Distribution =
	| ''
	| 'open'
	| 'commanderDecision'
	| 'commanderDistribution'
	// Deep Sea
	| 'draft'
	| 'oneMember'
	| 'captainOrVolunteer'
	| 'volunteers'
	| 'free';
/** Deep Sea communication mode of a mission: "" = normal. */
export type CommMode = '' | 'currents' | 'rapture' | 'unfamiliar' | 'none';
/** The mode of an attempt: unfamiliar terrain has been resolved by the terrain card. */
export type AttemptComm = Exclude<CommMode, 'unfamiliar'>;
export type SetupAction = 'swapTokens' | 'moveToken' | 'silenceMember' | 'drawAfterFirstTrick';
export type PollKind = 'yesNo' | 'wellIll';
/** One row of the mission table. */
export interface MissionDef {
	number: number;
	tasks: number;
	/** In placement order. */
	tokens: Token[];
	distribution: Distribution;
	deadZone: boolean;
	/** Disruption: nobody communicates before this trick; 0 = no restriction. */
	fromTrick: number;
	golden: boolean;
	setup: SetupAction[];
	poll: PollKind | '';
	objectives: ObjectiveSpec[];

	// Deep Sea.
	/** "" = normal. */
	commMode: CommMode;
	/** Total difficulty of the drawn task cards; 0 = none drawn. */
	difficulty: number;
	/** Distribution "volunteers": how many members take the tasks. */
	volunteers: number;
	/** The captain receives no task. */
	captainSkipped: boolean;
	/** The captain's first task is one of the highest difficulty. */
	hardestFirst: boolean;
	/** Task card numbers laid out instead of drawing. */
	fixedTasks: number[];
	/** The logbook counts no attempts. */
	uncounted: boolean;
}
/** GET /api/missions. */
export interface MissionCatalogue {
	edition: string;
	missions: MissionDef[];
}

/** "" in a dead zone. */
export type Position = 'highest' | 'only' | 'lowest' | '';
export interface Communication {
	card: Card;
	position: Position;
	/** The trick it was made before. */
	trick: number;
}
/** A stack of the dummy: the face-up card (null once played), and whether a hidden card lies under it. */
export interface Stack {
	up: Card | null;
	hidden: boolean;
}
export interface Member {
	seat: number;
	/** "" for the dummy. */
	playerId: string;
	/** "" for the dummy. */
	name: string;
	dummy: boolean;
	cardCount: number;
	tricksWon: number;
	/** The first communication while its card is still held. */
	communication: Communication | null;
	/** All communications whose card is still held, in order (rapture of the deep allows several). */
	communications: Communication[];
	/** Once per attempt. */
	communicated: boolean;
	/** The dummy's stacks; [] for a human. */
	stacks: Stack[];
	mustAct: boolean;
	/** Null until the debrief; then every card still held. */
	cards: Card[] | null;
}
export type Status = 'open' | 'fulfilled' | 'failed';
export interface Task {
	id: number;
	/** Null while face down and for Deep Sea task cards. */
	card: Card | null;
	token: Token;
	/** A seat, or OWNER_UNASSIGNED. */
	owner: number;
	status: Status;
	faceDown: boolean;

	// Deep Sea.
	/** The task card number 1..96; 0 for Planet Nine and while face down. */
	cardId: number;
	/** Shown even face down; 0 for Planet Nine. */
	difficulty: number;
	/**
	 * The card's objective; null for Planet Nine and while face down. The objective of a hidden
	 * prediction of another member has count -1 until the debrief.
	 */
	objective: Objective | null;
	/** A prediction task: the number has been noted. */
	predicted: boolean;
}
export interface ObjectiveView extends Objective {
	status: Status;
}
export interface Poll {
	kind: PollKind;
	/** One entry per seat: null = not answered or not asked. */
	answers: (boolean | null)[];
	/** The objective whose owner is chosen, else -1. */
	objective: number;
	/** Commander's distribution: the revealed task, else -1. */
	task: number;
}
export type Direction = 'left' | 'right';
export interface Pass {
	direction: Direction;
	/** The viewer's own choice. */
	chosen: Card | null;
	/** Seats that have not chosen yet. */
	waiting: number[];
}
/** The viewer's side of the last pass or draw. */
export interface Exchange {
	kind: 'pass' | 'draw';
	gave: Card | null;
	got: Card | null;
}
export interface CommOption {
	card: Card;
	position: Position;
}
export interface Actions {
	/** Empty unless the viewer may communicate now. */
	communicate: CommOption[];
	sendDistress: boolean;
	handover: boolean;
	/** Deep Sea: the viewer may pass in the draft now. */
	pass: boolean;
	/** Deep Sea: the viewer is the captain and may replace displayed tasks. */
	replaceTasks: boolean;
}
export interface Play {
	seat: number;
	card: Card;
}
export interface CompletedTrick {
	leader: number;
	plays: Play[];
	winner: number;
}
export type Reason = 'completed' | 'taskFailed' | 'tokenOrder' | 'objectiveFailed' | 'gaveUp';
export interface AttemptResult {
	success: boolean;
	reason: Reason;
	/** -1 when not applicable (also objective, seat). */
	task: number;
	objective: number;
	trick: number;
	seat: number;
}
/** "" in the debrief. */
export type Step =
	'adjustTokens' | 'poll' | 'pick' | 'draft' | 'assign' | 'predict' | 'pass' | 'play' | '';
export type PickKind =
	'silence' | 'roles' | 'objective' | 'decision' | 'distribute' | 'oneMember' | 'volunteers' | '';
export interface Attempt {
	number: number;
	/** The viewer's seat. */
	seat: number;
	commander: number;
	step: Step;
	pick: PickKind;
	actors: number[];
	/** Play: the seat to play; draft: the seat to take a task; else -1. */
	turn: number;
	trickNumber: number;
	totalTricks: number;
	members: Member[];
	hand: Card[];
	/** The viewer's own, or the dummy's when the viewer acts for it. */
	playable: Card[];
	tasks: Task[];
	objectives: ObjectiveView[];
	/** Mission 11: the silenced seat, else -1. */
	silenced: number;
	/** Step pick, for the commander: the seats they may pick. */
	pickable: number[];
	poll: Poll | null;
	pass: Pass | null;
	exchange: Exchange | null;
	/** In play order. */
	trick: Play[];
	trickLeader: number;
	lastTrick: CompletedTrick | null;
	distress: boolean;
	actions: Actions;
	result: AttemptResult | null;

	// Deep Sea.
	/** This attempt's communication mode (never "unfamiliar": the terrain card has decided). */
	comm: AttemptComm;
	/** Unfamiliar terrain: the colour card that picked the mode, else null. */
	terrain: Card | null;
	/** Rapture of the deep: shared sonar tokens left, else -1. */
	tokensLeft: number;
	/** Communication ends with the first trick (10, 13 when the tasks were passed on). */
	commFirstTrickOnly: boolean;
	/** The members the captain chose to take the tasks. */
	chosen: number[];
	/** Automatic re-deals before this attempt. */
	redeals: number;
	/** Step draft, for the one who picks: the task ids they may take. */
	takeable: number[];
	/** Step assign, for the captain: per task id the seats it may go to. */
	assignable: number[][];
	/** Step predict: the task id being predicted, else -1. */
	predict: number;
}
/** The host-selectable lobby settings. */
export interface GameOptions {
	/** The distress signal (rules §9) is on; off by default. */
	distress: boolean;
}
export interface PlayerView {
	gameId: string;
	revision: number;
	phase: Phase;
	you: string;
	hostId: string;
	players: PlayerSummary[];
	edition: string;
	options: GameOptions;
	/** Selected (lobby) or being played. */
	mission: number;
	missionDef: MissionDef | null;
	/** Sorted by mission. */
	logbook: LogEntry[];
	/** Null in the lobby. */
	attempt: Attempt | null;
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
/** Events are public: they never carry a hidden card. */
export type GameEvent =
	| { type: 'playerJoined'; playerId: string; name: string }
	| { type: 'playerLeft'; playerId: string; hostId: string }
	| { type: 'missionSelected'; mission: number }
	| { type: 'optionsChanged' }
	| { type: 'attemptStarted'; mission: number; number: number; commander: number }
	| { type: 'tokensAdjusted'; from: number; to: number }
	| { type: 'pollAnswered'; seat: number; yes: boolean }
	| { type: 'picked'; seats: number[] }
	| { type: 'taskPicked'; taskId: number; seat: number }
	| { type: 'distressSignal'; direction?: Direction; skipped: boolean }
	| { type: 'passCardChosen'; seat: number }
	| { type: 'cardsPassed'; direction: Direction }
	| { type: 'taskHandedOver'; taskId: number; from: number; to: number }
	| { type: 'communicated'; seat: number; card: Card; position?: Position }
	| { type: 'cardPlayed'; seat: number; card: Card }
	| { type: 'trickWon'; number: number; seat: number }
	| { type: 'cardsDrawn' }
	| { type: 'redealt'; times: number; commander: number }
	| { type: 'taskReplaced'; taskId: number }
	| { type: 'taskPassed'; seat: number }
	/** `tricks` is -1 for a hidden prediction. */
	| { type: 'predicted'; taskId: number; seat: number; tricks: number }
	| { type: 'attemptEnded'; success: boolean; reason: Reason }
	| { type: 'backInLobby' }
	| { type: 'gameEnded'; byPlayerId: string };
/** The Crew's own error codes, next to the ones every game shares (shared/frontend wire.ts). */
export type GameErrorCode =
	| 'not_host'
	| 'not_commander'
	| 'wrong_phase'
	| 'wrong_step'
	| 'game_full'
	| 'name_taken'
	| 'duplicate_player'
	| 'not_your_turn'
	| 'cannot_leave'
	| 'invalid_name'
	| 'unknown_player'
	| 'not_enough_players'
	| 'invalid_mission'
	| 'card_not_in_hand'
	| 'must_follow_suit'
	| 'cannot_communicate'
	| 'already_communicated'
	| 'invalid_position'
	| 'invalid_seat'
	| 'invalid_task'
	| 'uneven_distribution'
	| 'cannot_pass_trump'
	| 'already_answered'
	| 'not_allowed'
	| 'invalid_direction'
	| 'invalid_token_move'
	| 'invalid_result_choice'
	| 'invalid_prediction';
export type ApiErrorCode = CommonErrorCode | ClientErrorCode | GameErrorCode;
export type ErrorBody = SharedErrorBody<ApiErrorCode>;
export type ServerMessage = SharedServerMessage<PlayerView, GameEvent, ApiErrorCode>;
// sent as ClientCommand & { requestId: string }
export type ClientCommand =
	| { type: 'setOptions'; options: GameOptions }
	| { type: 'selectMission'; mission: number }
	| { type: 'startMission' }
	| { type: 'adjustTokens'; from: number; to: number }
	| { type: 'skipTokens' }
	| { type: 'answerPoll'; yes: boolean }
	| { type: 'pick'; seats: number[] }
	| { type: 'pickTask'; taskId: number }
	| { type: 'sendDistress'; direction: Direction }
	| { type: 'skipDistress' }
	| { type: 'choosePassCard'; card: Card }
	| { type: 'handover'; taskId: number; seat: number }
	| { type: 'communicate'; card: Card; position: Position }
	| { type: 'playCard'; card: Card }
	| { type: 'giveUp' }
	| { type: 'passTask' }
	| { type: 'replaceTask'; taskId: number }
	| { type: 'predict'; taskId: number; tricks: number }
	| { type: 'assignTask'; taskId: number; seat: number }
	/** Deep Sea: after a loss, draw new tasks and/or do not count the attempt; absent = false. */
	| { type: 'nextAttempt'; newTasks?: boolean; uncounted?: boolean }
	| { type: 'nextMission' }
	/** Deep Sea: after a loss, do not count the attempt; absent = false. */
	| { type: 'backToLobby'; uncounted?: boolean }
	| { type: 'endGame' };
