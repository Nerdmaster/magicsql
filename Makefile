all:
	go build

test:
	go test

format:
	find . -name "*.go" | xargs gofmt -l -w -s
