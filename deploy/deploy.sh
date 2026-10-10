#!/usr/bin/env bash
# Deploys apps of the committed HEAD to the server. For each app: upload its subtree, rebuild its
# image there (native arm64), restart its container and install its Caddy snippet; then reload
# Caddy once. Uncommitted changes are not deployed.
#
# Usage from Git Bash on your PC (first time: deploy/setup-server.sh):
#   deploy/deploy.sh                       # every app
#   deploy/deploy.sh regicide              # one app
#   deploy/deploy.sh the-game start-page   # several
#
# An app is a top-level directory with <app>/Dockerfile, <app>/deploy/compose.yaml and
# <app>/deploy/<app>.caddy. Its source lands in /opt/card-games/<app>/src on the server.
set -euo pipefail

HOST=${DEPLOY_HOST:-root@yoshi-und-kruemli.de}
KEY=${DEPLOY_KEY:-$HOME/.ssh/hetzner_ed25519}
SITE=https://yoshi-und-kruemli.de

cd "$(git rev-parse --show-toplevel)"
rev=$(git rev-parse --short HEAD)

all_apps() {
	git ls-tree -d --name-only HEAD | while read -r d; do
		git cat-file -e "HEAD:$d/deploy/compose.yaml" 2>/dev/null && echo "$d"
	done
	return 0
}

if [ $# -eq 0 ]; then
	mapfile -t apps < <(all_apps)
else
	apps=("$@")
fi

for app in "${apps[@]}"; do
	if ! git cat-file -e "HEAD:$app/deploy/compose.yaml" 2>/dev/null ||
		! git cat-file -e "HEAD:$app/deploy/$app.caddy" 2>/dev/null; then
		echo "unknown app '$app': not committed, or missing deploy/compose.yaml or deploy/$app.caddy" >&2
		echo "apps: $(all_apps | tr '\n' ' ')" >&2
		exit 1
	fi
done

ssh -i "$KEY" "$HOST" '[ -d /opt/caddy/conf/sites ] && docker network inspect web >/dev/null 2>&1' ||
	{ echo "shared Caddy missing: run deploy/setup-server.sh first" >&2; exit 1; }

for app in "${apps[@]}"; do
	echo "==> $app"
	[ -z "$(git status --porcelain -- "$app")" ] || echo "warning: uncommitted changes in $app/ are not deployed" >&2
	git -c core.autocrlf=false archive --format=tar "HEAD:$app" | ssh -i "$KEY" "$HOST" "set -eu
dir=/opt/card-games/$app
rm -rf \$dir/src.new
mkdir -p \$dir/src.new
tar -x -C \$dir/src.new
rm -rf \$dir/src
mv \$dir/src.new \$dir/src
echo $rev >\$dir/REVISION
cd \$dir/src
docker compose -f deploy/compose.yaml up -d --build
cp deploy/$app.caddy /opt/caddy/conf/sites/$app.caddy
"
done

ssh -i "$KEY" "$HOST" 'set -eu
docker exec caddy caddy reload --config /etc/caddy/Caddyfile
docker image prune -f >/dev/null
'
for app in "${apps[@]}"; do
	path=$(git show "HEAD:$app/deploy/$app.caddy" | sed -n 's#^handle_path \(/[^/]*\)/\*.*#\1#p')
	echo "deployed $app $rev to $SITE$path/"
done
