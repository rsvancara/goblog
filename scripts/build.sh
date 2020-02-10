#!/bin/bash


#
# Builds a tagged release of the docker file
# author: Randall Svancara
#
#

set -xe

# LATEST TAG
GIT_TAG=$(git describe --abbrev=0 --tags)

# COPY Artifacts
cp /home/artifacts/geoip/*.mmdb db/

# Build the docker image
docker build -t rsvancara/goblog:${GIT_TAG} .

# Push to repository
docker push rsvancara/goblog:${GIT_TAG}
