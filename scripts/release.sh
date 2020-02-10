#!/bin/bash

#
# Release the current tag to test
#

set -xe

# LATEST TAG
GIT_TAG=$(git describe --abbrev=0 --tags)

# Update helm chart with current tag
echo "apiVersion: ${GIT_TAG}" >> helm/goblog/Chart.yaml

helm3 upgrade goblog helm/goblog


