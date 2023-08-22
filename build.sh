#!/usr/bin/env sh

set -eux

name=terraform-provider-teamcity
version=0.0.${BUILD_NUMBER:-dev}
dir=build

rm -rf $dir
export CGO_ENABLED=0

for os in linux darwin windows; do
  for arch in amd64 arm64; do
    GOOS=$os GOARCH=$arch go build -trimpath -ldflags="-s -w" -o $dir/${os}_$arch/${name}_v$version .
  done
done
