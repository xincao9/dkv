FROM centos

MAINTAINER xincao9@gmail.com

ADD dkv /dkv

ENTRYPOINT [ "/dkv" ]