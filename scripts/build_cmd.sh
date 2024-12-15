#!/bin/bash

GOARCH=amd64 go build -ldflags "-s -w" -o ./build/oas-yaml-combine ./cmd/oas-yaml-combine/main.go
GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o ./build/oas-yaml-combine.exe ./cmd/oas-yaml-combine/main.go