GOPATH:=$(shell go env GOPATH)

test:
	mkdir -p  $(HOME)/.dkv
	cp config.yaml $(HOME)/.dkv
	go test -v ./... -cover
	rm $(HOME)/.dkv/config.yaml

build:test
	go build -tags=jsoniter -o dkv main.go

docker:build
	docker build . -t dkv:latest

install:build
	mkdir -p /usr/local/dkv/data
	mv dkv /usr/local/dkv/
	cp config-prod.yaml /usr/local/dkv/config.yaml
	ln -s /usr/local/dkv/dkv /usr/local/bin/dkv