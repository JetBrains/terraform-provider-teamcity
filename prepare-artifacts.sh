#!/usr/bin/env sh

name=terraform-provider-teamcity
version=0.0.${BUILD_NUMBER:-dev}
dir=build

for os in linux darwin windows; do
  for arch in amd64 arm64; do
    zip $dir/${name}_${version}_${os}_${arch}.zip -j $dir/${os}_$arch/${name}_v$version
  done
done

cp terraform-registry-manifest.json $dir/${name}_${version}_manifest.json

cd $dir
shasum -a 256 *.zip *.json >${name}_${version}_SHA256SUMS
