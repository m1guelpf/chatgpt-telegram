GIT_COMMIT=$(shell git describe --always)

all: build
default: build

build:
	go build

clean:
	rm chatgpt-telegram
