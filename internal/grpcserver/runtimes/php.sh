#!/bin/bash
# Install PHP
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

PHP_VERSION="${RUNTIME_VERSION}"
PHP_PREFIX="php"

if [[ "${PHP_VERSION}" != "latest" ]]; then
  PHP_PREFIX="php${PHP_VERSION}"
  sudo apt-add-repository --yes ppa:ondrej/php
fi

sudo apt-get --assume-yes --quiet --quiet update
sudo apt-get --assume-yes --quiet --quiet install \
  composer \
  "${PHP_PREFIX}"-apcu \
  "${PHP_PREFIX}"-cli \
  "${PHP_PREFIX}"-ctype \
  "${PHP_PREFIX}"-curl \
  "${PHP_PREFIX}"-dom \
  "${PHP_PREFIX}"-fileinfo \
  "${PHP_PREFIX}"-gd \
  "${PHP_PREFIX}"-iconv \
  "${PHP_PREFIX}"-imagick \
  "${PHP_PREFIX}"-intl \
  "${PHP_PREFIX}"-mbstring \
  "${PHP_PREFIX}"-mysql \
  "${PHP_PREFIX}"-mysqli \
  "${PHP_PREFIX}"-opcache \
  "${PHP_PREFIX}"-pdo \
  "${PHP_PREFIX}"-pgsql \
  "${PHP_PREFIX}"-phar \
  "${PHP_PREFIX}"-posix \
  "${PHP_PREFIX}"-simplexml \
  "${PHP_PREFIX}"-sqlite3 \
  "${PHP_PREFIX}"-tokenizer \
  "${PHP_PREFIX}"-xml \
  "${PHP_PREFIX}"-xmlwriter \
  "${PHP_PREFIX}"-zip
