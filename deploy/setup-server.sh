#!/usr/bin/env bash
# Idempotent server bootstrap: Docker, firewall, automatic security updates, SSH hardening, the
# shared Caddy proxy in /opt/caddy and the "deploy" user that GitHub Actions deploys as. Run from
# Git Bash on your PC (as root over your own key). Setup of the CI deploy: deploy/README.md.
#
#   deploy/setup-server.sh                                    # keeps the deploy user's key as is
#   CI_PUBKEY=~/.ssh/card-games-ci.pub deploy/setup-server.sh # installs (or rotates) the CI key
#
# Re-run it after changing deploy/caddy/ or deploy/server/ (it ships the committed HEAD): the repo
# owns the Caddyfile, and a server copy that differs is kept as Caddyfile.bak before it is replaced.
set -euo pipefail

HOST=${DEPLOY_HOST:-root@yoshi-und-kruemli.de}
KEY=${DEPLOY_KEY:-$HOME/.ssh/hetzner_ed25519}

cd "$(git rev-parse --show-toplevel)"
pubkey=
if [ -n "${CI_PUBKEY:-}" ]; then
	grep -qE '^ssh-(ed25519|rsa) ' "$CI_PUBKEY" || { echo "$CI_PUBKEY is not an SSH public key" >&2; exit 1; }
	pubkey=$(head -n1 "$CI_PUBKEY" | base64 -w0) # shell-safe on the remote command line
fi

remote=$(cat <<'EOF'
set -eu
tmp=$(mktemp -d)
tar -x -C "$tmp" # stdin first: nothing below may consume it
export DEBIAN_FRONTEND=noninteractive
apt-get update -qq
apt-get install -y -qq docker.io docker-compose-v2 docker-buildx git ufw unattended-upgrades >/dev/null

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
cp "$tmp/caddy/compose.yaml" /opt/caddy/compose.yaml
if [ -f /opt/caddy/conf/Caddyfile ] && ! cmp -s "$tmp/caddy/conf/Caddyfile" /opt/caddy/conf/Caddyfile; then
	cp /opt/caddy/conf/Caddyfile /opt/caddy/conf/Caddyfile.bak
	echo "replaced a differing Caddyfile; the previous one is /opt/caddy/conf/Caddyfile.bak"
fi
cp "$tmp/caddy/conf/Caddyfile" /opt/caddy/conf/Caddyfile
docker compose -f /opt/caddy/compose.yaml up -d
docker exec caddy caddy reload --config /etc/caddy/Caddyfile

# The CI deploy user. Being in the docker group makes it root-equivalent, so its key is pinned to
# the deploy script (forced command, "restrict": no shell, no forwarding, no pty), and that script
# deploys only commits of main. The password is "*": no password login, but unlike a locked
# account sshd still accepts its key.
id deploy >/dev/null 2>&1 || useradd --create-home --shell /bin/bash deploy
usermod -p '*' -aG docker deploy
install -m 755 "$tmp/server/card-games-deploy" /usr/local/bin/card-games-deploy
install -d -o deploy -g deploy /opt/card-games
chown -R deploy:deploy /opt/card-games /opt/caddy/conf/sites

# The deploy key lives in a root-owned file, not in ~deploy/.ssh, so the account cannot change its
# own keys. Password logins are off for everyone (root already allowed keys only).
install -d -m 755 /etc/ssh/authorized_keys
if [ -n "$CI_PUBKEY_B64" ]; then
	printf 'restrict,command="/usr/local/bin/card-games-deploy" %s\n' "$(echo "$CI_PUBKEY_B64" | base64 -d)" \
		>/etc/ssh/authorized_keys/deploy
	chmod 644 /etc/ssh/authorized_keys/deploy
	echo "installed the CI deploy key for user deploy"
fi
dropin=/etc/ssh/sshd_config.d/10-card-games.conf
cat >"$dropin" <<'SSHD'
# Installed by deploy/setup-server.sh (Card-Games repo).
PasswordAuthentication no
KbdInteractiveAuthentication no
Match User deploy
	AuthorizedKeysFile /etc/ssh/authorized_keys/deploy
SSHD
mkdir -p /run/sshd # sshd -t needs it while ssh.socket has not started the service yet
if ! sshd -t; then
	rm -f "$dropin"
	echo "sshd rejected $dropin; removed it again" >&2
	exit 1
fi
systemctl try-reload-or-restart ssh
rm -rf "$tmp"
EOF
)

git -c core.autocrlf=false archive --format=tar HEAD:deploy |
	ssh -i "$KEY" "$HOST" "CI_PUBKEY_B64='$pubkey'
$remote"
echo "server ready; deploys run from GitHub Actions (or deploy/deploy.sh as a fallback)"
