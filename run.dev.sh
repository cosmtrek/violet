#!/usr/bin/env bash

docker build --build-arg http_proxy=http://172.28.18.219:8964 --build-arg https_proxy=http://172.28.18.219:8964 \
    -t cosmtrek/violet .

docker run --name "violet" --rm -p 6060:6060 -p 8100:8100 cosmtrek/violet