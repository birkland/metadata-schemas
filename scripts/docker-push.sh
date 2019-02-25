#!/bin/sh

source $(dirname $0)/.definitions.sh

# We presume the image has already been built as :latest
docker tag ${DOCKER_REPO_NAME}:latest ${DOCKER_REPO_NAME}:${DOCKER_TAG}

if [ -z "$DOCKER_USERNAME" ]; then
  echo "${DOCKER_PASSWORD}" | docker login -u "${DOCKER_USERNAME}" --password-stdin
fi
	
# If this is a tag, push a tag.  Otherwise, push to latest
GIT_TAG=`git describe --tags 2>/dev/null`
if [ -n "$GIT_TAG" ]; then
    DOCKER_TAG=${DATE}-${GIT_TAG}
    docker push ${DOCKER_REPO_NAME}:${DOCKER_TAG}
else 
    docker push ${DOCKER_REPO_NAME}:latest
fi

