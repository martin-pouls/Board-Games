# syntax=docker/dockerfile:1
# Single image: Go server + built SPA. Built natively on the (arm64) server by deploy/compose.yaml.

FROM node:24-alpine AS web
WORKDIR /src/frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci --no-audit --no-fund
COPY frontend/ ./
# Sub-path the SPA is served under (see vite.config.ts); empty = site root.
ARG BASE_PATH=""
RUN BASE_PATH="$BASE_PATH" npm run build

FROM golang:1.27-alpine AS api
WORKDIR /src/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server \
	&& mkdir -p /out/data

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=api /out/server /app/server
COPY --from=web /src/frontend/build /app/static
# Owned by nonroot so a fresh named volume mounted here is writable.
COPY --from=api --chown=65532:65532 /out/data /data
ENV ADDR=:8080 DB_PATH=/data/regicide.db STATIC_DIR=/app/static
EXPOSE 8080
VOLUME /data
ENTRYPOINT ["/app/server"]
