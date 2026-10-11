// The access gate in front of the apps (Caddy forward_auth) answers 401 `login_required` to
// requests without a valid login cookie. A page that is already open then reloads, so the
// browser hits the gate's login redirect.
import type { ErrorBody } from './wire.js';

export const LOGIN_REQUIRED = 'login_required';

let reloaded = false;

const browserReload = () => {
	if (typeof location !== 'undefined') location.reload();
};

/** Reloads the page, at most once per page lifetime. A no-op without `location` (tests, SSR). */
export function reloadForLogin(reload: () => void = browserReload): void {
	if (reloaded) return;
	reloaded = true;
	reload();
}

/** Re-arms reloadForLogin (tests only). */
export function resetLoginReload(): void {
	reloaded = false;
}

/**
 * Asks the server whether the login is still valid: GET /api/health. On 401 `login_required` it
 * reloads the page and returns true. Never throws.
 */
export async function checkLogin(
	basePath: string,
	fetchFn: typeof fetch = (...args) => fetch(...args),
	reload?: () => void
): Promise<boolean> {
	try {
		const res = await fetchFn(basePath + '/api/health');
		if (res.status !== 401) return false;
		const body = (await res.json()) as { error?: Partial<ErrorBody> };
		if (body.error?.code !== LOGIN_REQUIRED) return false;
		reloadForLogin(reload);
		return true;
	} catch {
		return false;
	}
}
