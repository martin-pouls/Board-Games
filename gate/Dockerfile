# syntax=docker/dockerfile:1
# Single static Go binary: the access gate (forward_auth target + login page). Built for arm64 by CI
# (.github/workflows/deploy.yml) from deploy/compose.yaml, or on the server by deploy/deploy.sh.
#
# The build context is the repository root (the app uses shared/), so paths below are from the
# root. The image keeps the repo layout (/src/gate, /src/shared): go.mod finds shared/ at
# ../../shared as it does in a checkout. Dockerfile.dockerignore next to this file applies.

FROM golang:1.27-alpine AS api
WORKDIR /src/gate/backend
# go.mod replaces the shared module with ../../shared/backend.
COPY shared/backend/ /src/shared/backend/
COPY gate/backend/go.mod ./
RUN go mod download
COPY gate/backend/ ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/server ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
# Links the image to the repo on ghcr.io; the server prunes old images by it.
LABEL org.opencontainers.image.source=https://github.com/martin-pouls/Board-Games
COPY --from=api /out/server /app/server
# The secret and the passwords are mounted read-only at /secrets (see deploy/compose.yaml).
ENV ADDR=:8080 GATE_SECRET_FILE=/secrets/secret GATE_PASSWORDS_FILE=/secrets/passwords
EXPOSE 8080
ENTRYPOINT ["/app/server"]
