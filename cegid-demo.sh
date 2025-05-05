#!/bin/bash

cd plugins/cegid
go build -buildmode=plugin -o plugin.so plugin.go

cd ../autoload_cegid
go build -buildmode=plugin -o plugin.so plugin.go

cd ../..
go run main.go