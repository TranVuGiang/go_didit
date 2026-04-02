.PHONY: lint test tidy

tidy:
	go mod tidy

lint:
	golangci-lint run ./...

test:
	go test -v -race ./...
