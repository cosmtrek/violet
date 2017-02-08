#!/usr/bin/env bash

git pull origin master
docker build -t cosmtrek/violet .
docker stop violet
docker rm violet
docker run --name "violet" -d -p 6060:6060 -p 8100:8100 cosmtrek/violet
docker logs violet >> $(pwd)/logs/production.log