#!/usr/bin/env bash
# Copies Fox Duet's committed files into a new top-level game directory and renames the app
# identity (directory, Go module, title, ports, DB, container, base path, Caddy snippet).
# Game logic is NOT touched: the scaffold work package strips it to a lobby-only skeleton.
#
#   scaffold.sh <dir> "<Title>" <vite-port> <backend-port>
#   scaffold.sh sail "Sail" 5180 8086
#
# Skips fox-duet/CLAUDE.md and docs/rules.md (the new game writes its own). Prints the files that
# still mention Fox Duet's game (fox, forest, decree, gem, ...): the scaffold agent's work list.
set -euo pipefail

root=$(git rev-parse --show-toplevel)
cd "$root"

[[ $# -eq 4 ]] || { sed -n '2,10p' "$0"; exit 2; }
dir=$1 title=$2 vport=$3 bport=$4

[[ $dir =~ ^[a-z][a-z0-9-]*$ ]] || { echo "dir must be lower-case kebab: $dir" >&2; exit 2; }
[[ -e $dir ]] && { echo "$dir already exists" >&2; exit 2; }
[[ $vport =~ ^[0-9]+$ && $bport =~ ^[0-9]+$ ]] || { echo "ports must be numbers" >&2; exit 2; }
for p in "$vport" "$bport"; do
	if grep -qE "[\" ]$p[\" ]" .claude/skills/run/run.sh; then
		echo "port $p is taken (see .claude/skills/run/run.sh)" >&2
		exit 2
	fi
done

src=fox-duet
mapfile -t files < <(git ls-files "$src" | grep -vE "^$src/(CLAUDE\.md|docs/rules\.md)$")
for f in "${files[@]}"; do
	dst="$dir/${f#"$src"/}"
	dst=${dst//fox-duet.caddy/$dir.caddy}
	mkdir -p "$(dirname "$dst")"
	cp "$f" "$dst"
done

# Identity renames. Order matters: the module path before the bare directory name.
esc_title=$(printf '%s' "$title" | sed 's/[&/\]/\\&/g')
grep -rlIE 'fox-duet|Fox Duet|5176|8082' "$dir" | while read -r f; do
	sed -i \
		-e "s#github.com/martin-pouls/fox-duet#github.com/martin-pouls/$dir#g" \
		-e "s#fox-duet#$dir#g" \
		-e "s/Fox Duet/$esc_title/g" \
		-e "s/port: 5176/port: $vport/" \
		-e "s#localhost:8082#localhost:$bport#" \
		-e "s/\":8082\"/\":$bport\"/" \
		"$f"
done

echo "Scaffolded $dir/ from $src/ (${#files[@]} files): module github.com/martin-pouls/$dir/backend, Vite :$vport, backend :$bport."
echo
echo "Files that still carry Fox Duet's game (strip or rewrite them):"
grep -rliIE 'fox|forest|decree|gem|traveller|levelPref|Foxtrot' "$dir" \
	--exclude=package-lock.json --exclude-dir=node_modules --exclude-dir=ui | sed 's/^/  /' || true
echo
echo "Not done here (scaffold work package): npm ci in $dir/frontend, strip the domain to a lobby-only"
echo "game, comments naming other apps' ports, and the integration list in the add-app skill."
