.PHONY: build test test-integration lint clean \
       portal-install portal-dev portal-build portal-lint \
       orchestrator sync-templates sync-jobs kind-setup orchestrator-dev

BINARY := bin/integrations

build:
	go build -o $(BINARY) ./cmd/integrations

orchestrator:
	go build -o bin/orchestrator ./cmd/orchestrator

sync-templates:
	go build -o bin/sync-templates ./cmd/sync-templates
	doppler run -- ./bin/sync-templates agents

sync-jobs: sync-templates

test:
	go test -race -coverprofile=coverage.out ./internal/...
	go tool cover -func=coverage.out

test-integration:
	doppler run -- go test -race -tags=integration -v ./internal/...

lint:
	go vet ./...

clean:
	rm -rf bin/ coverage.out

# Orchestrator dev targets
kind-setup:
	bash scripts/kind-setup.sh

orchestrator-dev:
	doppler run -- go run ./cmd/orchestrator

# Docker images
docker-agent-base:
	GOOS=linux GOARCH=arm64 go build -o docker/agent-base/bin/integrations ./cmd/integrations
	docker build -t ghcr.io/emdash-projects/agent-base:dev -f docker/agent-base/Dockerfile .

docker-export-creds:
	go build -o docker/export-creds/bin/export-creds ./cmd/export-creds
	docker build -t ghcr.io/emdash-projects/export-creds:latest -f docker/export-creds/Dockerfile .

# Portal targets
portal-install:
	cd portal && npm install

portal-dev:
	cd portal && doppler run -p agents -c dev -- npm run dev

portal-build:
	cd portal && doppler run -p agents -c dev -- npm run build

portal-lint:
	cd portal && npm run lint
