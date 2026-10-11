// Imported by each game's vite.config.ts: how an app compiles the shared frontend code.
//
// The apps import shared code as source through the alias `@board-games/shared` (Vite
// `resolve.alias`, mirrored by `paths` in the app's tsconfig.json). Every package the shared code imports is a
// peerDependency here and resolved from the app (`resolve.dedupe`), so there is one copy of
// svelte, bits-ui, mode-watcher, svelte-sonner, ... per app. node_modules of this directory is
// only for its own check, lint and test. Docker builds do not install it.
import { fileURLToPath } from 'node:url';
import pkg from './package.json' with { type: 'json' };

/** The alias apps import shared code under, e.g. `@board-games/shared/session.js`. */
export const sharedAlias = '@board-games/shared';

/** The source directory the alias points to. */
export const sharedSrc = fileURLToPath(new URL('./src', import.meta.url));

/** This directory (for `server.fs.allow`). */
export const sharedRoot = fileURLToPath(new URL('.', import.meta.url));

/** Packages the shared code imports; the app resolves them from its own node_modules. */
export const sharedDedupe = Object.keys(pkg.peerDependencies);
