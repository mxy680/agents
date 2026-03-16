.PHONY: build test test-integration lint clean

BINARY := bin/integrations

build:
	go build -o $(BINARY) ./cmd/integrations

test:
	go test -race -coverprofile=coverage.out ./internal/...
	go tool cover -func=coverage.out

test-integration:
	doppler run -- go test -race -tags=integration -v ./internal/...

lint:
	go vet ./...

clean:
	rm -rf bin/ coverage.out
