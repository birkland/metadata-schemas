#!/bin/sh 

source $(dirname $0)/.definitions.sh

docker build . -t ${DOCKER_REPO_NAME}:latest
