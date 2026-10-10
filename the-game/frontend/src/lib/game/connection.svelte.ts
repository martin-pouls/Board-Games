import { SvelteMap } from 'svelte/reactivity';
import type {
	ApiErrorCode,
	ClientCommand,
	GameEvent,
	GameOptions,
	PlayerSummary,
	PlayerView
} from '#lib/types.js';
import { nextBackoff, parseServerMessage, shouldAcceptView, wsUrl } from './protocol.js';

export type ConnectionStatus = 'connecting' | 'open' | 'reconnecting' | 'closed';

export class ActionError extends Error {
	code: ApiErrorCode;
	constructor(code: ApiErrorCode, message?: string) {
		super(message ?? code);
		this.name = 'ActionError';
		this.code = code;
	}
}

const REQUEST_TIMEOUT_MS = 10_000;

interface Pending {
	resolve: () => void;
	reject: (e: ActionError) => void;
	timer: ReturnType<typeof setTimeout>;
}

export class GameConnection {
	view = $state<PlayerView | null>(null);
	online = $state<string[]>([]);
	status = $state<ConnectionStatus>('connecting');
	fatal = $state<'unauthorized' | 'game_not_found' | null>(null);
	lastEvents = $state<GameEvent[]>([]);

	readonly code: string;
	private token: string;
	private ws: WebSocket | null = null;
	private nextId = 1;
	private pending = new SvelteMap<string, Pending>();
	private attempt = 0;
	private retryTimer: ReturnType<typeof setTimeout> | null = null;
	private closed = false;
	private listening = false;

	constructor(code: string, token: string) {
		this.code = code;
		this.token = token;
	}

	get me(): PlayerSummary | undefined {
		return this.view?.players.find((p) => p.isYou);
	}
	get isHost(): boolean {
		return !!this.view && this.view.hostId === this.view.you;
	}
	get isMyTurn(): boolean {
		return this.view?.phase === 'playing' && this.view.turn?.activePlayerId === this.view.you;
	}

	connect(): void {
		if (this.closed || this.ws) return;
		this.addListeners();
		this.open();
	}

	close(): void {
		this.closed = true;
		this.removeListeners();
		this.clearRetry();
		const ws = this.ws;
		this.ws = null;
		if (ws) {
			ws.onopen = ws.onmessage = ws.onclose = ws.onerror = null;
			try {
				ws.close();
			} catch {
				// ignore
			}
		}
		this.rejectAll('disconnected');
		this.status = 'closed';
	}

	setOptions(options: GameOptions): Promise<void> {
		return this.send({ type: 'setOptions', options });
	}
	startGame(): Promise<void> {
		return this.send({ type: 'startGame' });
	}
	chooseFirstPlayer(playerId: string): Promise<void> {
		return this.send({ type: 'chooseFirstPlayer', playerId });
	}
	playCard(card: number, pile: number): Promise<void> {
		return this.send({ type: 'playCard', card, pile });
	}
	endTurn(): Promise<void> {
		return this.send({ type: 'endTurn' });
	}
	rematch(): Promise<void> {
		return this.send({ type: 'rematch' });
	}

	private send(cmd: ClientCommand): Promise<void> {
		const ws = this.ws;
		if (!ws || ws.readyState !== WebSocket.OPEN || this.status !== 'open') {
			return Promise.reject(new ActionError('disconnected'));
		}
		const requestId = String(this.nextId++);
		return new Promise<void>((resolve, reject) => {
			const timer = setTimeout(() => {
				this.pending.delete(requestId);
				reject(new ActionError('timeout'));
			}, REQUEST_TIMEOUT_MS);
			this.pending.set(requestId, { resolve, reject, timer });
			try {
				ws.send(JSON.stringify({ ...cmd, requestId }));
			} catch {
				clearTimeout(timer);
				this.pending.delete(requestId);
				reject(new ActionError('disconnected'));
			}
		});
	}

	private rejectAll(code: ApiErrorCode) {
		for (const p of this.pending.values()) {
			clearTimeout(p.timer);
			p.reject(new ActionError(code));
		}
		this.pending.clear();
	}

	private open() {
		this.status = this.attempt === 0 && !this.view ? 'connecting' : 'reconnecting';
		let ws: WebSocket;
		try {
			ws = new WebSocket(wsUrl(globalThis.location, this.code, this.token));
		} catch {
			this.scheduleReconnect();
			return;
		}
		this.ws = ws;
		ws.onopen = () => {
			if (this.ws !== ws) return;
			this.attempt = 0;
			this.status = 'open';
		};
		ws.onmessage = (ev) => {
			if (this.ws !== ws || typeof ev.data !== 'string') return;
			this.handle(ev.data);
		};
		ws.onclose = (ev) => {
			if (this.ws !== ws) return;
			this.ws = null;
			this.rejectAll('disconnected');
			if (ev.code === 4401) return this.fail('unauthorized');
			if (ev.code === 4404) return this.fail('game_not_found');
			this.scheduleReconnect();
		};
	}

	private fail(reason: 'unauthorized' | 'game_not_found') {
		this.fatal = reason;
		this.closed = true;
		this.removeListeners();
		this.clearRetry();
		this.status = 'closed';
	}

	private handle(data: string) {
		const msg = parseServerMessage(data);
		if (!msg) return;
		switch (msg.type) {
			case 'view':
				if (shouldAcceptView(this.view, msg.view)) {
					this.view = msg.view;
					this.lastEvents = msg.events;
				}
				break;
			case 'presence':
				this.online = msg.online;
				break;
			case 'ack': {
				const p = this.pending.get(msg.requestId);
				if (p) {
					clearTimeout(p.timer);
					this.pending.delete(msg.requestId);
					p.resolve();
				}
				break;
			}
			case 'error': {
				const p = msg.requestId ? this.pending.get(msg.requestId) : undefined;
				if (p && msg.requestId) {
					clearTimeout(p.timer);
					this.pending.delete(msg.requestId);
					p.reject(new ActionError(msg.error.code, msg.error.message));
				}
				break;
			}
		}
	}

	private clearRetry() {
		if (this.retryTimer) clearTimeout(this.retryTimer);
		this.retryTimer = null;
	}

	private scheduleReconnect() {
		if (this.closed) return;
		this.status = 'reconnecting';
		this.clearRetry();
		const delay = nextBackoff(this.attempt++);
		this.retryTimer = setTimeout(() => {
			this.retryTimer = null;
			if (!this.closed && !this.ws) this.open();
		}, delay);
	}

	private reconnectNow = () => {
		if (this.closed || this.ws) return;
		this.clearRetry();
		this.attempt = 0;
		this.open();
	};
	private onVisibility = () => {
		if (document.visibilityState === 'visible') this.reconnectNow();
	};

	private addListeners() {
		if (this.listening || typeof window === 'undefined') return;
		this.listening = true;
		document.addEventListener('visibilitychange', this.onVisibility);
		window.addEventListener('online', this.reconnectNow);
	}
	private removeListeners() {
		if (!this.listening) return;
		this.listening = false;
		document.removeEventListener('visibilitychange', this.onVisibility);
		window.removeEventListener('online', this.reconnectNow);
	}
}
