FROM golang:1.11.1-stretch
MAINTAINER Masataka Kuwabara <kuwabara@pocke.me>

RUN apt-get update && apt-get upgrade -y && \
    apt-get install -y \
    tmux \
    git \
    locales \
    gcc && \
    apt-get clean && \
    rm -rf /var/lib/apt/lists/*

RUN mkdir -p /go/src/github.com/pocke/ptmux
RUN localedef -i en_US -c -f UTF-8 -A /usr/share/locale/locale.alias en_US.UTF-8

ENV GO111MODULE=on \
    LANG=en_US.utf8

WORKDIR /go/src/github.com/pocke/ptmux
CMD go test -v
