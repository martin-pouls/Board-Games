#!/usr/bin/env bash
# One-time, idempotent server bootstrap: Docker, firewall, automatic security updates and the
# shared Caddy proxy in /opt/caddy. Run from Git Bash on your PC, then deploy with deploy/deploy.sh.
# An existing /opt/caddy/conf/Caddyfile is kept, so server-side edits survive a re-run.
set -euo pipefail

HOST=${DEPLOY_HOST:-root@yoshi-und-kruemli.de}
KEY=${DEPLOY_KEY:-$HOME/.ssh/hetzner_ed25519}

cd "$(git rev-parse --show-toplevel)"
git -c core.autocrlf=false archive --format=tar HEAD:deploy/caddy | ssh -i "$KEY" "$HOST" 'set -eu
tmp=$(mktemp -d)
tar -x -C "$tmp" # stdin first: nothing below may consume it
export DEBIAN_FRONTEND=noninteractive
apt-get update -qq
apt-get install -y -qq docker.io docker-compose-v2 docker-buildx ufw unattended-upgrades >/dev/null

# Rotating container logs (the default json-file driver grows without limit).
if [ ! -f /etc/docker/daemon.json ]; then
	echo "{ \"log-driver\": \"local\" }" >/etc/docker/daemon.json
	systemctl restart docker
fi
systemctl enable --now docker >/dev/null

ufw allow OpenSSH >/dev/null
ufw allow 80/tcp >/dev/null
ufw allow 443 >/dev/null # TCP + UDP (HTTP/3)
ufw --force enable >/dev/null

mkdir -p /opt/caddy/conf/sites
cp "$tmp/compose.yaml" /opt/caddy/compose.yaml
[ -f /opt/caddy/conf/Caddyfile ] || cp "$tmp/conf/Caddyfile" /opt/caddy/conf/Caddyfile
rm -rf "$tmp"
docker compose -f /opt/caddy/compose.yaml up -d
docker exec caddy caddy reload --config /etc/caddy/Caddyfile
'
echo "server ready; now run deploy/deploy.sh"
