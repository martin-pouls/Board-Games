#!/usr/bin/env bash
# One-time setup of the local stack (deploy/local.sh) inside WSL2 Ubuntu: the same Docker packages
# as the server (deploy/setup-server.sh). Run from the repo (e.g. /mnt/c/Users/<you>/github/...),
# then log out of WSL and in again (wsl --shutdown) so the docker group applies.
set -euo pipefail

sudo apt-get update -qq
sudo DEBIAN_FRONTEND=noninteractive apt-get install -y -qq docker.io docker-compose-v2 docker-buildx git openssl >/dev/null
sudo systemctl enable --now docker >/dev/null
sudo usermod -aG docker "$USER"
# Windows checkouts belong to another owner as seen from WSL; deploy/local.sh calls git.
git config --global --get-all safe.directory | grep -qx '\*' || git config --global --add safe.directory '*'
echo "done: run 'wsl --shutdown' in Windows, reopen Ubuntu, then deploy/local.sh"
