.PHONY: build run dev

build:
	go build -o bin/kineticgo ./cmd

run: build
	./bin/kineticgo

dev:
	go run ./cmd
