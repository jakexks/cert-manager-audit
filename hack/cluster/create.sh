#!/usr/bin/env bash
set -e
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"

DOCKER_CMD=${DOCKER_CMD:-$(which docker)}
if [[ ! -x "$DOCKER_CMD" ]]; then
  echo "Please install Docker or set DOCKER_CMD to the path to your docker client"
  exit 1
fi

KIND_CMD=${KIND_CMD:-$(which kind)}
if [[ ! -x "$KIND_CMD" ]]; then
  echo "Please install kind or set KIND_CMD to the path to your docker client"
  exit 1
fi

# "$DOCKER_CMD" build -f "$SCRIPT_DIR/Dockerfile" -t node-with-audit-config "$SCRIPT_DIR"

"$KIND_CMD" delete cluster --name audit-test

sed "s#__AUDIT_CONFIG_PATH__#$SCRIPT_DIR#g" hack/cluster/kind-cluster.yaml > "$SCRIPT_DIR/cluster.yaml"

"$KIND_CMD" create cluster --config "$SCRIPT_DIR/cluster.yaml"
