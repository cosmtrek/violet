#!/usr/bin/env bash

export violet=/go/src/github.com/cosmtrek/violet

/go/bin/violet &
/go/bin/tweets-search -d=./example/tweets-search/dist