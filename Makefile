#!/usr/bin/make -f

PWD=$(shell pwd)

build:
	go build -o rfplayer-bin ./cmd/rfplayer/...

install:
	go install ./cmd/rfplayer/...