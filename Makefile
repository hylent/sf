# Makefile

BUILD_PWD=$(shell pwd)

local:
	go build -ldflags "-X github.com/hylent/sf/logger.buildPwd=${BUILD_PWD}" -o bin/sf
