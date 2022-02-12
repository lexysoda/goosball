include .env
export

.PHONY: run
run: build goosball.db
	go run .

.PHONY: build
build:
	go build .

goosball.db:
	sqlite3 goosball.db < sql/init.sql
