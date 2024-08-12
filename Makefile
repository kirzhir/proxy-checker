.PHONY: all cli server

all: cli server

build:
	go build -o ./bin/pc