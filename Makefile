.PHONY: build test lint clean

BINARY := bin/integrations

build:
	go build -o $(BINARY) ./cmd/integrations

test:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

lint:
	go vet ./...

clean:
	rm -rf bin/ coverage.out
