#!/bin/bash
# Install Python
set -euo pipefail

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

# Force LibSSL to 1.1.1 to avoid conflicts 
# with old Ruby and Python versions
sudo apt-add-repository --yes ppa:rael-gc/rvm
sudo apt-get --assume-yes --quiet --quiet update
sudo apt-get --assume-yes --quiet --quiet remove libssl-dev

sudo rm --recursive --force /etc/apt/preferences.d/rael-gc-rvm-precise-pin-900
sudo touch /etc/apt/preferences.d/rael-gc-rvm-precise-pin-900
echo 'Package: *' | sudo tee /etc/apt/preferences.d/rael-gc-rvm-precise-pin-900 > /dev/null
echo 'Pin: release o=LP-PPA-rael-gc-rvm' | sudo tee --append /etc/apt/preferences.d/rael-gc-rvm-precise-pin-900 > /dev/null
echo 'Pin-Priority: 900' | sudo tee --append /etc/apt/preferences.d/rael-gc-rvm-precise-pin-900 > /dev/null

sudo apt-get --assume-yes --quiet --quiet install libssl-dev

cd ~

PYTHON_VERSION="${RUNTIME_VERSION}"

sudo apt-get --assume-yes --quiet --quiet update
sudo apt-get --assume-yes --quiet --quiet install \
  libbz2-dev \
  libffi-dev \
  liblzma-dev \
  libncursesw5-dev \
  libreadline-dev \
  libsqlite3-dev \
  libxml2-dev \
  libxmlsec1-dev \
  llvm \
  make \
  tk-dev \
  xz-utils \
  zlib1g-dev

rm --recursive --force .pyenv
curl --silent --show-error --location --fail https://github.com/pyenv/pyenv-installer/raw/master/bin/pyenv-installer | bash

if ! grep --silent --fixed-strings 'export PATH=$PATH:$HOME/.pyenv/bin:$HOME/.pyenv/shims' .zshrc; then
  echo '' >> .zshrc
  echo 'export PATH=$PATH:$HOME/.pyenv/bin:$HOME/.pyenv/shims' >> .zshrc
fi

if ! grep --silent --fixed-strings 'eval "$(pyenv init -)"' .zshrc; then
  echo 'eval "$(pyenv init -)"' >> .zshrc
fi

if ! grep --silent --fixed-strings 'eval "$(pyenv virtualenv-init -)"' .zshrc; then
  echo 'eval "$(pyenv virtualenv-init -)"' >> .zshrc
fi

sudo --set-home --login --user eleven -- env \
	PYTHON_VERSION="${PYTHON_VERSION}" \
zsh << 'EOF'

set -euo pipefail

source ~/.zshrc

if [[ "${PYTHON_VERSION}" == "latest" ]]; then
  PYTHON_VERSION="$(pyenv install --list | grep --extended-regexp '^\s*[0-9][0-9.]*[0-9]\s*$' | tail -1 | xargs)"
fi

pyenv install "${PYTHON_VERSION}"
pyenv global "${PYTHON_VERSION}"

pip install virtualenv pipenv

EOF
