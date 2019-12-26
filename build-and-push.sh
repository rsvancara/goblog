#!/bin/sh
# Build and push a docker image
set -e

NAME=$1
TAG=$2

if [ -z "$TAG" ]; then
  echo "Building untagged image"
  docker build .
else
  docker build -t "$NAME":"$TAG" .
  ./docker-push.sh "$NAME":"$TAG"
fi