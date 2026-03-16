#!/bin/bash
set -ex

rm -f /tmp/sqlite.db

env \
GTS_DB_TYPE=sqlite \
GTS_DB_ADDRESS=/tmp/sqlite.db \
GTS_LOG_LEVEL=trace \
GTS_PORT=${GTS_DB_PORT:-${RANDOM}} \
go run -v -tags='debug nopostgres' ./cmd/gotosocial testrig start