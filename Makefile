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
