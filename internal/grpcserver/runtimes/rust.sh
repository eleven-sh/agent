#!/bin/bash
# Install Rust
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

RUST_VERSION="${RUNTIME_VERSION}"

if [[ -f ".cargo/bin/rustup" ]]; then
	.cargo/bin/rustup self uninstall -y
fi
rm --recursive --force .cargo
rm --recursive --force .zfunc/_rustup

if [[ "${RUST_VERSION}" == "latest" ]]; then
  curl --proto '=https' --tlsv1.2 --silent --show-error --location --fail https://sh.rustup.rs | sh -s -- -y
else
  curl --proto '=https' --tlsv1.2 --silent --show-error --location --fail https://sh.rustup.rs | sh -s -- -y --default-toolchain "${RUST_VERSION}"
fi

.cargo/bin/rustup component add rls-preview rust-analysis rust-src
.cargo/bin/rustup completions zsh > .zfunc/_rustup
