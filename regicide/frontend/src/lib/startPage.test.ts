import { describe, expect, it } from 'vitest';
import { privacyHref, startPageHref } from './startPage.js';

describe('startPageHref', () => {
	it('uses the start page path on the same host in production', () => {
		expect(startPageHref(new URL('https://yoshi-und-kruemli.de/regicide/'), false)).toBe(
			'/card-games/'
		);
	});

	it("uses the start page's dev port on the same hostname in dev", () => {
		expect(startPageHref(new URL('http://localhost:5173/'), true)).toBe('http://localhost:5175/');
		expect(startPageHref(new URL('http://192.168.1.20:5174/'), true)).toBe(
			'http://192.168.1.20:5175/'
		);
	});
});

describe('privacyHref', () => {
	it('points to the privacy page of the start page', () => {
		expect(privacyHref(new URL('https://yoshi-und-kruemli.de/the-game/'), false)).toBe(
			'/card-games/datenschutz/'
		);
		expect(privacyHref(new URL('http://localhost:5173/'), true)).toBe(
			'http://localhost:5175/datenschutz/'
		);
	});
});
