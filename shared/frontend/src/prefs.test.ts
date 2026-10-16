import { describe, expect, it } from 'vitest';
import { createPref, isRecord, type PrefStorageLike } from './prefs.js';

function memoryStorage(): PrefStorageLike & { m: Map<string, string> } {
	const m = new Map<string, string>();
	return { m, getItem: (k) => m.get(k) ?? null, setItem: (k, v) => void m.set(k, v) };
}

const parseLevel = (v: unknown) => (v === 1 || v === 2 || v === 3 ? v : null);

describe('pref', () => {
	it('is null until saved, then round-trips', () => {
		const s = memoryStorage();
		const p = createPref('test:level', parseLevel, () => s);
		expect(p.load()).toBeNull();
		p.save(2);
		expect(s.m.get('test:level')).toBe('2');
		expect(p.load()).toBe(2);
	});

	it('drops values parse rejects and unparsable JSON', () => {
		const s = memoryStorage();
		const p = createPref('test:level', parseLevel, () => s);
		s.m.set('test:level', '7');
		expect(p.load()).toBeNull();
		s.m.set('test:level', '{oops');
		expect(p.load()).toBeNull();
	});

	it('behaves as empty when storage is missing or throws', () => {
		const throwing: PrefStorageLike = {
			getItem: () => {
				throw new Error('blocked');
			},
			setItem: () => {
				throw new Error('blocked');
			}
		};
		for (const storage of [() => undefined, () => throwing]) {
			const p = createPref('test:level', parseLevel, storage);
			expect(() => p.save(1)).not.toThrow();
			expect(p.load()).toBeNull();
		}
	});

	it('isRecord accepts only plain objects', () => {
		expect(isRecord({ a: 1 })).toBe(true);
		expect(isRecord(null)).toBe(false);
		expect(isRecord([])).toBe(false);
		expect(isRecord('x')).toBe(false);
	});
});
