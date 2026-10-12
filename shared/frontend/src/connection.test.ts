import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { ActionError, GameSocket, STALE_AFTER_HIDDEN_MS } from './connection.svelte.js';
import type { BaseView } from './wire.js';

class FakeWebSocket {
	static readonly OPEN = 1;
	static instances: FakeWebSocket[] = [];
	readyState = 0;
	sent: string[] = [];
	closed = false;
	onopen: (() => void) | null = null;
	onmessage: ((ev: { data: unknown }) => void) | null = null;
	onclose: ((ev: { code: number }) => void) | null = null;
	onerror: (() => void) | null = null;
	constructor(readonly url: string) {
		FakeWebSocket.instances.push(this);
	}
	send(data: string) {
		this.sent.push(data);
	}
	close() {
		this.closed = true;
	}
	accept() {
		this.readyState = FakeWebSocket.OPEN;
		this.onopen?.();
	}
}

/** A minimal document: the visibility state and its listeners. */
const doc = {
	visibilityState: 'visible',
	listeners: [] as (() => void)[],
	addEventListener(_: string, fn: () => void) {
		this.listeners.push(fn);
	},
	removeEventListener(_: string, fn: () => void) {
		this.listeners = this.listeners.filter((l) => l !== fn);
	},
	setVisibility(state: 'visible' | 'hidden') {
		this.visibilityState = state;
		for (const l of this.listeners) l();
	}
};

type Cmd = { type: 'play' };
class TestSocket extends GameSocket<BaseView, never, Cmd, 'disconnected' | 'timeout'> {
	play() {
		return this.send({ type: 'play' });
	}
}

const sockets = () => FakeWebSocket.instances;
const last = () => sockets()[sockets().length - 1];

beforeEach(() => {
	vi.useFakeTimers();
	FakeWebSocket.instances = [];
	doc.visibilityState = 'visible';
	doc.listeners = [];
	vi.stubGlobal('WebSocket', FakeWebSocket);
	vi.stubGlobal('document', doc);
	vi.stubGlobal('window', { addEventListener() {}, removeEventListener() {} });
	vi.stubGlobal('location', { protocol: 'http:', host: 'localhost' });
});

afterEach(() => {
	vi.useRealTimers();
	vi.unstubAllGlobals();
});

function connected(): TestSocket {
	const s = new TestSocket('ABCD', 'tok', '');
	s.connect();
	last().accept();
	return s;
}

describe('GameSocket half-dead sockets', () => {
	it('replaces the socket when the tab was hidden for a while', () => {
		const s = connected();
		const first = last();
		doc.setVisibility('hidden');
		vi.advanceTimersByTime(STALE_AFTER_HIDDEN_MS);
		doc.setVisibility('visible');
		expect(first.closed).toBe(true);
		expect(sockets()).toHaveLength(2);
		expect(s.status).toBe('connecting');
		s.close();
	});

	it('keeps the socket after a short absence', () => {
		const s = connected();
		doc.setVisibility('hidden');
		vi.advanceTimersByTime(STALE_AFTER_HIDDEN_MS - 1);
		doc.setVisibility('visible');
		expect(sockets()).toHaveLength(1);
		expect(sockets()[0].closed).toBe(false);
		expect(s.status).toBe('open');
		s.close();
	});

	it('rejects pending commands when it replaces the socket', async () => {
		const s = connected();
		doc.setVisibility('hidden');
		vi.advanceTimersByTime(STALE_AFTER_HIDDEN_MS);
		const p = s.play();
		doc.setVisibility('visible');
		await expect(p).rejects.toEqual(new ActionError('disconnected'));
		s.close();
	});

	it('replaces the socket after a command timed out', async () => {
		const s = connected();
		const first = last();
		const p = s.play();
		vi.advanceTimersByTime(10_000);
		await expect(p).rejects.toEqual(new ActionError('timeout'));
		expect(first.closed).toBe(true);
		expect(sockets()).toHaveLength(2);
		s.close();
	});
});
