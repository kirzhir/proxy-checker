.PHONY: all cli server

all: cli server

cli:
	go build -o ./bin/cli ./cmd/cli

server:
	go build -o ./bin/server ./cmd/server

build:
	go build -o ./bin/cli ./cmd/cli && go build -o ./bin/server ./cmd/server