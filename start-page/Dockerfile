# syntax=docker/dockerfile:1
# Static start page: prerendered SvelteKit build served by a plain Caddy file server. Built
# for arm64 by CI (.github/workflows/deploy.yml) from deploy/compose.yaml, or on the server by the
# fallback deploy/deploy.sh; the shared Caddy proxies to it.
#
# The build context is the repository root (the page frame comes from shared/), so paths below are
# from the root. Dockerfile.dockerignore next to this file applies.

FROM node:24-alpine AS web
WORKDIR /src/start-page/frontend
COPY start-page/frontend/package.json start-page/frontend/package-lock.json ./
RUN npm ci --no-audit --no-fund
# Compiled as source with the app; its imports resolve from the app's node_modules.
COPY shared/frontend/ /src/shared/frontend/
COPY start-page/frontend/ ./
# Sub-path the page is served under (see vite.config.ts); empty = site root.
ARG BASE_PATH=""
RUN BASE_PATH="$BASE_PATH" npm run build

FROM caddy:2-alpine
# Links the image to the repo on ghcr.io; the server prunes old images by it.
LABEL org.opencontainers.image.source=https://github.com/martin-pouls/Board-Games
COPY --from=web /src/start-page/frontend/build /srv
EXPOSE 8080
CMD ["caddy", "file-server", "--root", "/srv", "--listen", ":8080"]
