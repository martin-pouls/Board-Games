#!/usr/bin/env bash
# Decides which apps a run of .github/workflows/deploy.yml tests, builds and deploys, and writes
# them as a JSON array to $GITHUB_OUTPUT ("apps=[...]"). Run from the repo root. It also writes
# whether the run checks shared/backend and shared/frontend ("shared_backend=true|false", same for
# frontend), once for all apps, and whether it deploys the shared Caddy proxy ("caddy=true|false").
#
#   REQUESTED  set (manual run): "all" or space-separated app names, "caddy" among them for the
#              proxy; exactly those ("all" includes caddy).
#   BASE       otherwise (push): the commit the last successful push run deployed. The apps with
#              changes since then, and caddy when deploy/caddy/ changed: a cancelled or failed run
#              is caught up by the next one. Everything when BASE is empty or not an ancestor of
#              HEAD (first run, rewritten history).
#
# An app is a top-level directory with deploy/compose.yaml (the same rule as deploy/deploy.sh); it
# needs no Dockerfile of its own. An app depends on its own directory and on every top-level
# directory its Dockerfile copies from the build context (COPY without --from), where the
# Dockerfile is the one named by "dockerfile:" in its deploy/compose.yaml (relative to the build
# context, the repository root). So a change in shared/ counts for every app that copies it, and a
# change in the-crew/ also counts for deep-sea, which builds the-crew/Dockerfile.
# A shared part is checked when it changed since BASE, and on runs without BASE (manual, first)
# when one of the apps uses it.
set -euo pipefail

# The Dockerfile app $1 is built from.
dockerfile_of() {
	local f
	f=$(sed -n 's/^[[:space:]]*dockerfile:[[:space:]]*\([^[:space:]#]*\).*/\1/p' "$1/deploy/compose.yaml" | head -n1)
	f=${f//[\"\']/}
	echo "${f:-$1/Dockerfile}"
}

# The top-level directories the Dockerfile $1 copies from the build context, one per line.
copy_dirs() {
	awk '$1 == "COPY" && !/--from/ {
		n = 0
		for (i = 2; i <= NF; i++) if ($i !~ /^--/) src[++n] = $i
		for (i = 1; i < n; i++) { split(src[i], p, "/"); if (p[1] != "" && p[1] != ".") print p[1] }
	}' "$1" 2>/dev/null || true
}

all=()
declare -A deps # app -> space-separated top-level directories whose changes redeploy it
for d in */; do
	d=${d%/}
	[ -f "$d/deploy/compose.yaml" ] || continue
	all+=("$d")
	deps[$d]=" $d $(copy_dirs "$(dockerfile_of "$d")" | sort -u | tr '\n' ' ')"
done

# Whether app $1 is built with shared/$2.
uses() { grep -q "^COPY shared/$2" "$(dockerfile_of "$1")" 2>/dev/null; }

is_app() {
	local a
	for a in "${all[@]}"; do [ "$a" = "$1" ] && return 0; done
	return 1
}

apps=()
# Set on push runs with BASE; changed_shared then lists the files in shared/ changed since BASE.
compared=
changed_shared=
caddy=false
if [ -n "${REQUESTED:-}" ]; then
	if [ "$REQUESTED" = all ]; then
		apps=("${all[@]}")
		caddy=true
	else
		for a in $REQUESTED; do
			if [ "$a" = caddy ]; then
				caddy=true
				continue
			fi
			is_app "$a" || { echo "unknown app '$a'; apps: ${all[*]} caddy" >&2; exit 1; }
			apps+=("$a")
		done
	fi
elif [ -n "${BASE:-}" ] && git merge-base --is-ancestor "$BASE" HEAD 2>/dev/null; then
	echo "changes since $BASE:"
	while read -r d; do
		for a in "${all[@]}"; do
			if [[ "${deps[$a]}" == *" $d "* && " ${apps[*]-} " != *" $a "* ]]; then apps+=("$a"); fi
		done
	done < <(git diff --name-only "$BASE" HEAD | cut -d/ -f1 | sort -u)
	changed_shared=$(git diff --name-only "$BASE" HEAD -- shared/)
	[ -z "$(git diff --name-only "$BASE" HEAD -- deploy/caddy/)" ] || caddy=true
	compared=1
else
	echo "no previous deploy to compare with: every app and caddy"
	apps=("${all[@]}")
	caddy=true
fi

json=$(printf '"%s",' "${apps[@]}")
json="[${json%,}]"
[ ${#apps[@]} -gt 0 ] || json='[]'
echo "apps: $json"
echo "apps=$json" >>"${GITHUB_OUTPUT:-/dev/null}"
echo "caddy: $caddy"
echo "caddy=$caddy" >>"${GITHUB_OUTPUT:-/dev/null}"

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
