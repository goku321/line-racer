#!/bin/bash

trap 'kill %1; kill %2' SIGINT

set -m

go build

go run main.go -nodeType racer -port 30691 &
go run main.go -nodeType racer -port 37002 &
go run main.go