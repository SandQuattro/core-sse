include .env

PROJECTNAME=sse-demo_core

MAC_ARCH=arm64
ARCH=amd64

VERSION=$(shell git describe --tags --always --long --dirty)

WINDOWS=$(PROJECTNAME)_windows_$(ARCH).exe
LINUX=$(PROJECTNAME)_linux_$(ARCH)
DARWIN=$(PROJECTNAME)_darwin_$(ARCH)

# Go переменные.
GOBASE=$(shell pwd)
GOPATH=$(shell go env GOPATH)
GOBIN=$(GOBASE)/bin
GOFILES=$(wildcard *.go)

# Перенаправление вывода ошибок в файл, чтобы мы показывать его в режиме разработки.
STDERR=./tmp/.$(PROJECTNAME)-stderr.txt

# PID-файл будет хранить идентификатор процесса, когда он работает в режиме разработки
PID=./tmp/.$(PROJECTNAME)-api-server.pid

# Make пишет работу в консоль Linux. Сделаем его silent.
# MAKEFLAGS += --silent

.PHONY: all test clean migrations

all: test build

test:
	@echo "Running tests..."
	@go test ./...

build: windows linux darwin
	@echo version: $(VERSION)

windows: $(WINDOWS)

linux: $(LINUX)

darwin: $(DARWIN)

$(WINDOWS):
	@echo "Building windows app..."
	@env GOOS=windows GOARCH=$(ARCH) go build -v -o bin/$(WINDOWS) -ldflags="-s -w -X main.version=$(VERSION)" ./cmd/core/main.go

$(LINUX):
	@echo "Building linux app..."
	@env GOOS=linux GOARCH=$(ARCH) go build -v -o bin/$(LINUX) -ldflags="-s -w -X main.version=$(VERSION)" ./cmd/core/main.go

$(DARWIN):
	@echo "Building macos app..."
	@env GOOS=darwin GOARCH=$(MAC_ARCH) go build -v -o bin/$(DARWIN) -ldflags="-s -w -X main.version=$(VERSION)" ./cmd/core/main.go

clean:
	@echo "Cleaning up..."
	@go clean
	@rm -f ./bin/$(WINDOWS) ./bin/$(LINUX) ./bin/$(DARWIN)

migrations:
	@go run ./cmd/migrator -config=conf/application.conf -migrations-path=./migrations

run:
	@echo "Running instances..."
	@nohup ./bin/$(LINUX) -config=conf/application.conf -port=9002 &

rebuild:
	@echo "Rebuilding app..."
	@kill -2 `cat RUNNING_PID`
	@rm nohup.out
	@git pull origin main
	@make linux
	@nohup ./bin/$(LINUX) -config=conf/application.conf -port=9002 &

restart:
	@echo "Restart instance..."
	@kill -2 `cat RUNNING_PID`
	@rm nohup.out
	@nohup ./bin/$(LINUX) -config=conf/application.conf -port=9002 &