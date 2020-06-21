GOPATH:=$(shell go env GOPATH)

test:
	go test -v ./... -cover

build:
	go build -tags=jsoniter -o dkv main.go

docker:build
	docker build . -t dkv:latest

install:build
	mkdir -p /usr/local/dkv/conf
	mkdir -p /usr/local/dkv/bin
	cp ./resource/conf/* /usr/local/dkv/conf
	cp dkv /usr/local/dkv/bin
	export $PATH=$PATH:/usr/local/dkv/bin