import { describe, expect, it } from 'vitest';
import { parseFlight } from './flightPref.js';

describe('parseFlight', () => {
	it('keeps a scenario with its abilities', () => {
		expect(
			parseFlight({ scenario: 'tgu-red', abilities: ['workingTogether', 'synchronisation'] })
		).toEqual({
			scenario: 'tgu-red',
			abilities: ['synchronisation', 'workingTogether']
		});
	});

	it('drops unknown abilities and duplicates', () => {
		expect(
			parseFlight({ scenario: 'yul-green', abilities: ['mastery', 'mastery', 'radar', 7] })
		).toEqual({
			scenario: 'yul-green',
			abilities: ['mastery']
		});
	});

	it('rejects anything that is not a flight', () => {
		for (const v of [null, 'tgu-red', [], {}, { scenario: 3, abilities: [] }, { scenario: 'a' }]) {
			expect(parseFlight(v)).toBeNull();
		}
	});
});
