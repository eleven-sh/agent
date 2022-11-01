#!/bin/bash
# Install Node.js
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

# NVM uses the "NODE_VERSION" env var 
# to install node during install.
# We don't want that as we want to install 
# it manually for better error management.
NODE_VERSION="${RUNTIME_VERSION}"

if [[ "${NODE_VERSION}" == "latest" ]]; then
  NODE_VERSION="node"
fi

rm --recursive --force .nvm
curl --silent --show-error --location --fail https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.1/install.sh | bash

sudo --set-home --login --user eleven -- env \
	NODE_VERSION="${NODE_VERSION}" \
zsh << 'EOF'

set -euo pipefail

source ~/.zshrc

nvm install "${NODE_VERSION}"

npm config set python /usr/bin/python --global
npm config set python /usr/bin/python
npm install -g typescript
npm install -g yarn

EOF
