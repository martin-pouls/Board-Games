#!/usr/bin/env bash
# Runs the "before declaring work done" checks for the apps affected by the current changes:
# backend `go vet ./... && go test ./...`, frontend `npm run check && npm run lint && npm test`.
#
#   check.sh                  apps touched by uncommitted changes (staged, unstaged, untracked)
#   check.sh --base <ref>     apps touched since <ref> (e.g. origin/main), plus uncommitted changes
#   check.sh --all            every app
#   check.sh fox-duet shared  these apps (shared expands to every app, as below)
#
# Apps are discovered, not listed: any top-level directory with backend/go.mod or
# frontend/package.json. A change in shared/ checks every app. Directories without code
# (deep-sea/, deploy/, the repo root) need no checks: Deep Sea is the-crew's code.
# Checks of different apps run in parallel; each part's log is kept for failures.
set -uo pipefail

root=$(git rev-parse --show-toplevel) || exit 1
cd "$root" || exit 1

all_apps=()
for d in */; do
	d=${d%/}
	[[ -f $d/backend/go.mod || -f $d/frontend/package.json ]] && all_apps+=("$d")
done

base=""
explicit=()
everything=false
while (($#)); do
	case $1 in
		--all) everything=true ;;
		--base) base=${2:?--base needs a ref}; shift ;;
		-h|--help) sed -n '2,13p' "$0"; exit 0 ;;
		*) explicit+=("${1%/}") ;;
	esac
	shift
done

declare -A selected=()
if $everything; then
	for a in "${all_apps[@]}"; do selected[$a]=1; done
else
	if ((${#explicit[@]})); then
		dirs=("${explicit[@]}")
	else
		mapfile -t dirs < <(
			{
				git diff --name-only HEAD
				git ls-files --others --exclude-standard
				[[ -n $base ]] && git diff --name-only "$base"...HEAD
			} | grep / | grep -v '\.md$' | cut -d/ -f1 | sort -u
		)
	fi
	for d in "${dirs[@]}"; do
		if [[ $d == shared ]]; then
			for a in "${all_apps[@]}"; do selected[$a]=1; done
		elif [[ " ${all_apps[*]} " == *" $d "* ]]; then
			selected[$d]=1
		elif [[ $d == deep-sea ]] && ((${#explicit[@]})); then
			selected[the-crew]=1
		elif ((${#explicit[@]})); then
			echo "unknown app: $d (apps: ${all_apps[*]})" >&2
			exit 2
		fi
	done
fi

if ((${#selected[@]} == 0)); then
	echo "No app code changed: nothing to check."
	exit 0
fi

logs=$(mktemp -d)
pids=()
parts=()
for app in $(printf '%s\n' "${!selected[@]}" | sort); do
	if [[ -f $app/backend/go.mod ]]; then
		(cd "$app/backend" && go vet ./... && go test ./...) >"$logs/$app-backend.log" 2>&1 &
		pids+=($!); parts+=("$app/backend")
	fi
	if [[ -f $app/frontend/package.json ]]; then
		if [[ ! -d $app/frontend/node_modules ]]; then
			echo "$app/frontend: node_modules missing, run npm ci there first" >"$logs/$app-frontend.log"
			(exit 1) &
		else
			(cd "$app/frontend" && npm run check && npm run lint && npm test) >"$logs/$app-frontend.log" 2>&1 &
		fi
		pids+=($!); parts+=("$app/frontend")
	fi
done

echo "Checking: ${parts[*]}"
failed=()
for i in "${!pids[@]}"; do
	wait "${pids[$i]}" || failed+=("${parts[$i]}")
done

for p in "${failed[@]}"; do
	log="$logs/${p%/*}-${p#*/}.log"
	echo
	echo "=== FAIL $p (full log: $log) ==="
	tail -n 60 "$log"
done

echo
for p in "${parts[@]}"; do
	if [[ " ${failed[*]} " == *" $p "* ]]; then echo "FAIL  $p"; else echo "ok    $p"; fi
done
((${#failed[@]} == 0))
