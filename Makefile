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

.PHONY: build-docker
build-docker:
	docker build -t docker.m6r.eu/goosball .

.PHONY: run-docker
run-docker: build-docker goosball.db
	docker run --env-file .env -v "${PWD}/goosball.db:/app/goosball.db" -p "127.0.0.1:1337:8080" docker.m6r.eu/goosball
