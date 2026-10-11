#!/usr/bin/env bash
# Local prod-like stack: every app's real image (built from its deploy/compose.yaml, like CI) behind
# a real Caddy with the server's site config (deploy/caddy/conf/common.caddy, the gate and all
# snippets), at http://localhost:8000. Runs in WSL2 with Docker (deploy/local/setup-wsl.sh).
#
#   deploy/local.sh                 # build and (re)start every app, then Caddy
#   deploy/local.sh regicide gate   # only these apps (the others keep running)
#   deploy/local.sh down            # stop and remove everything (game databases stay in volumes)
#
# Gate passwords: deploy/local/.gate/passwords ("label=password" lines), made on the first run
# with "local=local". Edit it and run "deploy/local.sh gate" to change them. Images are tagged
# :local and never pushed.
set -euo pipefail

cd "$(git rev-parse --show-toplevel)"
here=deploy/local
command -v docker >/dev/null || { echo "docker missing: run deploy/local/setup-wsl.sh" >&2; exit 1; }
docker info >/dev/null 2>&1 || { echo "docker is not running or you are not in the docker group (log out and in again after setup-wsl.sh)" >&2; exit 1; }

all_apps() { for f in */deploy/compose.yaml; do echo "${f%%/*}"; done; }
compose() { # app, compose args...
	local app=$1; shift
	local files=(-f "$app/deploy/compose.yaml")
	[ "$app" = gate ] && files+=(-f "$here/gate.override.yaml")
	TAG=local LOCAL_GATE_DIR="$PWD/$here/.gate" docker compose "${files[@]}" "$@"
}

if [ "${1:-}" = down ]; then
	for app in $(all_apps); do compose "$app" down; done
	docker compose -f "$here/compose.yaml" down
	rm -rf "$here/.sites"
	exit 0
fi

apps=("$@")
[ ${#apps[@]} -gt 0 ] || mapfile -t apps < <(all_apps)
for app in "${apps[@]}"; do
	[ -f "$app/deploy/compose.yaml" ] && [ -f "$app/deploy/$app.caddy" ] ||
		{ echo "unknown app '$app': missing deploy/compose.yaml or deploy/$app.caddy" >&2; exit 1; }
done

docker network inspect web >/dev/null 2>&1 || docker network create web >/dev/null

# The gate's secret and passwords. World-readable: the container runs as uid 65532, and these
# are local test values only.
mkdir -p "$here/.gate" "$here/.sites"
[ -s "$here/.gate/secret" ] || openssl rand -hex 32 >"$here/.gate/secret"
[ -s "$here/.gate/passwords" ] || printf '# label=password, one per line\nlocal=local\n' >"$here/.gate/passwords"
chmod 755 "$here/.gate"
chmod 644 "$here/.gate/secret" "$here/.gate/passwords"

for app in "${apps[@]}"; do
	echo "==> $app: building"
	compose "$app" build
	compose "$app" up -d --force-recreate
	cp "$app/deploy/$app.caddy" "$here/.sites/$app.caddy"
done

docker compose -f "$here/compose.yaml" up -d
docker exec caddy-local caddy reload --config /etc/caddy/Caddyfile
echo "==> up: http://localhost:8000/board-games/ (gate passwords: $here/.gate/passwords)"
