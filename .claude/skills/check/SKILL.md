---
name: check
description: Run the repo's "before declaring work done" checks (go vet/test, svelte-check, lint, vitest) for exactly the apps a change touches, in parallel. Use before saying any code change is done, before committing, or when asked to run the tests/checks.
---

# Checks for the affected apps

Run from anywhere in the repo (Bash tool, Git Bash):

```bash
bash .claude/skills/check/check.sh            # apps touched by uncommitted changes
bash .claude/skills/check/check.sh --base origin/main   # also everything committed since origin/main
bash .claude/skills/check/check.sh fox-duet   # named apps; shared = every app, deep-sea = the-crew
bash .claude/skills/check/check.sh --all
```

Per app it runs `go vet ./... && go test ./...` in `backend/` and `npm run check && npm run lint && npm test` in `frontend/`, all apps in parallel. A change in `shared/` checks every app. It takes a few minutes for every app, so run it with a 10-minute timeout.

- With no arguments it sees only uncommitted work. After committing (or in a worktree branch), pass `--base main` or name the apps.
- The output ends with one `ok` / `FAIL` line per part. Failures print the last 60 log lines and the path of the full log.
- `npm run lint` fails on formatting: fix it with `npm run format` in that app's `frontend/`, then run the check again.
- Report failures with their output. Don't call the work done while a part fails, unless the failure was already there before your change (say so).
