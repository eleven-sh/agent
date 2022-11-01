#!/bin/bash
# Install Go
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

GO_VERSION="${RUNTIME_VERSION}"

if [[ "${GO_VERSION}" == "latest" ]]; then
  GO_VERSION=$(curl --fail --silent --show-error --location "https://go.dev/VERSION?m=text")
else
  GO_VERSION="go${RUNTIME_VERSION}"
fi

ARCH=$(arch | sed s/aarch64/arm64/ | sed s/x86_64/amd64/)

curl --fail --silent --show-error --location "https://go.dev/dl/${GO_VERSION}.linux-${ARCH}.tar.gz" --output /tmp/go.tar.gz
sudo rm --recursive --force /usr/local/go
sudo tar --directory /usr/local --extract --file /tmp/go.tar.gz
rm --recursive --force /tmp/go.tar.gz

if ! grep --silent --fixed-strings 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' .zshrc; then
  echo '' >> .zshrc
  echo 'export PATH=$PATH:/usr/local/go/bin:$HOME/go/bin' >> .zshrc
fi

sudo --set-home --login --user eleven -- << 'EOF'

set -euo pipefail

source ~/.zshrc

go install github.com/ramya-rao-a/go-outline@latest
go install github.com/cweill/gotests/gotests@latest
go install github.com/fatih/gomodifytags@latest
go install github.com/josharian/impl@latest
go install github.com/haya14busa/goplay/cmd/goplay@latest
go install github.com/go-delve/delve/cmd/dlv@latest
go install honnef.co/go/tools/cmd/staticcheck@latest
go install golang.org/x/tools/gopls@latest

EOF
