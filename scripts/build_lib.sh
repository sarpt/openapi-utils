#!/bin/bash

go build -buildmode=c-shared -o ./build/liboasyamlcombine.so ./lib/oas-yaml-combine/main.go