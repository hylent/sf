# Makefile

BUILD_PWD=$(shell pwd)

local:
	go build -o bin/sf

proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative *.proto
	protoc-go-inject-tag -input="*.pb.go"
