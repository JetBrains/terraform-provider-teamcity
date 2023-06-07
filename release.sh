#!/usr/bin/env sh

set -eux

apk add --no-cache gpg-agent
printenv GPG_PRIVATE_KEY | gpg --batch --import
git tag "v0.0.$BUILD_NUMBER"
goreleaser release --clean
