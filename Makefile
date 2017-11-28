.PHONY: all test format lint

all:
	go build

test:
	go test

format:
	find . -name "*.go" | xargs gofmt -l -w -s

lint:
	golint ./...
	go vet ./...
