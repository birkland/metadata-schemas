#!/bin/sh

DOCKER_REPO_NAME="oapass/schema-service"

if [ ! -z ${DOCKER_USERNAME+x} ]; then
  echo "${DOCKER_PASSWORD}" | docker login -u "${DOCKER_USERNAME}" --password-stdin
fi
	
# If this is a tag, push a tag.  Otherwise, push to latest
GIT_TAG=`git describe --tags 2>/dev/null`
if [ -n "$GIT_TAG" ]; then
    DOCKER_TAG=${GIT_TAG}
    docker tag ${DOCKER_REPO_NAME}:latest ${DOCKER_REPO_NAME}:${DOCKER_TAG}
    docker push ${DOCKER_REPO_NAME}:${DOCKER_TAG}
fi

docker push ${DOCKER_REPO_NAME}:latest

