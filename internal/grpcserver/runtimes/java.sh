#!/bin/bash
# Install Java
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

sudo add-apt-repository --yes ppa:linuxuprising/java
sudo apt-get --assume-yes --quiet --quiet update

echo 'oracle-java17-installer shared/accepted-oracle-license-v1-3 select true' | sudo debconf-set-selections
sudo apt-get install --assume-yes --quiet --quiet gradle oracle-java17-installer

MAVEN_VERSION=3.8.6
MAVEN_HOME=/usr/share/maven

sudo rm --recursive --force "${MAVEN_HOME}"
sudo mkdir --parents "${MAVEN_HOME}"

curl --silent --show-error --location --fail https://apache.osuosl.org/maven/maven-3/"${MAVEN_VERSION}"/binaries/apache-maven-"${MAVEN_VERSION}"-bin.tar.gz --output /tmp/maven.tar.gz
sudo tar --extract --gzip --directory "${MAVEN_HOME}" --strip-components=1 --file /tmp/maven.tar.gz

rm --recursive --force /tmp/maven.tar.gz

if ! grep --silent --fixed-strings 'export PATH=$PATH:'"${MAVEN_HOME}"'/bin' .zshrc; then
  echo '' >> .zshrc
  echo 'export PATH=$PATH:'"${MAVEN_HOME}"'/bin' >> .zshrc
fi
