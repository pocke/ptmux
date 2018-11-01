FROM golang:1.11.1-alpine3.7
MAINTAINER Masataka Kuwabara <kuwabara@pocke.me>

RUN apk --update add tmux bash git && \
  rm -rf /var/cache/apk/* && \
  mkdir -p /go/src/github.com/pocke/ptmux && \
  echo 'set -g base-index 1' > ~/.tmux.conf

WORKDIR /go/src/github.com/pocke/ptmux
CMD go test -v
