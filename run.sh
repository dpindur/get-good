#!/bin/bash

docker run -it --rm -v "$(pwd):/app/get-good" -w /app/get-good golang:1.11beta3 bash