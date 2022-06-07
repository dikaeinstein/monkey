.PHONY: lint
lint:
	go run github.com/golangci/golangci-lint/cmd/golangci-lint run

.PHONY: test
test:
	go test -race ./...

.PHONY: build
build:
	go build -o monkey cmd/main.go

.PHONY: clean
clean:
	rm ./monkey

.PHONY: all test clean
