#!/bin/bash

set -e -x -u

go fmt ./cmd/...

go build -o controller ./cmd/controller/...
ls -la ./controller

echo SUCCESS
