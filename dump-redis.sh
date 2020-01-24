#!/bin/bash
# Usage: ./dump-redis.sh LB:YELLS > loudbot-yells.txt

if [ -z $(which redis-cli) ]; then
  echo "Must have redis-cli installed"
  exit 1
fi

if [ -z "$1" ]; then
  echo "Must provide redis set/key"
  exit 1
fi
R_KEY="$1"

if [ -z "$2" ]; then
  R_HOST="127.0.0.1"
else
  R_HOST="$2"
fi

redis-cli -h $R_HOST sscan $R_KEY 0 count 999999 | tail +1
