#!/bin/bash

#
# Release the current tag to test
#

set -xe

# LATEST TAG
GIT_TAG=$(git describe --abbrev=0 --tags)


helm3 upgrade goblog helm/goblog --set apiVersion=${GIT_TAG}


