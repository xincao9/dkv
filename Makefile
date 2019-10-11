GOPATH:=$(shell go env GOPATH)

build:
	go build -tags=jsoniter -o dkv main.go

test:
	mkdir -p  $(HOME)/.dkv
	cp config.yaml $(HOME)/.dkv
	go test -v ./... -cover

docker:
	docker build . -t dkv:latest
