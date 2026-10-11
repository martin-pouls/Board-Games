# Deployment

Every push to `main` deploys itself through GitHub Actions ([.github/workflows/deploy.yml](../.github/workflows/deploy.yml)):

1. **plan:** picks the apps (`<app>/deploy/compose.yaml`) that changed since the last successful deploy. A change in `shared/` counts for every app whose `Dockerfile` copies it (`COPY shared/...`). A failed or cancelled run is caught up by the next one.
2. **test:** for each app, `go vet` + `go test` and `npm run check` + `lint` + `test`. For apps that use `shared/`, the shared backend and frontend are checked first.
3. **build:** on a native arm64 runner, `docker buildx bake` builds the app from its `compose.yaml` and pushes `ghcr.io/martin-pouls/board-games/<app>:<commit sha>`. Layers are cached in the GitHub Actions cache.
4. **deploy:** SSH as `deploy@yoshi-und-kruemli.de`, sending only the commit, the app names and a pull token. The key can only run [server/board-games-deploy](server/board-games-deploy). That script fetches `main` into its own clone (`/opt/board-games/.repo`), refuses commits that are not on `main`, and takes `<app>/deploy/` from that commit. It pulls the images (with a token that expires when the job ends), writes `TAG=<sha>` into `/opt/board-games/<app>/deploy/.env`, restarts the containers, installs the Caddy snippets and reloads Caddy. The job then fetches every deployed URL.

Images are never built on the server. Tags are immutable (one per commit), so what runs is always a known commit (`/opt/board-games/<app>/REVISION`).

## Security model

- **Dedicated key, dedicated user.** The CI key is used only for this and only by GitHub. On the server it belongs to `deploy`, not `root`, with `restrict,command="/usr/local/bin/board-games-deploy"`. The key file is `/etc/ssh/authorized_keys/deploy`, owned by root, so the account cannot add keys of its own. A leaked key gives no shell, no port forwarding and no other commands.
- **Only commits of `main` can be deployed.** Being in the `docker` group, `deploy` is root-equivalent, so the script trusts nothing from the caller but a commit SHA and app names. Compose files and Caddy snippets come from the server's own clone of `main` (only `main` is fetched, so commits from other branches or forks never get there), and images from ghcr.io, where only this repository's workflows can push. A leaked key can at most redeploy an older version of `main`.
- **SSH:** password logins are off for every user (`/etc/ssh/sshd_config.d/10-board-games.conf`); root logs in with keys only.
- **Environment secret.** The key is stored in the GitHub environment `production`, which only the `main` branch may use. Pull requests (also from forks) never see it, and the deploy job never runs for them.
- **Pinned host key.** The server's host key is stored as a variable and checked strictly (`StrictHostKeyChecking=yes`), with no `ssh-keyscan` at deploy time.
- **Least privilege for the workflow.** `permissions: {}` at the top; each job gets only what it needs (`packages: write` only for build). Actions are pinned to commit SHAs, and Dependabot updates them.
- **No long-lived registry credential on the server.** The job's own `GITHUB_TOKEN` (`packages: read`) is passed over SSH for one pull, kept in a temporary Docker config and deleted afterwards.

## One-time setup

From Git Bash at the repo root, after this directory is committed (setup-server.sh ships `HEAD`):

1. **Make a key pair just for CI** (no passphrase, CI can't type one):
   ```sh
   ssh-keygen -t ed25519 -N "" -C "github-actions board-games deploy" -f ~/.ssh/board-games-ci
   ```
2. **Install the deploy user, script and key on the server** (as root over your own key, as before):
   ```sh
   CI_PUBKEY=~/.ssh/board-games-ci.pub deploy/setup-server.sh
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
6. Push to `main`, or run the workflow by hand (*Actions → Deploy → Run workflow*). The first run deploys every app.

After the first push the images show up under the repository's *Packages*. They may stay private: the server pulls with the job's token. If a deploy fails at `pulling ...` with *denied* or *unauthorized*, set each package to public (*Package settings → Change visibility*); the source is public anyway. A failed pull changes nothing on the server.

## Everyday use

- **Deploy:** push to `main`. Only changed apps are rebuilt and restarted (a change in `shared/` rebuilds the apps that use it).
- **Deploy by hand / redeploy:** *Actions → Deploy → Run workflow*, with `all` or e.g. `regicide start-page`.
- **Roll back:** preferably `git revert` the bad commit and push, so that `main` matches the server. For a quick roll back, open the run of the good commit and click *Re-run all jobs*: it deploys every app at that commit (cached build, same image tags). The next push then deploys only the apps it changes, so run the workflow by hand with `all` to bring the rest forward again. Don't re-run runs from before the rename to Board-Games (October 2026): their start page lives at `/card-games/`, which the current Caddy config redirects away.
- **Server or Caddyfile changes** (`deploy/caddy/`, `deploy/server/`): commit, then re-run `deploy/setup-server.sh` (without `CI_PUBKEY` the deploy key stays as it is, pinned to the current deploy script).
- **Fallback without GitHub:** `deploy/deploy.sh [app...]` uploads the committed `<app>/` and `shared/` and builds on the server as root, tagging the image `local-<rev>`. The next CI deploy replaces it.

## Build context

Apps that use [shared/](../shared/) (their `Dockerfile` has `COPY shared/...`) build with the repository root as context: `context: ../..` and `dockerfile: <app>/Dockerfile` in `compose.yaml`, with `<app>/Dockerfile.dockerignore` keeping the context down to `<app>/` and `shared/`. The image keeps the repo layout (`/src/<app>`, `/src/shared`), so `go.mod`'s `replace => ../../shared/backend` and the frontend's alias to `../../shared/frontend/src` resolve as in a checkout. On the server the context resolves to `/opt/board-games`, but images are only pulled there, never built (except by `deploy.sh`, which uploads `<app>/` and `shared/` for that).
