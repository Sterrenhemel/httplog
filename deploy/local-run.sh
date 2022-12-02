#!/bin/bash

IMAGE=ghcr.io/sterrenhemel/httplog
IMAGE_TAG=local

CURDIR=$(pwd)

if [ -d "deploy" ]; then
  # 在root目录下
  echo "not in deploy dir"
  echo "assume in root dir"
  RUNTIME_ROOT=${CURDIR}
else
  # 在build目录下
  echo "in deploy/ dir ..."
  RUNTIME_ROOT=${CURDIR}/..
fi

DEPLOY_DIR=${RUNTIME_ROOT}/deploy

# local build main
build/build.sh
# local build image
docker build -f ${DEPLOY_DIR}/docker/local/Dockerfile -t $IMAGE:$IMAGE_TAG .
#docker push
docker run $IMAGE:$IMAGE_TAG