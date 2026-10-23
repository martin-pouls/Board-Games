#!/usr/bin/env bash
# Starts apps for local use in the background (backend + Vite dev server on their fixed ports) and
# waits until they answer. Servers that already answer are left alone (e.g. the user's own).
#
#   run.sh fox-duet [the-crew ...]   start these apps (deep-sea = the-crew with EDITION=deepSea)
#   run.sh stop fox-duet [...]       stop whatever listens on these apps' ports
#   run.sh stop --all                stop every app's ports
#   run.sh status                    which app ports answer
#
# Ports mirror the table in the root CLAUDE.md. Logs: $TEMP/board-games-run/<app>-{backend,frontend}.log.
# The gate is not supported (its backend needs secrets; see gate/CLAUDE.md).
set -uo pipefail

root=$(git rev-parse --show-toplevel) || exit 1
logdir="${TEMP:-${TMPDIR:-/tmp}}/board-games-run"
mkdir -p "$logdir"

# app -> "code-dir frontend-port backend-port vite-mode" (backend-port - = frontend only)
declare -A apps=(
	[the-game]="the-game 5173 8080 -"
	[regicide]="regicide 5174 8081 -"
	[start-page]="start-page 5175 - -"
	[fox-duet]="fox-duet 5176 8082 -"
	[the-crew]="the-crew 5177 8084 -"
	[deep-sea]="the-crew 5179 8085 deepSea"
	[sky-team]="sky-team 5180 8086 -"
)

answers() { curl -s -o /dev/null --max-time 2 "$1"; }

# PIDs listening on a TCP port (Windows netstat, else lsof).
pids_on() {
	if command -v netstat >/dev/null && [[ $(uname -s) == MINGW* || $(uname -s) == MSYS* ]]; then
		netstat -ano | awk -v p=":$1" '$1 == "TCP" && $4 == "LISTENING" && substr($2, length($2) - length(p) + 1) == p { print $5 }' | sort -u
	else
		lsof -ti "tcp:$1" -sTCP:LISTEN 2>/dev/null
	fi
}

kill_port() {
	local pid
	for pid in $(pids_on "$1"); do
		if command -v taskkill >/dev/null; then taskkill //F //T //PID "$pid" >/dev/null 2>&1; else kill "$pid"; fi
	done
}

wait_for() { # url name log
	local i
	for ((i = 0; i < 120; i++)); do
		answers "$1" && return 0
		sleep 1
	done
	echo "FAIL  $2 did not answer at $1 within 120 s. Last log lines ($3):" >&2
	tail -n 30 "$3" >&2
	return 1
}

start() { # app
	local dir fport bport mode
	read -r dir fport bport mode <<<"${apps[$1]}"
	local env=()
	[[ $mode == deepSea ]] && env=(EDITION=deepSea)

	if [[ $bport != - ]]; then
		local blog="$logdir/$1-backend.log" url="http://localhost:$bport/api/health"
		if answers "$url"; then
			echo "ok    $1 backend already running on :$bport"
		else
			local exe="$logdir/$1-server.exe"
			(cd "$root/$dir/backend" && go build -o "$exe" ./cmd/server) >"$blog" 2>&1 || {
				echo "FAIL  $1 backend build:" >&2; tail -n 30 "$blog" >&2; return 1
			}
			(cd "$root/$dir/backend" && env "${env[@]}" nohup "$exe" >>"$blog" 2>&1 &)
			wait_for "$url" "$1 backend" "$blog" || return 1
			echo "ok    $1 backend on :$bport"
		fi
	fi

	local flog="$logdir/$1-frontend.log" url="http://localhost:$fport/"
	if answers "$url"; then
		echo "ok    $1 frontend already running: $url"
	else
		[[ -d $root/$dir/frontend/node_modules ]] || { echo "FAIL  $1: run npm ci in $dir/frontend" >&2; return 1; }
		local args=(dev)
		[[ $mode != - ]] && args+=(--mode "$mode")
		(cd "$root/$dir/frontend" && nohup node node_modules/vite/bin/vite.js "${args[@]}" >"$flog" 2>&1 &)
		wait_for "$url" "$1 frontend" "$flog" || return 1
		echo "ok    $1 frontend: $url"
	fi
	[[ $bport != - ]] && echo "      fixtures: ${url}dev/fixtures"
	return 0
}

stop() { # app
	local dir fport bport mode
	read -r dir fport bport mode <<<"${apps[$1]}"
	kill_port "$fport"
	[[ $bport != - ]] && kill_port "$bport"
	echo "ok    $1 stopped"
}

check_app() {
	[[ -n ${apps[$1]+x} ]] && return 0
	echo "unknown app '$1' (known: ${!apps[*]})" >&2
	exit 2
}

cmd=start
case ${1:-} in
	stop|status) cmd=$1; shift ;;
	-h|--help|'') sed -n '2,11p' "$0"; exit 0 ;;
esac

if [[ $cmd == status ]]; then
	for a in the-game regicide start-page fox-duet the-crew deep-sea sky-team; do
		read -r dir fport bport mode <<<"${apps[$a]}"
		f=down; answers "http://localhost:$fport/" && f=up
		b=""; [[ $bport != - ]] && { b="  backend :$bport down"; answers "http://localhost:$bport/api/health" && b="  backend :$bport up"; }
		printf '%-11s frontend :%s %s%s\n' "$a" "$fport" "$f" "$b"
	done
	exit 0
fi

targets=("$@")
[[ $cmd == stop && ${targets[0]:-} == --all ]] && targets=(the-game regicide start-page fox-duet the-crew deep-sea sky-team)
for a in "${targets[@]}"; do check_app "$a"; done

status=0
for a in "${targets[@]}"; do "$cmd" "$a" || status=1; done
exit $status
