#!/bin/bash
# 
# Eleven agent instance init.
# 
# This is the second script to run during the 
# creation of the instance (after the cloud-init one).
# 
# In a nutshell, this script:
#   - change the instance hostname
#   - install required system packages (like Caddy)
#   - configure the user "eleven"
# 
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

# -- Setting hostname

sudo hostnamectl set-hostname "${ENV_NAME_SLUG}"
echo -e "\n127.0.0.1 ${ENV_NAME_SLUG}" | sudo tee --append /etc/hosts > /dev/null

# -- Installing required packages dependencies

log "Installing packages dependencies"

sudo apt-get --assume-yes --quiet --quiet update
sudo apt-get --assume-yes --quiet --quiet install \
  apt-transport-https \
  build-essential \
  ca-certificates \
  curl \
  debian-archive-keyring \
  debian-keyring \
  git \
  gnupg \
  grep \
  jq \
  locales \
  lsb-release \
  software-properties-common \
  tzdata \
  wget

# -- Setting locales / timezone

sudo locale-gen en_US
sudo locale-gen en_US.UTF-8

sudo update-locale
sudo timedatectl set-timezone Etc/UTC

# -- Installing required packages

log "Installing required packages"

# Caddy
if [[ ! -f "/usr/share/keyrings/caddy-stable-archive-keyring.gpg" ]]; then
  curl --fail --silent --show-error --location --tlsv1 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor --output /usr/share/keyrings/caddy-stable-archive-keyring.gpg
fi

if [[ ! -f "/etc/apt/sources.list.d/caddy-stable.list" ]]; then
  curl --fail --silent --show-error --location --tlsv1 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list > /dev/null
fi

sudo apt-get --assume-yes --quiet --quiet update
sudo apt-get --assume-yes --quiet --quiet install caddy

sudo systemctl disable --now caddy
sudo systemctl enable --now caddy-api

# -- Configuring the user "eleven"

log "Configuring the user \"eleven\""

cd ~

# Installing ZSH
rm --recursive --force .zshrc
rm --recursive --force .zfunc
rm --recursive --force .zprofile
rm --recursive --force .zshenv
sudo apt-get --assume-yes --quiet --quiet install zsh
mkdir --parents .zfunc

# Installing OhMyZSH and some plugins
rm --recursive --force .oh-my-zsh
sh -c "$(curl --fail --silent --show-error --location https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)"
git clone --quiet https://github.com/zsh-users/zsh-autosuggestions ${ZSH_CUSTOM:-~/.oh-my-zsh/custom}/plugins/zsh-autosuggestions
git clone --quiet https://github.com/zsh-users/zsh-syntax-highlighting.git ${ZSH_CUSTOM:-~/.oh-my-zsh/custom}/plugins/zsh-syntax-highlighting

# Changing default shell for the user "eleven"
sudo usermod --shell $(which zsh) eleven

# Adding ".zshrc", ".zprofile" and ".zshenv" to home folder
curl --silent --show-error --location --fail https://raw.githubusercontent.com/eleven-sh/sandbox-base/main/zsh/.zshrc --output .zshrc
curl --silent --show-error --location --fail https://raw.githubusercontent.com/eleven-sh/sandbox-base/main/zsh/.zprofile --output .zprofile
curl --silent --show-error --location --fail https://raw.githubusercontent.com/eleven-sh/sandbox-base/main/zsh/.zshenv --output .zshenv

# Creating home folder required directories
mkdir --parents workspace
mkdir --parents .ssh
mkdir --parents .vscode-server

sudo chown --recursive eleven:eleven ~

# Adding VSCode configuration directory
mkdir --parents "${VSCODE_CONFIG_DIR_PATH}"
sudo chown --recursive eleven:eleven "${ELEVEN_CONFIG_DIR_PATH}"
sudo chmod 700 "${VSCODE_CONFIG_DIR_PATH}"

# Adding GitHub SSH keys
if [[ ! -f ".ssh/eleven-github" ]]; then
	ssh-keygen -t ed25519 -C "${GITHUB_USER_EMAIL}" -f .ssh/eleven-github -q -N ""
fi

chmod 644 .ssh/eleven-github.pub
chmod 600 .ssh/eleven-github

if ! grep --silent --fixed-strings "IdentityFile ~/.ssh/eleven-github" .ssh/config; then
	rm --force .ssh/config
  echo "Host github.com" >> .ssh/config
	echo "  User git" >> .ssh/config
	echo "  Hostname github.com" >> .ssh/config
	echo "  PreferredAuthentications publickey" >> .ssh/config
	echo "  IdentityFile ~/.ssh/eleven-github" >> .ssh/config
fi

chmod 600 .ssh/config

if ! grep --silent --fixed-strings "github.com" .ssh/known_hosts; then
  ssh-keyscan github.com >> .ssh/known_hosts
fi

# Configuring Git
git config --global pull.rebase false

git config --global user.name "${USER_FULL_NAME}"
git config --global user.email "${GITHUB_USER_EMAIL}"
