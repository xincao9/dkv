FROM golang:1.12

LABEL maintainer="xincao9@gmail.com"

WORKDIR /go/src/dkv
COPY . .
RUN make install

EXPOSE 9090
EXPOSE 6380

CMD ["/usr/local/dkv/bin/dkv", "-conf=config-prod.yaml"]
