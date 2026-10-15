#!/usr/bin/env bash
# Decides which apps a run of .github/workflows/deploy.yml tests, builds and deploys, and writes
# them as a JSON array to $GITHUB_OUTPUT ("apps=[...]"). Run from the repo root. It also writes
# whether the run checks shared/backend and shared/frontend ("shared_backend=true|false", same for
# frontend), once for all apps.
#
#   REQUESTED  set (manual run): "all" or space-separated app names; exactly those.
#   BASE       otherwise (push): the commit the last successful push run deployed. The apps with
#              changes since then: a cancelled or failed run is caught up by the next one. Every
#              app when BASE is empty or not an ancestor of HEAD (first run, rewritten history).
#
# An app is a top-level directory with deploy/compose.yaml (the same rule as deploy/deploy.sh).
# An app uses shared/ when its Dockerfile copies it; a change in shared/ counts for each of them.
# A shared part is checked when it changed since BASE, and on runs without BASE (manual, first)
# when one of the apps uses it.
set -euo pipefail

all=()
users=() # apps built with shared/
for d in */; do
	d=${d%/}
	[ -f "$d/deploy/compose.yaml" ] || continue
	all+=("$d")
	grep -q '^COPY shared/' "$d/Dockerfile" 2>/dev/null && users+=("$d")
done

# Whether app $1 is built with shared/$2.
uses() { grep -q "^COPY shared/$2" "$1/Dockerfile" 2>/dev/null; }

is_app() {
	local a
	for a in "${all[@]}"; do [ "$a" = "$1" ] && return 0; done
	return 1
}

apps=()
# Set on push runs with BASE; changed_shared then lists the files in shared/ changed since BASE.
compared=
changed_shared=
if [ -n "${REQUESTED:-}" ]; then
	if [ "$REQUESTED" = all ]; then
		apps=("${all[@]}")
	else
		for a in $REQUESTED; do
			is_app "$a" || { echo "unknown app '$a'; apps: ${all[*]}" >&2; exit 1; }
			apps+=("$a")
		done
	fi
elif [ -n "${BASE:-}" ] && git merge-base --is-ancestor "$BASE" HEAD 2>/dev/null; then
	echo "changes since $BASE:"
	while read -r d; do
		if [ "$d" = shared ]; then
			for u in "${users[@]}"; do
				[[ " ${apps[*]} " == *" $u "* ]] || apps+=("$u")
			done
		elif is_app "$d" && [[ " ${apps[*]} " != *" $d "* ]]; then
			apps+=("$d")
		fi
	done < <(git diff --name-only "$BASE" HEAD | cut -d/ -f1 | sort -u)
	changed_shared=$(git diff --name-only "$BASE" HEAD -- shared/)
	compared=1
else
	echo "no previous deploy to compare with: every app"
	apps=("${all[@]}")
fi

json=$(printf '"%s",' "${apps[@]}")
json="[${json%,}]"
[ ${#apps[@]} -gt 0 ] || json='[]'
echo "apps: $json"
echo "apps=$json" >>"${GITHUB_OUTPUT:-/dev/null}"

for part in backend frontend; do
	check=false
	if [ -n "${compared:-}" ]; then
		grep -q "^shared/$part/" <<<"$changed_shared" && check=true
	else
		for a in "${apps[@]}"; do uses "$a" "$part" && check=true; done
	fi
	echo "shared/$part: $check"
	echo "shared_$part=$check" >>"${GITHUB_OUTPUT:-/dev/null}"
done
