#!/usr/bin/env bash
# Deploys the committed HEAD to the server: upload the source, rebuild the image there (native
# arm64), restart the container and reload Caddy. Uncommitted changes are not deployed.
# Usage from Git Bash on your PC: deploy/deploy.sh   (first time: deploy/setup-server.sh)
set -euo pipefail

HOST=${DEPLOY_HOST:-root@yoshi-und-kruemli.de}
KEY=${DEPLOY_KEY:-$HOME/.ssh/hetzner_ed25519}

cd "$(git rev-parse --show-toplevel)"
[ -z "$(git status --porcelain)" ] || echo "warning: uncommitted changes are not deployed" >&2
rev=$(git rev-parse --short HEAD)

git -c core.autocrlf=false archive --format=tar HEAD | ssh -i "$KEY" "$HOST" "set -eu
rm -rf /opt/the-game/src.new
mkdir -p /opt/the-game/src.new
tar -x -C /opt/the-game/src.new
rm -rf /opt/the-game/src
mv /opt/the-game/src.new /opt/the-game/src
echo $rev >/opt/the-game/REVISION
cd /opt/the-game/src
docker compose -f deploy/compose.yaml up -d --build
cp deploy/the-game.caddy /opt/caddy/conf/sites/the-game.caddy
docker exec caddy caddy reload --config /etc/caddy/Caddyfile
docker image prune -f >/dev/null
"
echo "deployed $rev to https://yoshi-und-kruemli.de/the-game/"
