ENV_FILE ?= .env

fmt:
	gofmt -w ./...

lint:
	golangci-lint run ./...

test:
	go test ./...

run:
	@set -a; \
		[ -f $(ENV_FILE) ] && . $(ENV_FILE); \
		set +a; \
		go run ./cmd/server
