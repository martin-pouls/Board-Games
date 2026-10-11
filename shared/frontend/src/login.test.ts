import { beforeEach, describe, expect, it, vi } from 'vitest';
import { checkLogin, reloadForLogin, resetLoginReload } from './login.js';

const json = (status: number, body: unknown) =>
	Promise.resolve(new Response(JSON.stringify(body), { status }));
const errBody = (code: string) => ({ error: { code, message: code } });

beforeEach(resetLoginReload);

describe('reloadForLogin', () => {
	it('reloads only once', () => {
		const reload = vi.fn();
		reloadForLogin(reload);
		reloadForLogin(reload);
		expect(reload).toHaveBeenCalledTimes(1);
	});

	it('is a no-op without location', () => {
		reloadForLogin();
		expect(typeof location).toBe('undefined');
	});
});

describe('checkLogin', () => {
	it('reloads on 401 login_required', async () => {
		const reload = vi.fn();
		const fetchFn = vi.fn(() => json(401, errBody('login_required')));
		expect(await checkLogin('/base', fetchFn, reload)).toBe(true);
		expect(fetchFn).toHaveBeenCalledWith('/base/api/health');
		expect(await checkLogin('/base', fetchFn, reload)).toBe(true);
		expect(reload).toHaveBeenCalledTimes(1);
	});

	it('does not reload on 401 unauthorized', async () => {
		const reload = vi.fn();
		expect(await checkLogin('', () => json(401, errBody('unauthorized')), reload)).toBe(false);
		expect(reload).not.toHaveBeenCalled();
	});

	it('does not reload on 200', async () => {
		const reload = vi.fn();
		expect(await checkLogin('', () => json(200, { status: 'ok' }), reload)).toBe(false);
		expect(reload).not.toHaveBeenCalled();
	});

	it('returns false on network errors and non-JSON bodies', async () => {
		const reload = vi.fn();
		expect(await checkLogin('', () => Promise.reject(new TypeError('x')), reload)).toBe(false);
		expect(
			await checkLogin('', () => Promise.resolve(new Response('nope', { status: 401 })), reload)
		).toBe(false);
		expect(reload).not.toHaveBeenCalled();
	});
});
