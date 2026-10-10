# syntax=docker/dockerfile:1
# Static start page: prerendered SvelteKit build served by a plain Caddy file server. Built
# natively on the (arm64) server by deploy/compose.yaml; the shared Caddy proxies to it.

FROM node:24-alpine AS web
WORKDIR /src/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci --no-audit --no-fund
COPY frontend/ ./
# Sub-path the page is served under (see vite.config.ts); empty = site root.
ARG BASE_PATH=""
RUN BASE_PATH="$BASE_PATH" npm run build

FROM caddy:2-alpine
COPY --from=web /src/frontend/build /srv
EXPOSE 8080
CMD ["caddy", "file-server", "--root", "/srv", "--listen", ":8080"]
