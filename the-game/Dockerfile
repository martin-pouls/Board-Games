# syntax=docker/dockerfile:1
# Single image: Go server + built SPA. Built for arm64 by CI (.github/workflows/deploy.yml) from
# deploy/compose.yaml, or on the server by the fallback deploy/deploy.sh.
#
# The build context is the repository root (the app uses shared/), so paths below are from the
# root. The image keeps the repo layout (/src/the-game, /src/shared): the app finds shared/ at
# ../../shared as it does in a checkout. Dockerfile.dockerignore next to this file applies.

FROM node:25-alpine@sha256:bdf2cca6fe3dabd014ea60163eca3f0f7015fbd5c7ee1b0e9ccb4ced6eb02ef4 AS web
WORKDIR /src/the-game/frontend
COPY the-game/frontend/package.json the-game/frontend/package-lock.json ./
RUN npm ci --no-audit --no-fund
# Compiled as source with the app; its imports resolve from the app's node_modules.
COPY shared/frontend/ /src/shared/frontend/
COPY the-game/frontend/ ./
# Sub-path the SPA is served under (see vite.config.ts); empty = site root.
ARG BASE_PATH=""
RUN BASE_PATH="$BASE_PATH" npm run build

FROM golang:1.27-alpine@sha256:f92b6ef800e499660581efdabdf25d9d817a9d124eaf900924f0504e7e27e12d AS api
WORKDIR /src/the-game/backend
# go.mod replaces the shared module with ../../shared/backend.
COPY shared/backend/ /src/shared/backend/
COPY the-game/backend/go.mod the-game/backend/go.sum ./
RUN go mod download
COPY the-game/backend/ ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server \
	&& mkdir -p /out/data

FROM gcr.io/distroless/static-debian13:nonroot@sha256:e2e927ec666bae08560abb3c55d0659eceabb657f56b6782ab500a9fc7f555e3
# Links the image to the repo on ghcr.io; the server prunes old images by it.
LABEL org.opencontainers.image.source=https://github.com/martin-pouls/Board-Games
COPY --from=api /out/server /app/server
COPY --from=web /src/the-game/frontend/build /app/static
# Owned by nonroot so a fresh named volume mounted here is writable.
COPY --from=api --chown=65532:65532 /out/data /data
ENV ADDR=:8080 DB_PATH=/data/the-game.db STATIC_DIR=/app/static
EXPOSE 8080
VOLUME /data
ENTRYPOINT ["/app/server"]
