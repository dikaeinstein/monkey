lint:
	golangci-lint run

test:
	GO111MODULE=on go test -race ./...

build:
	GO111MODULE=on go build -o monkey cmd/main.go
