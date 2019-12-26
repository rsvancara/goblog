#!/bin/sh
# Push a docker image only if it does not exist in the repository

IMAGE="$1"
REPO=$(echo "$IMAGE" | cut -d/ -f1)
NAME=$(echo "$IMAGE" | cut -d/ -f2- | cut -d: -f1)
TAG=$(echo "$IMAGE" | cut -d/ -f2- | cut -d: -f2)

curl --fail \
  --silent \
  --show-error \
  --location \
  "https://$REPO/v2/$NAME/manifests/$TAG" >/dev/null

case "$?" in
"0")
  echo "$IMAGE has already been pushed"
  ;;
"22")
  # curl exits with code 22 for a HTTP 404
  docker push "$IMAGE"
  ;;
*)
  echo "Unexpected curl exit code: $?"
  exit "$?"
  ;;
esac