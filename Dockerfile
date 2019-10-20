FROM debian:jessie

LABEL maintainer="xincao9@gmail.com"

ADD dkv /dkv

ENTRYPOINT [ "/dkv" ]