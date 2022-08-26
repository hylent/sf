# Makefile

BUILD_PWD=$(shell pwd)

local:proto
	go build -ldflags "-X github.com/hylent/sf/logger.buildPwd=${BUILD_PWD}" -o bin/sf

proto:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative demo/proto/*.proto
