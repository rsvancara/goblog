#!/bin/bash

#
# Release the current tag to test
#

set -xe

# LATEST TAG
GIT_TAG=$(git describe --abbrev=0 --tags)

#sed -i "s/VERSION/${GIT_TAG}/g" helm/visualintrigue/Chart.yaml
sed -i "s/VERSION/${GIT_TAG}/g" helm/diytinytrailer/Chart.yaml
sed -i "s/VERSION/${GIT_TAG}/g" helm/tryingadventure/Chart.yaml

#helm3 upgrade visualintrigue helm/visualintrigue 
helm3 upgrade diytinytrailer helm/diytinytrailer 
helm3 upgrade tryingadventure helm/tryingadventure 
