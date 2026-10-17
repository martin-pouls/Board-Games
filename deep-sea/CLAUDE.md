# Deep Sea

"The Crew: Mission Deep Sea" at `/deep-sea/`. **This app has no code of its own.** It is [the-crew/](../the-crew/CLAUDE.md) built as the Deep Sea edition: the image comes from `the-crew/Dockerfile` with `BASE_PATH=/deep-sea`, `EDITION=deepSea` and `DB_FILE=deep-sea.db` (see [deploy/compose.yaml](deploy/compose.yaml)). This directory holds only `deploy/` (compose file and Caddy snippet) and this file.

- Game code, rules and look: [the-crew/CLAUDE.md](../the-crew/CLAUDE.md). Deep Sea plan, rules and mission data: [the-crew/docs/deep-sea/](../the-crew/docs/deep-sea/).
- Its own URL, container (`deep-sea`), volume and database (`deep-sea.db`), game retention (720h) and start-page tile.
- Local ports: Vite `:5179`, backend `:8085`.
- CI redeploys this app when `deep-sea/`, `the-crew/` or `shared/` changes: `.github/scripts/plan-deploy.sh` follows the `COPY` sources of the Dockerfile named in `deploy/compose.yaml`.
