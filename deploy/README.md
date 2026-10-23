# Deployment

The site answers on `yoshi-und-kruemli.de` and `martin-pouls.de` (same Caddy site block, so every app is on both; the `www.` names redirect). Both domains' DNS A records point at the server, and Caddy gets the certificates itself.

Every push to `main` deploys itself through GitHub Actions ([.github/workflows/deploy.yml](../.github/workflows/deploy.yml)):

1. **plan:** picks the apps (`<app>/deploy/compose.yaml`) that changed since the last successful deploy. An app depends on its own directory and on every top-level directory its `Dockerfile` copies from (`COPY shared/...`), where the Dockerfile is the one named by `dockerfile:` in its `compose.yaml`: a change in `shared/` rebuilds every app that copies it, and a change in `the-crew/` rebuilds both `the-crew` and `deep-sea` (which has no code of its own and builds `the-crew/Dockerfile` with `EDITION=deepSea`). A change in `deploy/caddy/` deploys the shared Caddy proxy (`caddy`). A failed or cancelled run is caught up by the next one.
2. **test:** for each app, `go vet` + `go test` and `npm run check` + `lint` + `test`. **test-shared** checks `shared/backend` and `shared/frontend` the same way, once per run and only when they changed (on manual runs: when a chosen app uses them).
3. **build**, at the same time as the tests: on a native arm64 runner, `docker buildx bake` builds the app from its `compose.yaml` and pushes `ghcr.io/martin-pouls/board-games/<app>:<commit sha>`. Layers are cached in the GitHub Actions cache.
4. **deploy**, once all tests and builds passed: SSH as `deploy@yoshi-und-kruemli.de`, sending only the commit, the app names and a pull token. The key can only run [server/board-games-deploy](server/board-games-deploy). That script fetches `main` into its own clone (`/opt/board-games/.repo`), refuses commits that are not on `main`, and takes `<app>/deploy/` from that commit. It pulls the images (with a token that expires when the job ends) and validates the Caddy config that will result (new snippets, and for `caddy` the new `Caddyfile`, `common.caddy` and image) with `caddy validate`; a failed pull or an invalid config changes nothing. Then it writes `TAG=<sha>` into `/opt/board-games/<app>/deploy/.env`, installs the snippets (and `deploy/caddy/` into `/opt/caddy`), starts Caddy (a new image recreates it: every app is unreachable for a second and open WebSockets reconnect), restarts the app containers and reloads Caddy. The job then fetches every deployed URL. Behind the [password gate](#password-gate) that only proves Caddy and the gate answer (a redirect to the login counts as up), not the app itself.

Images are never built on the server. Tags are immutable (one per commit), so what runs is always a known commit (`/opt/board-games/<app>/REVISION`, `/opt/caddy/REVISION`).

## Security model

- **Dedicated key, dedicated user.** The CI key is used only for this and only by GitHub. On the server it belongs to `deploy`, not `root`, with `restrict,command="/usr/local/bin/board-games-deploy"`. The key file is `/etc/ssh/authorized_keys/deploy`, owned by root, so the account cannot add keys of its own. A leaked key gives no shell, no port forwarding and no other commands.
- **Only commits of `main` can be deployed.** Being in the `docker` group, `deploy` is root-equivalent, so the script trusts nothing from the caller but a commit SHA and app names. Compose files and the Caddy config come from the server's own clone of `main` (only `main` is fetched, so commits from other branches or forks never get there), and images from ghcr.io, where only this repository's workflows can push. A leaked key can at most redeploy an older version of `main`.
- **SSH:** password logins are off for every user (`/etc/ssh/sshd_config.d/10-board-games.conf`); root logs in with keys only.
- **Logs for 24 hours at most** (data minimisation). Caddy writes no access log, and the apps log no IP addresses. Containers log to journald (`docker logs` and `journalctl` still work), which keeps entries 22 hours (`/etc/systemd/journald.conf.d/10-board-games.conf`), and `/etc/cron.hourly/board-games-logs` rotates and vacuums the journal every hour. Nothing goes to syslog, and ufw logging is off. Debug a problem the same day.
- **Environment secret.** The key is stored in the GitHub environment `production`, which only the `main` branch may use. Pull requests (also from forks) never see it, and the deploy job never runs for them.
- **Pinned host key.** The server's host key is stored as a variable and checked strictly (`StrictHostKeyChecking=yes`), with no `ssh-keyscan` at deploy time.
- **Least privilege for the workflow.** `permissions: {}` at the top; each job gets only what it needs (`packages: write` only for build). Actions are pinned to commit SHAs, and Dependabot updates them.
- **No long-lived registry credential on the server.** The job's own `GITHUB_TOKEN` (`packages: read`) is passed over SSH for one pull, kept in a temporary Docker config and deleted afterwards.

## Password gate

The start page and the games ask for a shared password ([gate/](../gate/CLAUDE.md)). The site root and other paths stay public. Caddy asks the gate before every request to a protected app (`(gate)` snippet in [caddy/conf/common.caddy](caddy/conf/common.caddy); an app opts in with `import gate` in its own snippet). The gate checks the login cookie: page loads without one are redirected to `/gate/login`, and other requests get `401`.

- **Several passwords**, one per line in a file on your PC, each with a label: `family=…`, `friends=…`. A cookie is tied to its label: changing or removing one line logs out only those who used it.
- **Install or change them:** `GATE_PASSWORDS=~/gate-passwords deploy/setup-server.sh` (the file travels inside the SSH stream; keep it out of the repo). The server copy is `/etc/board-games/gate/passwords`, mode 600, owned by the container's user (uid 65532). Without `GATE_PASSWORDS` the installed file stays; the first run fails without it.
- **The secret** (`/etc/board-games/gate/secret`) is generated by setup-server.sh once and kept. Deleting it and re-running logs everyone out.
- **First rollout:** run setup-server.sh with the passwords *before* pushing the apps that `import gate`. Without the `(gate)` snippet Caddy refuses the new config, and without the gate container the protected paths answer 502.

## One-time setup

From Git Bash at the repo root, after this directory is committed (setup-server.sh ships `HEAD`):

1. **Make a key pair just for CI** (no passphrase, CI can't type one):
   ```sh
   ssh-keygen -t ed25519 -N "" -C "github-actions board-games deploy" -f ~/.ssh/board-games-ci
   ```
2. **Install Docker, the deploy user, script and key and the gate passwords on the server** (as root over your own key, as before; the passwords file is described under [Password gate](#password-gate)):
   ```sh
   CI_PUBKEY=~/.ssh/board-games-ci.pub GATE_PASSWORDS=~/gate-passwords deploy/setup-server.sh
   ```
3. **Get the server's host key line** over your already trusted root connection:
   ```sh
   echo "yoshi-und-kruemli.de $(ssh -i ~/.ssh/hetzner_ed25519 root@yoshi-und-kruemli.de "cut -d' ' -f1,2 /etc/ssh/ssh_host_ed25519_key.pub")"
   ```
4. **On GitHub:** *Settings → Environments → New environment* `production`:
   - *Deployment branches and tags:* **Selected branches** → `main`.
   - *Environment secret* `DEPLOY_SSH_KEY`: the whole content of `~/.ssh/board-games-ci` (the private key, including the `BEGIN`/`END` lines).
   - *Environment variable* `DEPLOY_KNOWN_HOSTS`: the line from step 3.
   - Optional: *Required reviewers* if every deploy should wait for a click.
5. **Remove the private key from your PC** once it is in GitHub (`rm ~/.ssh/board-games-ci`); keep the `.pub`. To rotate, repeat steps 1, 2 and 4.
6. Push to `main`, or run the workflow by hand (*Actions → Deploy → Run workflow*). The first run deploys Caddy and every app.

After the first push the images show up under the repository's *Packages*. They may stay private: the server pulls with the job's token. If a deploy fails at `pulling ...` with *denied* or *unauthorized*, set each package to public (*Package settings → Change visibility*); the source is public anyway. A failed pull changes nothing on the server.

## Everyday use

- **Deploy:** push to `main`. Only changed apps are rebuilt and restarted (a change in `shared/` rebuilds the apps that use it).
- **Deploy by hand / redeploy:** *Actions → Deploy → Run workflow*, with `all` or e.g. `regicide start-page` (`caddy` for the proxy).
- **Roll back:** preferably `git revert` the bad commit and push, so that `main` matches the server. For a quick roll back, open the run of the good commit and click *Re-run all jobs*: it deploys every app at that commit (cached build, same image tags). The next push then deploys only the apps it changes, so run the workflow by hand with `all` to bring the rest forward again. Don't re-run runs from before the rename to Board-Games (October 2026): their start page lives at `/card-games/`, which the current Caddy config redirects away.
- **Caddy changes** (`deploy/caddy/`: config or image): push, like an app.
- **Server changes** (`deploy/server/`, `deploy/setup-server.sh`): commit, then re-run `deploy/setup-server.sh` (without `CI_PUBKEY` the deploy key stays as it is, pinned to the current deploy script; without `GATE_PASSWORDS` the gate passwords stay). CI can't update its own deploy script.
- **Base images** (the `FROM` lines and the Caddy image in `deploy/caddy/` and `deploy/local/`) are pinned by digest. Dependabot opens a grouped pull request each month to move the pins. Merging deploys the apps and Caddy with the new pins.
- **Old images:** [cleanup-images.yml](../.github/workflows/cleanup-images.yml) runs every Monday and deletes images on ghcr.io older than a week, keeping each app's newest one (the running one). A manual run is a dry run unless unchecked. Re-running a deploy whose images were deleted rebuilds them.
- **Fallback without GitHub:** `deploy/deploy.sh [app...]` uploads the committed `<app>/` and `shared/` and builds on the server as root, tagging the image `local-<rev>`; `caddy` installs the committed `deploy/caddy/`. The next CI deploy replaces it.

## Local stack

`deploy/local.sh` runs the production setup on your PC: every app's image built from its `compose.yaml`, behind a real Caddy with the server's site config (`caddy/conf/common.caddy`: the gate, the snippets, the redirects) at http://localhost:8000. Only the site address differs (`local/Caddyfile`, plain http, no `www.` blocks). It needs WSL2 with Docker; Windows 11 Home has no Docker otherwise.

1. Once, as admin in Windows: `wsl --install -d Ubuntu`, reboot, start *Ubuntu* and create the Linux user.
2. Once, in Ubuntu from the repo (`/mnt/c/Users/<you>/github/Board-Games`): `deploy/local/setup-wsl.sh` (Docker, same packages as the server), then `wsl --shutdown` and reopen Ubuntu.
3. `deploy/local.sh` (all apps) or `deploy/local.sh regicide gate` (only these), then open http://localhost:8000. The gate password is `local`; edit `local/.gate/passwords` (`label=password` lines, gitignored) and run `deploy/local.sh gate` to change it. `deploy/local.sh down` stops everything; game databases stay in Docker volumes.

It builds the working tree, uncommitted changes included. Use the browser on Windows at `localhost`: on another host name the gate's `Secure` cookie needs HTTPS.

- WSL stops its VM shortly after the last Ubuntu window closes, and the stack with it (the site then looks "crashed"; `docker ps` shows every container as *Up 1 second* once WSL starts again). Keep a window open while testing, or keep WSL alive from Windows with `wsl -d Ubuntu -- sleep infinity` (e.g. in a spare terminal); the containers come back on their own when WSL starts again.
- Every local request reaches the gate from the same Docker address, so after 10 wrong passwords the login is rate limited for all browsers. `docker restart gate` resets it.

## Build context

Apps that use [shared/](../shared/) (their `Dockerfile` has `COPY shared/...`) build with the repository root as context: `context: ../..` and `dockerfile: <app>/Dockerfile` in `compose.yaml`, with `<app>/Dockerfile.dockerignore` keeping the context down to `<app>/` and `shared/`. The image keeps the repo layout (`/src/<app>`, `/src/shared`), so `go.mod`'s `replace => ../../shared/backend` and the frontend's alias to `../../shared/frontend/src` resolve as in a checkout. On the server the context resolves to `/opt/board-games`, but images are only pulled there, never built (except by `deploy.sh`, which uploads the directories the Dockerfile copies from for that).

**Deep Sea** (`/deep-sea/`) is the one app without a Dockerfile of its own: `deep-sea/deploy/compose.yaml` builds `the-crew/Dockerfile` (`dockerfile: the-crew/Dockerfile`) with the build args `BASE_PATH=/deep-sea`, `EDITION=deepSea` and `DB_FILE=deep-sea.db`, and pushes the result as `board-games/deep-sea`. It has its own container, volume, Caddy snippet and gate protection, and is redeployed together with `the-crew` and whenever `shared/` changes. The server-side deploy script needs nothing special for it: it only reads `deep-sea/deploy` and pulls the image.
