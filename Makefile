GOPATH:=$(shell go env GOPATH)

test:
	mkdir -p  $(HOME)/.dkv
	cp config.yaml $(HOME)/.dkv
	go test -v ./... -cover
	rm $(HOME)/.dkv/config.yaml

build:
	go build -tags=jsoniter -o dkv main.go

docker:build
	docker build . -t dkv:latest

install:build
	mkdir -p /usr/local/dkv/data
	chmod 777 /usr/local/dkv/data
	mkdir -p /usr/local/dkv/data/m
	chmod 777 /usr/local/dkv/data/m
	mkdir -p /usr/local/dkv/data/s
	chmod 777 /usr/local/dkv/data/s
	mv dkv /usr/local/dkv/
	cp config-prod.yaml /usr/local/dkv/config.yaml
	cp config-m.yaml /usr/local/dkv/config-m.yaml
	cp config-s.yaml /usr/local/dkv/config-s.yaml
	ln -s /usr/local/dkv/dkv /usr/local/bin/dkv
