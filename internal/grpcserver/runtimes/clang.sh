#!/bin/bash
# Install Clang (C/C++)
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

curl --silent --show-error --location --fail https://apt.llvm.org/llvm-snapshot.gpg.key | sudo apt-key add -

sudo apt-add-repository --yes "deb http://apt.llvm.org/jammy/ llvm-toolchain-jammy main"

sudo apt-get install --assume-yes --quiet --quiet \
  clang-format \
  clang-tools \
  cmake \
  clangd-14

sudo update-alternatives --install /usr/bin/clangd clangd /usr/bin/clangd-14 100
