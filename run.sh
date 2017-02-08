#!/usr/bin/env bash

docker build -t cosmtrek/violet .
docker run --name "violet" --rm -p 6060:6060 -p 8100:8100 cosmtrek/violet