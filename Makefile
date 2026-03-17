.PHONY: build test test-integration lint clean portal-install portal-dev portal-build portal-lint

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

# Portal targets
portal-install:
	cd portal && npm install

portal-dev:
	cd portal && npm run dev

portal-build:
	cd portal && npm run build

portal-lint:
	cd portal && npm run lint
