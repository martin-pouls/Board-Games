import { readFileSync } from 'node:fs';
import { describe, expect, it } from 'vitest';
import { cap, edition, editionTerms, terms } from './edition.js';

describe('edition terms', () => {
	it('has the same keys for both editions', () => {
		expect(Object.keys(editionTerms.deepSea).sort()).toEqual(
			Object.keys(editionTerms.planetNine).sort()
		);
	});

	it('has no empty term', () => {
		for (const t of Object.values(editionTerms)) {
			for (const [key, value] of Object.entries(t)) {
				expect(value, key).toBeTruthy();
			}
		}
	});

	it('differs between the editions where the game names things', () => {
		const [p, d] = [editionTerms.planetNine, editionTerms.deepSea];
		for (const key of ['title', 'trump', 'trumps', 'dummy', 'commander', 'commToken'] as const) {
			expect(p[key], key).not.toBe(d[key]);
		}
	});

	it('uses Planet Nine by default (the tests run without EDITION)', () => {
		expect(edition).toBe('planetNine');
		expect(terms).toBe(editionTerms.planetNine);
		expect(terms.title).toBe('The Crew');
	});

	it('keeps the theme colours of app.html in sync (inline script, before first paint)', () => {
		const html = readFileSync(new URL('../app.html', import.meta.url), 'utf8');
		for (const t of Object.values(editionTerms)) {
			expect(html).toContain(t.themeColor.light);
			expect(html).toContain(t.themeColor.dark);
		}
	});

	it('capitalises the first letter', () => {
		expect(cap('captain')).toBe('Captain');
		expect(cap('')).toBe('');
	});
});
