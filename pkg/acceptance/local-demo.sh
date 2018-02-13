#!/usr/bin/env bash

set -eou pipefail
#set -x  # useful for debugging

docker_cleanup() {
    echo "cleaning up existing network and containers..."
    CONTAINERS='libri|catalog'
    docker ps | grep -E ${CONTAINERS} | awk '{print $1}' | xargs -I {} docker stop {} || true
    docker ps -a | grep -E ${CONTAINERS} | awk '{print $1}' | xargs -I {} docker rm {} || true
    docker network list | grep ${CONTAINERS} | awk '{print $2}' | xargs -I {} docker network rm {} || true
}

# optional settings (generally defaults should be fine, but sometimes useful for debugging)
CATALOG_LOG_LEVEL="${CATALOG_LOG_LEVEL:-INFO}"  # or DEBUG
CATALOG_TIMEOUT="${CATALOG_TIMEOUT:-5}"  # 10, or 20 for really sketchy network

# local and filesystem constants
LOCAL_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# container command constants
CATALOG_IMAGE="gcr.io/elxir-core-infra/catalog:snapshot" # develop

echo
echo "cleaning up from previous runs..."
docker_cleanup

echo
echo "creating catalog docker network..."
docker network create catalog

echo
echo "starting catalog..."
port=10100
name="catalog-0"
docker run --name "${name}" --net=catalog -d -p ${port}:${port} ${CATALOG_IMAGE} \
    start \
    --logLevel "${CATALOG_LOG_LEVEL}" \
    --serverPort ${port}
catalog_addrs="${name}:${port}"
catalog_containers="${name}"

echo
echo "testing catalog health..."
docker run --rm --net=catalog ${CATALOG_IMAGE} test health \
    --catalogs "${catalog_addrs}" \
    --logLevel "${CATALOG_LOG_LEVEL}"

echo
echo "testing catalog ..."
docker run --rm --net=catalog ${CATALOG_IMAGE} test io \
    --catalogs "${catalog_addrs}" \
    --logLevel "${CATALOG_LOG_LEVEL}"

echo
echo "cleaning up..."
docker_cleanup

echo
echo "All tests passed."
