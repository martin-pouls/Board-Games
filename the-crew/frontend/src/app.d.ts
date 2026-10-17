// See https://svelte.dev/docs/kit/types#app.d.ts
// for information about these interfaces
declare global {
	/** Build-time base path ('' in dev, e.g. '/the-crew' in production); set from BASE_PATH in vite.config.ts. */
	const __BASE_PATH__: string;
	/** Build-time edition ('planetNine' or 'deepSea'); set from EDITION (or the Vite mode) in vite.config.ts. */
	const __EDITION__: 'planetNine' | 'deepSea';

	namespace App {
		// interface Error {}
		// interface Locals {}
		// interface PageData {}
		// interface PageState {}
		// interface Platform {}
	}
}

export {};
