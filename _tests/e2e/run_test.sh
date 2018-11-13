#!/usr/bin/env bash

set -eu
set -o pipefail

cd "$(dirname $0)/../.."

DIR=_tests/e2e

getImageName() {
  echo gex-e2e-test-$1:go$GO_VERSION
}

cleanup() {
  rm -rf $DIR/{bin,vendor,go.*,Gopkg.*,tools.go}
}

testDep() {
  cleanup

  name=$(getImageName dep)
  docker build -t $name --build-arg GO_VERSION=$GO_VERSION -f ./$DIR/Dockerfile .
  docker run \
    --rm \
    -v $(pwd)/$DIR:/go/src/myapp \
    --env SNAPSHOT_DIR=.snapshots_dep \
    --env TEST_TARGET=dep \
    $name \
    sh -c 'curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh && dep init -v && go test -v'
}

testMod() {
  cleanup

  name=$(getImageName mod)
  docker build -t $name --build-arg GO_VERSION=$GO_VERSION -f ./$DIR/Dockerfile .
  docker run \
    --rm \
    -v $(pwd)/$DIR:/go/src/myapp \
    --env SNAPSHOT_DIR=.snapshots_mod \
    --env GO111MODULE=on \
    --env TEST_TARGET=mod \
    $name \
    sh -c 'go mod init && go test -v'
}


#  Entrypoint
#-----------------------------------------------
TARGET="${1:-}"
shift || true

case "$TARGET" in
  dep) testDep ;;
  mod) testMod ;;
  *)
    echo "Unknown command: $TARGET" 1>&2
    exit 1
esac
