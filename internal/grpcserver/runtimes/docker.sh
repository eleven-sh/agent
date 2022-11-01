#!/bin/bash
# Install Docker
set -o errexit
set -o nounset
set -o pipefail
set -o xtrace

log () {
  echo -e "${1}" >&2
}

# Remove "debconf: unable to initialize frontend: Dialog" warnings
echo 'debconf debconf/frontend select Noninteractive' | sudo debconf-set-selections

handleExit () {
  EXIT_CODE=$?
  exit "${EXIT_CODE}"
}

trap "handleExit" EXIT

cd ~

# Install docker
if [[ ! -f "/usr/share/keyrings/docker-archive-keyring.gpg" ]]; then
  curl --fail --silent --show-error --location https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor --output /usr/share/keyrings/docker-archive-keyring.gpg
fi

if [[ ! -f "/etc/apt/sources.list.d/docker.list" ]]; then
	echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release --codename --short) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
fi

sudo apt-get --assume-yes --quiet --quiet update
sudo apt-get --assume-yes --quiet --quiet remove docker docker-engine docker.io containerd runc
sudo apt-get --assume-yes --quiet --quiet install docker-ce docker-ce-cli containerd.io

# Install Docker compose
LATEST_COMPOSE_VERSION=$(curl --fail --silent --show-error --location "https://api.github.com/repos/docker/compose/releases/latest" | grep --only-matching --perl-regexp '(?<="tag_name": ").+(?=")')
sudo curl --fail --silent --show-error --location "https://github.com/docker/compose/releases/download/${LATEST_COMPOSE_VERSION}/docker-compose-$(uname --kernel-name)-$(uname --machine)" --output /usr/libexec/docker/cli-plugins/docker-compose
sudo chmod +x /usr/libexec/docker/cli-plugins/docker-compose
