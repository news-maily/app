#!/usr/bin/env bash

set -euxo pipefail

export $(egrep -v '^#' .env.local | xargs)
trap 'kill 0' SIGINT; go run mailbadger.go & \
  go run consumers/campaigner/main.go & \
  go run consumers/sender/main.go
