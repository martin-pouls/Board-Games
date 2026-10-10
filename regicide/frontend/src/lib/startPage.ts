/** Port of the start page's Vite dev server (`server.port` in start-page/frontend/vite.config.ts). */
export const START_PAGE_DEV_PORT = 5175;

/**
 * Where the link back to the overview of all games points. In production every app sits on the
 * same host behind Caddy, so it is the start page's path. In dev the start page runs its own Vite
 * server, so it is that port on the hostname this app was opened with (localhost or a LAN IP).
 */
export function startPageHref(url: Pick<URL, 'protocol' | 'hostname'>, dev: boolean): string {
	return dev ? `${url.protocol}//${url.hostname}:${START_PAGE_DEV_PORT}/` : '/card-games/';
}

/** The privacy policy (Datenschutzerklärung), a page of the start page app. */
export function privacyHref(url: Pick<URL, 'protocol' | 'hostname'>, dev: boolean): string {
	return `${startPageHref(url, dev)}datenschutz/`;
}
