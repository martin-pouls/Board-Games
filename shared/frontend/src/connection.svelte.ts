import { SvelteMap } from 'svelte/reactivity';
import { checkLogin } from './login.js';
import { nextBackoff, parseServerMessage, shouldAcceptView, wsUrl } from './protocol.js';
import { CLOSE_GAME_NOT_FOUND, CLOSE_UNAUTHORIZED, type BaseView } from './wire.js';

export type ConnectionStatus = 'connecting' | 'open' | 'reconnecting' | 'closed';

/** A command the server refused (its error code), or a transport failure (client codes). */
export class ActionError<C extends string = string> extends Error {
	code: C;
	constructor(code: C, message?: string) {
		super(message ?? code);
		this.name = 'ActionError';
		this.code = code;
	}
}

const REQUEST_TIMEOUT_MS = 10_000;

interface Pending<C extends string> {
	resolve: () => void;
	reject: (e: ActionError<C>) => void;
	timer: ReturnType<typeof setTimeout>;
}

/**
 * The live connection to one game: the player's current view, who is online, and commands
 * acknowledged by the server. It reconnects with backoff (and at once when the tab becomes
 * visible or the browser goes online) and stops for good on 4401/4404. Each game subclasses it
 * with its typed commands, calling `send`.
 *
 * V is the game's player view, E its events, Cmd its client commands and C its error codes
 * (which must include the client codes 'disconnected' and 'timeout').
 */
export class GameSocket<V extends BaseView, E, Cmd extends { type: string }, C extends string> {
	view = $state<V | null>(null);
	online = $state<string[]>([]);
	status = $state<ConnectionStatus>('connecting');
	fatal = $state<'unauthorized' | 'game_not_found' | null>(null);
	lastEvents = $state<E[]>([]);

	readonly code: string;
	private token: string;
	private basePath: string;
	private ws: WebSocket | null = null;
	private nextId = 1;
	private pending = new SvelteMap<string, Pending<C>>();
	private attempt = 0;
	private retryTimer: ReturnType<typeof setTimeout> | null = null;
	private closed = false;
	private listening = false;

	/** basePath is the app's base path (`__BASE_PATH__`, '' in dev). */
	constructor(code: string, token: string, basePath: string) {
		this.code = code;
		this.token = token;
		this.basePath = basePath;
	}

	get me(): V['players'][number] | undefined {
		return this.view?.players.find((p) => p.isYou);
	}
	get isHost(): boolean {
		return !!this.view && this.view.hostId === this.view.you;
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
		this.rejectAll('disconnected' as C);
		this.status = 'closed';
	}

	/** Sends cmd and resolves on the server's ack (after the resulting view arrived). */
	protected send(cmd: Cmd): Promise<void> {
		const ws = this.ws;
		if (!ws || ws.readyState !== WebSocket.OPEN || this.status !== 'open') {
			return Promise.reject(new ActionError('disconnected' as C));
		}
		const requestId = String(this.nextId++);
		return new Promise<void>((resolve, reject) => {
			const timer = setTimeout(() => {
				this.pending.delete(requestId);
				reject(new ActionError('timeout' as C));
			}, REQUEST_TIMEOUT_MS);
			this.pending.set(requestId, { resolve, reject, timer });
			try {
				ws.send(JSON.stringify({ ...cmd, requestId }));
			} catch {
				clearTimeout(timer);
				this.pending.delete(requestId);
				reject(new ActionError('disconnected' as C));
			}
		});
	}

	private rejectAll(code: C) {
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
			ws = new WebSocket(wsUrl(globalThis.location, this.basePath, this.code, this.token));
		} catch {
			this.scheduleReconnect();
			return;
		}
		this.ws = ws;
		let opened = false;
		ws.onopen = () => {
			if (this.ws !== ws) return;
			opened = true;
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
			this.rejectAll('disconnected' as C);
			if (ev.code === CLOSE_UNAUTHORIZED) return this.fail('unauthorized');
			if (ev.code === CLOSE_GAME_NOT_FOUND) return this.fail('game_not_found');
			// A failed upgrade hides its HTTP status: probe whether the login expired.
			if (!opened) void checkLogin(this.basePath);
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
		const msg = parseServerMessage<V, E, C>(data);
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
