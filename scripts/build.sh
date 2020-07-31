#!/bin/bash


#
# Builds a tagged release of the docker file
# author: Randall Svancara
#
#

set -xe 

# LATEST TAG
GIT_TAG=$(git describe --abbrev=0 --tags)

# vdub
/usr/bin/vdub

# COPY Artifacts
cp /usr/share/GeoIP/*.mmdb db/

# Build the docker image
docker build -t rsvancara/goblog:${GIT_TAG} --no-cache . 

# Push to repository
docker push rsvancara/goblog:${GIT_TAG}
