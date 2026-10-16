# syntax=docker/dockerfile:1
# Single static Go binary: the access gate (forward_auth target + login page). Built for arm64 by CI
# (.github/workflows/deploy.yml) from deploy/compose.yaml, or on the server by deploy/deploy.sh.
#
# The build context is the repository root (the app uses shared/), so paths below are from the
# root. The image keeps the repo layout (/src/gate, /src/shared): go.mod and the frontend find
# shared/ at ../../shared as they do in a checkout. Dockerfile.dockerignore next to this file applies.

FROM node:25-alpine@sha256:bdf2cca6fe3dabd014ea60163eca3f0f7015fbd5c7ee1b0e9ccb4ced6eb02ef4 AS web
WORKDIR /src/gate/frontend
COPY gate/frontend/package.json gate/frontend/package-lock.json ./
RUN npm ci --no-audit --no-fund
# Compiled as source with the app; its imports resolve from the app's node_modules.
COPY shared/frontend/ /src/shared/frontend/
COPY gate/frontend/ ./
# Sub-path the login page and its assets are served under (see vite.config.ts).
ARG BASE_PATH=""
# Writes the page into ../backend/internal/gate/ui/build, where the Go build embeds it.
RUN BASE_PATH="$BASE_PATH" npm run build

FROM golang:1.27-alpine@sha256:f92b6ef800e499660581efdabdf25d9d817a9d124eaf900924f0504e7e27e12d AS api
WORKDIR /src/gate/backend
# go.mod replaces the shared module with ../../shared/backend.
COPY shared/backend/ /src/shared/backend/
COPY gate/backend/go.mod ./
RUN go mod download
COPY gate/backend/ ./
COPY --from=web /src/gate/backend/internal/gate/ui/build ./internal/gate/ui/build
# The tests check the built login page too (TestBuiltLoginPage, skipped without a build).
RUN go test ./... \
	&& CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM gcr.io/distroless/static-debian13:nonroot@sha256:e2e927ec666bae08560abb3c55d0659eceabb657f56b6782ab500a9fc7f555e3
# Links the image to the repo on ghcr.io; the server prunes old images by it.
LABEL org.opencontainers.image.source=https://github.com/martin-pouls/Board-Games
COPY --from=api /out/server /app/server
# The secret and the passwords are mounted read-only at /secrets (see deploy/compose.yaml).
ENV ADDR=:8080 GATE_SECRET_FILE=/secrets/secret GATE_PASSWORDS_FILE=/secrets/passwords
EXPOSE 8080
ENTRYPOINT ["/app/server"]
