export type Phase = 'lobby' | 'playing' | 'won' | 'lost';
export type Step = 'play' | 'discard' | 'chooseNext' | '';
export type Suit = 'hearts' | 'diamonds' | 'clubs' | 'spades';
export type JesterSignal = 'none' | 'goodPlay' | 'ratherNot';
export type LossReason = 'cannotCoverDamage' | 'cannotAct' | 'gaveUp';
export type Grade = 'gold' | 'silver' | 'bronze';

/** suit '' + rank 0 is the Jester. Ranks: 1 Ace, 2..10, 11 J, 12 Q, 13 K. */
export interface Card {
	suit: Suit | '';
	rank: number;
}
export interface Rules {
	minPlayers: number;
	maxPlayers: number;
	maxHandSize: number;
	tavernJesters: number;
	soloJesters: number;
}
export interface PlayerSummary {
	id: string;
	name: string;
	seat: number;
	handCount: number;
	isHost: boolean;
	isYou: boolean;
	isActive: boolean;
	yieldedLastTurn: boolean;
	signal: JesterSignal;
	/** Only once the game is over. */
	revealedHand?: Card[];
}
export interface PlayedSet {
	playerId: string;
	cards: Card[];
	value: number;
	damage: number;
	doubled: boolean;
	/** Suits whose power is not in effect for this play (immunity). */
	blockedSuits: Suit[];
	jester: boolean;
}
export interface Enemy {
	card: Card;
	attack: number;
	health: number;
	damage: number;
	shield: number;
	effectiveAttack: number;
	immunity: Suit;
	immunityCancelled: boolean;
	plays: PlayedSet[];
}
export interface Turn {
	number: number;
	activePlayerId: string;
	step: Step;
	canYield: boolean;
	/** Effective attack to cover in the discard step, else 0. */
	damageToCover: number;
}
export interface Result {
	outcome: 'won' | 'lost';
	reason?: LossReason;
	playerId?: string;
	grade?: Grade;
	enemiesDefeated: number;
	soloJestersUsed: number;
}
export interface PlayerView {
	gameId: string;
	revision: number;
	phase: Phase;
	you: string;
	hostId: string;
	rules: Rules;
	players: PlayerSummary[];
	hand: Card[];
	enemy: Enemy | null;
	castleCount: number;
	tavernCount: number;
	discardCount: number;
	turn: Turn | null;
	soloJestersLeft: number;
	result: Result | null;
	nextGameId?: string;
}
export interface GameSummary {
	gameId: string;
	phase: Phase;
	hostName: string;
	playerCount: number;
	maxPlayers: number;
	joinable: boolean;
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
	| { type: 'gameStarted'; playerId: string }
	| { type: 'enemyRevealed'; card: Card }
	| { type: 'turnStarted'; playerId: string; number: number }
	| {
			type: 'cardsPlayed';
			playerId: string;
			cards: Card[];
			value: number;
			healed: number;
			drawn: number;
			damage: number;
			blockedSuits?: Suit[];
	  }
	| { type: 'jesterPlayed'; playerId: string }
	| { type: 'yielded'; playerId: string }
	| { type: 'enemyDefeated'; card: Card; exact: boolean }
	| { type: 'cardsDiscarded'; playerId: string; cards: Card[]; attack: number }
	| { type: 'attackBlocked'; playerId: string }
	| { type: 'nextPlayerChosen'; byPlayerId: string; playerId: string }
	| { type: 'signalSet'; playerId: string; signal: JesterSignal }
	| { type: 'soloJesterUsed'; playerId: string; drawn: number; left: number }
	| { type: 'gameWon'; grade?: Grade }
	| { type: 'gameLost'; reason: LossReason; playerId: string }
	| { type: 'rematchCreated'; gameId: string };
export type ApiErrorCode =
	| 'bad_request'
	| 'unauthorized'
	| 'not_host'
	| 'game_not_found'
	| 'not_found'
	| 'wrong_phase'
	| 'wrong_step'
	| 'game_full'
	| 'name_taken'
	| 'not_your_turn'
	| 'cannot_leave'
	| 'conflict'
	| 'invalid_name'
	| 'unknown_player'
	| 'card_not_in_hand'
	| 'illegal_play'
	| 'illegal_discard'
	| 'yield_not_allowed'
	| 'not_solo'
	| 'no_solo_jester'
	| 'illegal_signal'
	| 'rematch_exists'
	| 'internal'
	| 'network'
	| 'disconnected'
	| 'timeout'; // last three are client-side only
export interface ErrorBody {
	code: ApiErrorCode;
	message: string;
}
export type ServerMessage =
	| { type: 'view'; view: PlayerView; events: GameEvent[] }
	| { type: 'presence'; online: string[] }
	| { type: 'ack'; requestId: string }
	| { type: 'error'; requestId?: string; error: ErrorBody };
export type ClientCommand =
	| { type: 'startGame'; firstPlayerId?: string }
	| { type: 'play'; cards: Card[] }
	| { type: 'yield' }
	| { type: 'discard'; cards: Card[] }
	| { type: 'chooseNext'; playerId: string }
	| { type: 'useSoloJester' }
	| { type: 'setSignal'; signal: JesterSignal }
	| { type: 'giveUp' }
	| { type: 'rematch' };
// sent as ClientCommand & { requestId: string }
