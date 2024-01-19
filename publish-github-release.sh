#!/usr/bin/env sh

set -eu

owner="JetBrains"
repo="terraform-provider-teamcity"
version=v0.0.${BUILD_NUMBER:-dev}

release=$(cat <<EOF
{
  "tag_name": "$version",
  "name": "$version",
  "draft": true
}
EOF
)

echo "Creating a release draft..."
response=$(curl -s -S \
  -X POST \
  -H "Authorization: Bearer $GITHUB_TOKEN" \
  https://api.github.com/repos/$owner/$repo/releases \
  -d "$release"
)

id=$(echo $response | jq '.id')
echo "id=$id"

cd build
for file in *
do
  if [ -f "$file" ] # skip subdirectories
  then
    mime=$(file --mime-type -b $file)
    echo "Uploading $file..."
    curl -s -S \
      -X POST \
      -H "Authorization: Bearer $GITHUB_TOKEN" \
      -H "Content-Type: $mime" \
      "https://uploads.github.com/repos/$owner/$repo/releases/$id/assets?name=$file" \
      --data-binary @$file
  fi
done

echo "Publishing the release..."
curl -s -S -o /dev/null \
  -X PATCH \
  -H "Authorization: Bearer $GITHUB_TOKEN" \
  https://api.github.com/repos/$owner/$repo/releases/$id \
  -d '{"draft":false}'

unset GITHUB_TOKEN

echo "Done."
