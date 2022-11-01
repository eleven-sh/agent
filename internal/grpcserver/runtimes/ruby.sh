#!/bin/bash
# Install Ruby
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

RUBY_VERSION="${RUNTIME_VERSION}"

curl --silent --show-error --location --fail https://rvm.io/mpapis.asc | gpg --import -
curl --silent --show-error --location --fail https://rvm.io/pkuczynski.asc | gpg --import -

rm --recursive --force .rvm
curl --silent --show-error --location --fail https://get.rvm.io | bash -s master

if [[ "${RUBY_VERSION}" == "latest" ]]; then
  bash --login -c " \
  rvm requirements \
  && rvm use ruby --install --default \
  && rvm rubygems current \
  && gem install bundler --no-document"
else
  bash --login -c " \
  rvm requirements \
  && rvm install ${RUBY_VERSION} \
  && rvm use ${RUBY_VERSION} --default \
  && rvm rubygems current \
  && gem install bundler --no-document"
fi

if ! grep --silent --fixed-strings '[[ -s "$HOME/.rvm/scripts/rvm" ]] && . "$HOME/.rvm/scripts/rvm"' .zshrc; then
  echo '' >> .zshrc
  echo '[[ -s "$HOME/.rvm/scripts/rvm" ]] && . "$HOME/.rvm/scripts/rvm"' >> .zshrc
fi
