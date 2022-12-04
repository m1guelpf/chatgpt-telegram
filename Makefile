GIT_COMMIT=$(shell git describe --always)

.PHONY: all build clean test

all: build
default: build

build:
	go build

clean:
	rm chatgpt-telegram
