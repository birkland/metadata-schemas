#!/bin/sh

source $(dirname $0)/.definitions.sh

DATE=`date -I`
COMMITHASH=`git rev-parse HEAD | cut -c 1-8`
GIT_TAG=`git describe --tags 2>/dev/nul`
DOCKER_TAG="${DATE}-${COMMITHASH}-SNAPSHOT"

if [ -n "$GIT_TAG" ]; then 
    DOCKER_TAG=${DATE}-${GIT_TAG}
fi

docker tag ${DOCKER_REPO_NAME}:latest ${DOCKER_REPO_NAME}:${DOCKER_TAG}

docker push ${DOCKER_REPO_NAME}:latest
docker push ${DOCKER_REPO_NAME}:${DOCKER_TAG}

