GOPATH:=$(shell go env GOPATH)

test:
	mkdir -p  $(HOME)/.dkv
	cp resource/conf/config.yaml $(HOME)/.dkv
	go test -v ./... -cover
	rm $(HOME)/.dkv/config.yaml

build:
	go build -tags=jsoniter -o dkv main.go

docker:build
	docker build . -t dkv:latest

install:
	go build -tags=jsoniter -o dkv main.go
	mkdir -p /usr/local/dkv/conf
	mkdir -p /usr/local/dkv/bin
	cp ./resource/conf/* /usr/local/dkv/conf
	cp dkv /usr/local/dkv/bin
