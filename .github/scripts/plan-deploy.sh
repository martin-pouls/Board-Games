#!/usr/bin/env bash
# Decides which apps a run of .github/workflows/deploy.yml tests, builds and deploys, and writes
# them as a JSON array to $GITHUB_OUTPUT ("apps=[...]"). Run from the repo root.
#
#   REQUESTED  set (manual run): "all" or space-separated app names; exactly those.
#   BASE       otherwise (push): the commit the last successful push run deployed. The apps with
#              changes since then: a cancelled or failed run is caught up by the next one. Every
#              app when BASE is empty or not an ancestor of HEAD (first run, rewritten history).
#
# An app is a top-level directory with deploy/compose.yaml (the same rule as deploy/deploy.sh).
set -euo pipefail

all=()
for d in */; do
	d=${d%/}
	[ -f "$d/deploy/compose.yaml" ] && all+=("$d")
done

is_app() {
	local a
	for a in "${all[@]}"; do [ "$a" = "$1" ] && return 0; done
	return 1
}

apps=()
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
		is_app "$d" && apps+=("$d")
	done < <(git diff --name-only "$BASE" HEAD | cut -d/ -f1 | sort -u)
else
	echo "no previous deploy to compare with: every app"
	apps=("${all[@]}")
fi

json=$(printf '"%s",' "${apps[@]}")
json="[${json%,}]"
[ ${#apps[@]} -gt 0 ] || json='[]'
echo "apps: $json"
echo "apps=$json" >>"${GITHUB_OUTPUT:-/dev/null}"
