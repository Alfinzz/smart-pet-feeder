#!/bin/sh
set -eu

if [ "${RUN_MIGRATE:-false}" = "true" ] || [ "${RUN_MIGRATE:-false}" = "1" ]; then
  ./migrate up
fi

exec "$@"
