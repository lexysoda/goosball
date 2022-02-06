include .env
export

.PHONY: run
run: build
	go run .

.PHONY: build
build:
	go build .
