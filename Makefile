GOPATH:=$(shell go env GOPATH)

build:
	rm dkv
	go build -tags=jsoniter -o dkv main.go

test:
	mkdir -p  $(HOME)/.dkv
	cp config.yaml $(HOME)/.dkv
	go test -v ./... -cover
	rm $(HOME)/.dkv/config.yaml

docker:
	docker build . -t dkv:latest
