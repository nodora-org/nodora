#!/bin/sh

go run golang.org/x/tools/cmd/goyacc@v0.42.0 -o ./parser.go ./parser.y