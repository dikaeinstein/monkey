lint:
	golangci-lint run

test:
	go test -race ./...

build:
	go build -o monkey cmd/main.go

clean:
	rm ./monkey

.PHONY: all test clean
