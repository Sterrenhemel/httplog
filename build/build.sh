#!/bin/bash

# 通用
#
RUN_NAME="main"
CURDIR=$(pwd)

if [ -d "build" ]; then
  # 在root目录下
  echo "not in build dir"
  echo "assume in root dir"
  RUNTIME_ROOT=${CURDIR}
else
  # 在build目录下
  echo "in build/ dir ..."
  RUNTIME_ROOT=${CURDIR}/..
fi

OUTPUT_DIR=${RUNTIME_ROOT}/deploy/output
SRC_DIR=${RUNTIME_ROOT}
CONFIG_DIR=${RUNTIME_ROOT}/configs
TARGET_OS="linux"
TARGET_ARCH="amd64"
BUILD_TIME=`date -u '+%Y-%m-%d %H:%M:%S'`
# 自定义
# build 哪一个
TARGET_LIST=(
  http
)
SRC_LIST=(
  .                        # http server
)
OUTPUT_LIST=(
    $OUTPUT_DIR/bin/${RUN_NAME}
)

if [ "$#" -eq 0 ]; then
    # default value
    paramsArray=(http)
else
    # value
    paramsArray=( "$@" )
#    paramsArray+=(http)
fi

# 
mkdir -p $OUTPUT_DIR/bin

SRC_BUILD_DIR=()
TARGET_BUILD_LIST=()
OUTPUT_BUILD_LIST=()

for i in "${!TARGET_LIST[@]}"; do
  T="${TARGET_LIST[i]}"
  S="${SRC_LIST[i]}"
  O="${OUTPUT_LIST[i]}"
  if [[ ${paramsArray[@]} =~ ${T} ]]
  then
    TARGET_BUILD_LIST+=( $T )
    SRC_BUILD_DIR+=( $S )
    OUTPUT_BUILD_LIST+=( $O )
  fi
done

LEN=`expr ${#SRC_BUILD_DIR[@]} - 1`
for i in $(seq 0 $LEN);
do
  SRC_MAIN_DIR="${SRC_BUILD_DIR[$i]}"
  OUTPUT_MAIN="${OUTPUT_BUILD_LIST[$i]}"
#  echo $SRC_MAIN_DIR
#  echo $OUTPUT_MAIN
  echo "building $SRC_MAIN_DIR to $OUTPUT_MAIN"
  cd ${SRC_DIR} && env GOOS=${TARGET_OS} GOARCH=${TARGET_ARCH} CGO_ENABLED=0 go build -trimpath -ldflags="-X 'main.BuildTime=${BUILD_TIME}'" -o ${OUTPUT_MAIN} ${SRC_MAIN_DIR}
  chmod 777 ${OUTPUT_MAIN}
done

