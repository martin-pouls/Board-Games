import { describe, expect, it } from 'vitest';
import { startPageHref } from './startPage.js';

describe('startPageHref', () => {
	it('uses the start page path on the same host in production', () => {
		expect(startPageHref(new URL('https://yoshi-und-kruemli.de/regicide/'), false)).toBe(
			'/board-games/'
		);
	});

	it("uses the start page's dev port on the same hostname in dev", () => {
		expect(startPageHref(new URL('http://localhost:5173/'), true)).toBe('http://localhost:5175/');
		expect(startPageHref(new URL('http://192.168.1.20:5174/'), true)).toBe(
			'http://192.168.1.20:5175/'
		);
	});
});
