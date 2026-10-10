import { describe, expect, it } from 'vitest';
import { createSessionStore, type SessionStorageLike } from './session.js';

function memoryStorage(): SessionStorageLike {
	const m = new Map<string, string>();
	return {
		getItem: (k) => m.get(k) ?? null,
		setItem: (k, v) => void m.set(k, v),
		removeItem: (k) => void m.delete(k),
		key: (i) => [...m.keys()][i] ?? null,
		get length() {
			return m.size;
		}
	};
}

const s1 = { token: 't1', playerId: 'p_1', name: 'Anna' };

describe('session store', () => {
	it('saves, gets and clears case-insensitively', () => {
		const st = createSessionStore(memoryStorage());
		expect(st.get('K7MQ2')).toBeNull();
		st.save('k7mq2', s1);
		expect(st.get('K7MQ2')).toEqual(s1);
		st.clear('K7MQ2');
		expect(st.get('K7MQ2')).toBeNull();
	});

	it('lists only session keys', () => {
		const st = createSessionStore(memoryStorage());
		st.setNickname('Anna');
		st.save('AAAAA', s1);
		st.save('BBBBB', { ...s1, token: 't2' });
		const list = st.list();
		expect(list.map((x) => x.code).sort()).toEqual(['AAAAA', 'BBBBB']);
	});

	it('stores the nickname', () => {
		const st = createSessionStore(memoryStorage());
		expect(st.getNickname()).toBe('');
		st.setNickname('Ben');
		expect(st.getNickname()).toBe('Ben');
	});

	it('copies a session to a new code', () => {
		const st = createSessionStore(memoryStorage());
		st.save('AAAAA', s1);
		st.copy('AAAAA', 'BBBBB');
		expect(st.get('BBBBB')).toEqual(s1);
		st.copy('ZZZZZ', 'CCCCC');
		expect(st.get('CCCCC')).toBeNull();
	});

	it('ignores corrupt data and unavailable storage', () => {
		const mem = memoryStorage();
		mem.setItem('regicide:session:AAAAA', 'not json');
		expect(createSessionStore(mem).get('AAAAA')).toBeNull();
		const broken = createSessionStore(() => {
			throw new Error('denied');
		});
		expect(broken.get('AAAAA')).toBeNull();
		expect(broken.list()).toEqual([]);
		expect(broken.getNickname()).toBe('');
		expect(() => broken.save('AAAAA', s1)).not.toThrow();
		expect(createSessionStore(() => undefined).getNickname()).toBe('');
	});
});
