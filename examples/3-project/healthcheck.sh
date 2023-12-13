#!/usr/bin/env sh
set -eu

if [ $# -ne 1 ]
  then
    echo "Usage: $0 <hostname>"
    exit 1
fi

URL=$1
SLEEP=${SLEEP:-5} # 5 seconds
TIMEOUT=${TIMEOUT:-300} # 5 minutes

# BusyBox sets a common timeout for whole request,
# GNU interprets it as `--dns-timeout 5 --connect-timeout 5 --read-timeout 5`
COMMAND="wget --spider --timeout=5"
COMMAND="$COMMAND --tries=1 -nv" # additional options for GNU wget, BusyBox just ignores them
MAX_ATTEMPTS=$(( $TIMEOUT / $SLEEP )) # rounds down

if [ -n "${TOKEN:+exist}" ]; then
  AUTH="Authorization: Bearer $TOKEN"
  COMMAND="$COMMAND --header=\"$AUTH\""
elif [ -n "${HTTP_PASSWORD:+exist}" ]; then
  # `echo -n` does not work on macOS. But the server accepts it even with a newline.
  AUTH="Authorization: Basic $(echo ${HTTP_USERNAME:-}:$HTTP_PASSWORD | base64)"
  COMMAND="$COMMAND --header=\"$AUTH\""
fi

i=1
while true
do
    if eval "$COMMAND $URL"; then
        echo "Service is healthy."
        exit 0
    else
      if [ $i -ge $MAX_ATTEMPTS ]; then
        echo "Service health check failed after $MAX_ATTEMPTS attempts."
        exit 1
      fi
      sleep $SLEEP
      i=$(( $i + 1 ))
    fi
done
