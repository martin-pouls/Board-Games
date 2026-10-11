import { beforeEach, describe, expect, it, vi } from 'vitest';
import { ApiError, createRequest } from './api.js';
import { resetLoginReload } from './login.js';

const reply = (status: number, code: string) => () =>
	Promise.resolve(new Response(JSON.stringify({ error: { code, message: code } }), { status }));

beforeEach(resetLoginReload);

describe('createRequest', () => {
	it('reloads the page on login_required and still throws', async () => {
		const reload = vi.fn();
		vi.stubGlobal('location', { reload });
		try {
			const request = createRequest('', reply(401, 'login_required'));
			await expect(request('GET', '/api/x')).rejects.toMatchObject({
				status: 401,
				code: 'login_required'
			});
			expect(reload).toHaveBeenCalledTimes(1);
		} finally {
			vi.unstubAllGlobals();
		}
	});

	it('does not reload on a bad player token', async () => {
		const reload = vi.fn();
		vi.stubGlobal('location', { reload });
		try {
			const request = createRequest('', reply(401, 'unauthorized'));
			const err = (await request('GET', '/api/x').catch((e: unknown) => e)) as ApiError;
			expect(err).toBeInstanceOf(ApiError);
			expect(err.code).toBe('unauthorized');
			expect(reload).not.toHaveBeenCalled();
		} finally {
			vi.unstubAllGlobals();
		}
	});
});
