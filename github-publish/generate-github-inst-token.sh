GITHUB_TOKEN=$(curl --request POST \
--url "https://api.github.com/app/installations/$GITHUB_APP_INST_ID/access_tokens" \
--header "Accept: application/vnd.github+json" \
--header "Authorization: Bearer $JWT" \
--header "X-GitHub-Api-Version: 2022-11-28" \
--header "Content-Type: application/json" \
--data '{
  "repositories": ["terraform-provider-teamcity"]
}' | awk -F'"' '/token/{print $4}')

export GITHUB_TOKEN
unset JWT
