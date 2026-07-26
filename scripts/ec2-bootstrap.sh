#!/usr/bin/env bash
# One-time provisioning for a fresh EC2 host (Amazon Linux 2023 or Ubuntu).
# Installs Docker + the compose plugin and enables the service. Run once as the
# deploy user, e.g.:
#   ssh ec2-user@<host> 'bash -s' < scripts/ec2-bootstrap.sh
set -euo pipefail

if command -v dnf >/dev/null 2>&1; then
  # Amazon Linux 2023 / Fedora-family
  sudo dnf -y install docker
  sudo systemctl enable --now docker
  # compose plugin
  sudo mkdir -p /usr/libexec/docker/cli-plugins
  ARCH="$(uname -m)"
  sudo curl -fsSL "https://github.com/docker/compose/releases/latest/download/docker-compose-linux-${ARCH}" \
    -o /usr/libexec/docker/cli-plugins/docker-compose
  sudo chmod +x /usr/libexec/docker/cli-plugins/docker-compose
elif command -v apt-get >/dev/null 2>&1; then
  # Ubuntu / Debian
  sudo apt-get update
  sudo apt-get -y install ca-certificates curl
  sudo install -m 0755 -d /etc/apt/keyrings
  sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
  sudo chmod a+r /etc/apt/keyrings/docker.asc
  . /etc/os-release
  echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu ${VERSION_CODENAME} stable" \
    | sudo tee /etc/apt/sources.list.d/docker.list >/dev/null
  sudo apt-get update
  sudo apt-get -y install docker-ce docker-ce-cli containerd.io docker-compose-plugin
  sudo systemctl enable --now docker
else
  echo "Unsupported package manager. Install Docker + compose plugin manually." >&2
  exit 1
fi

# Allow the current user to run docker without sudo (re-login required to apply).
sudo usermod -aG docker "$USER" || true

echo "Docker installed. Log out and back in so group membership takes effect."
docker --version
docker compose version || true
