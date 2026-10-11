#!/usr/bin/env bash
# Idempotent server bootstrap: Docker, firewall, automatic security updates, SSH hardening, the
# shared Caddy proxy in /opt/caddy and the "deploy" user that GitHub Actions deploys as. Run from
# Git Bash on your PC (as root over your own key). Setup of the CI deploy: deploy/README.md.
#
#   deploy/setup-server.sh                                    # keeps the deploy user's key as is
#   CI_PUBKEY=~/.ssh/board-games-ci.pub deploy/setup-server.sh # installs (or rotates) the CI key
#   GATE_PASSWORDS=~/gate-passwords deploy/setup-server.sh     # installs (or changes) the passwords
#
# The gate (gate/, the shared password in front of the games) reads /etc/board-games/gate/: a
# secret this script generates once, and the passwords file, one "label=password" per line (e.g.
# "family=…"; a cookie stays valid as long as its label's password and the secret don't change).
# The passwords travel inside the tar stream, never on a command line. Without GATE_PASSWORDS an
# installed file is kept; on a first run the script fails without it.
#
# Re-run it after changing deploy/caddy/ or deploy/server/ (it ships the committed HEAD): the repo
# owns the Caddyfile and common.caddy; a server copy that differs is kept as .bak before it is replaced.
set -euo pipefail

HOST=${DEPLOY_HOST:-root@yoshi-und-kruemli.de}
KEY=${DEPLOY_KEY:-$HOME/.ssh/hetzner_ed25519}

cd "$(git rev-parse --show-toplevel)"
pubkey=
if [ -n "${CI_PUBKEY:-}" ]; then
	grep -qE '^ssh-(ed25519|rsa) ' "$CI_PUBKEY" || { echo "$CI_PUBKEY is not an SSH public key" >&2; exit 1; }
	pubkey=$(head -n1 "$CI_PUBKEY" | base64 -w0) # shell-safe on the remote command line
fi
if [ -n "${GATE_PASSWORDS:-}" ]; then
	[ -f "$GATE_PASSWORDS" ] || { echo "$GATE_PASSWORDS: no such file" >&2; exit 1; }
	# The gate checks the format too and refuses to start; this catches mistakes before the upload.
	bad=$(tr -d '' <"$GATE_PASSWORDS" | grep -nvE '^[[:space:]]*(#|$)|^[a-z0-9][a-z0-9-]{0,31}[[:space:]]*=[[:space:]]*[^[:space:]]' | cut -d: -f1 | tr '
' ' ') || true
	[ -z "$bad" ] || { echo "$GATE_PASSWORDS: lines $bad are not \"label=password\" (label: a-z, 0-9, -)" >&2; exit 1; }
	tr -d '' <"$GATE_PASSWORDS" | grep -qvE '^[[:space:]]*(#|$)' || { echo "$GATE_PASSWORDS: no passwords" >&2; exit 1; }
fi

remote=$(cat <<'EOF'
set -eu
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT # it may hold the gate passwords
tar -x -i -C "$tmp" # stdin first: nothing below may consume it (-i: the passwords are a second archive)
export DEBIAN_FRONTEND=noninteractive
apt-get update -qq
apt-get install -y -qq docker.io docker-compose-v2 docker-buildx git ufw unattended-upgrades cron >/dev/null

# Logs are kept for 24 hours at most (the privacy policy says so). Caddy writes no access log and
# the apps log no IP addresses, but error output may contain one. The containers log to journald,
# which keeps entries 22 hours; an hourly job closes the current journal file and deletes the old
# ones, so nothing is older than 24 hours. Nothing is forwarded to syslog (/var/log/syslog would keep
# it for weeks), and ufw logs no blocked packets.
mkdir -p /etc/systemd/journald.conf.d
cat >/etc/systemd/journald.conf.d/10-board-games.conf <<'JOURNALD'
# Installed by deploy/setup-server.sh (Board-Games repo): logs for 24 hours at most.
[Journal]
MaxRetentionSec=22h
ForwardToSyslog=no
JOURNALD
systemctl restart systemd-journald
cat >/etc/cron.hourly/board-games-logs <<'CRON'
#!/bin/sh
# Installed by deploy/setup-server.sh (Board-Games repo): no log entry older than 24 hours.
journalctl --quiet --rotate --vacuum-time=22h
CRON
chmod 755 /etc/cron.hourly/board-games-logs

# Existing containers keep the log driver they were created with: recreate them after a change.
recreate=
if [ "$(cat /etc/docker/daemon.json 2>/dev/null)" != '{ "log-driver": "journald" }' ]; then
	echo '{ "log-driver": "journald" }' >/etc/docker/daemon.json
	systemctl restart docker
	recreate=1
fi
systemctl enable --now docker >/dev/null

ufw allow OpenSSH >/dev/null
ufw allow 80/tcp >/dev/null
ufw allow 443 >/dev/null # TCP + UDP (HTTP/3)
ufw logging off >/dev/null
ufw --force enable >/dev/null

mkdir -p /opt/caddy/conf/sites
cp "$tmp/caddy/compose.yaml" /opt/caddy/compose.yaml
for f in Caddyfile common.caddy; do
	if [ -f "/opt/caddy/conf/$f" ] && ! cmp -s "$tmp/caddy/conf/$f" "/opt/caddy/conf/$f"; then
		cp "/opt/caddy/conf/$f" "/opt/caddy/conf/$f.bak"
		echo "replaced a differing $f; the previous one is /opt/caddy/conf/$f.bak"
	fi
	cp "$tmp/caddy/conf/$f" "/opt/caddy/conf/$f"
done
docker compose -f /opt/caddy/compose.yaml up -d ${recreate:+--force-recreate}
docker exec caddy caddy reload --config /etc/caddy/Caddyfile
if [ -n "$recreate" ]; then
	for f in /opt/board-games/*/deploy/compose.yaml; do
		if [ -f "$f" ]; then docker compose -f "$f" up -d --no-build --force-recreate; fi
	done
	echo "recreated the containers: they log to journald now"
fi

# The CI deploy user. Being in the docker group makes it root-equivalent, so its key is pinned to
# the deploy script (forced command, "restrict": no shell, no forwarding, no pty), and that script
# deploys only commits of main. The password is "*": no password login, but unlike a locked
# account sshd still accepts its key.
id deploy >/dev/null 2>&1 || useradd --create-home --shell /bin/bash deploy
usermod -p '*' -aG docker deploy
# The repo used to be Card-Games: move its deploy root (compose project names stay, so the running
# containers and their volumes carry over) and drop the old script.
if [ -d /opt/card-games ] && [ ! -e /opt/board-games ]; then
	mv /opt/card-games /opt/board-games
	echo "moved /opt/card-games to /opt/board-games"
fi
rm -f /usr/local/bin/card-games-deploy
install -m 755 "$tmp/server/board-games-deploy" /usr/local/bin/board-games-deploy
install -d -o deploy -g deploy /opt/board-games
chown -R deploy:deploy /opt/board-games /opt/caddy/conf/sites

# The deploy key lives in a root-owned file, not in ~deploy/.ssh, so the account cannot change its
# own keys. Password logins are off for everyone (root already allowed keys only).
install -d -m 755 /etc/ssh/authorized_keys
keyfile=/etc/ssh/authorized_keys/deploy
if [ -n "$CI_PUBKEY_B64" ]; then
	ci_key=$(echo "$CI_PUBKEY_B64" | base64 -d)
	echo "installed the CI deploy key for user deploy"
elif [ -f "$keyfile" ]; then
	# Keep the key, but (re)pin it to the current script: the forced command may name an old one.
	ci_key=$(grep -oE 'ssh-(ed25519|rsa) .*' "$keyfile" | head -n1)
else
	ci_key=
fi
if [ -n "$ci_key" ]; then
	printf 'restrict,command="/usr/local/bin/board-games-deploy" %s\n' "$ci_key" >"$keyfile"
	chmod 644 "$keyfile"
fi
dropin=/etc/ssh/sshd_config.d/10-board-games.conf
cat >"$dropin" <<'SSHD'
# Installed by deploy/setup-server.sh (Board-Games repo).
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
rm -f /etc/ssh/sshd_config.d/10-card-games.conf # its name before the rename, same content
systemctl try-reload-or-restart ssh

# The gate's secret and passwords, readable only by the gate container's user (distroless nonroot)
# and root. A new secret would log everyone out, so it is made once and kept.
gate=/etc/board-games/gate
# Owners via chown: the server's install (uutils) rejects numeric ids that have no user.
install -d -m 700 "$gate"
chown 65532:65532 "$gate"
[ -s "$gate/secret" ] || { openssl rand -hex 32 >"$gate/secret"; echo "generated the gate secret"; }
if [ -f "$tmp/gate-passwords" ]; then
	install -m 600 "$tmp/gate-passwords" "$gate/passwords"
	chown 65532:65532 "$gate/passwords"
	echo "installed the gate passwords"
	if docker container inspect gate >/dev/null 2>&1; then docker restart gate >/dev/null; fi
elif [ ! -s "$gate/passwords" ]; then
	echo "no gate passwords on the server yet: run again with GATE_PASSWORDS=<file>" >&2
	exit 1
fi
chown 65532:65532 "$gate/secret"
chmod 600 "$gate/secret"
EOF
)

{
	git -c core.autocrlf=false archive --format=tar HEAD:deploy
	if [ -n "${GATE_PASSWORDS:-}" ]; then
		tar -c -C "$(dirname "$GATE_PASSWORDS")" --transform 's|.*|gate-passwords|' "$(basename "$GATE_PASSWORDS")"
	fi
} |
	ssh -i "$KEY" "$HOST" "CI_PUBKEY_B64='$pubkey'
$remote"
echo "server ready; deploys run from GitHub Actions (or deploy/deploy.sh as a fallback)"
