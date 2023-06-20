#!/usr/bin/env sh

set -eux

apk add --no-cache gpg-agent
printenv GPG_PRIVATE_KEY | gpg --batch --import
goreleaser release --clean
