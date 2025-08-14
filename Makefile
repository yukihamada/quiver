.PHONY: all build build-all build-provider build-gateway build-aggregator clean clean-all test test-all test-unit test-integration test-e2e lint bench coverage help run stop demo deploy-local deploy-testnet deps

# Default target
all: build-all

# Alias for compatibility
build: build-all

# Build all services
build-all: build-provider build-gateway build-aggregator

build-provider:
	@echo "Building provider..."
	@cd provider && $(MAKE) build

build-gateway:
	@echo "Building gateway..."
	@cd gateway && $(MAKE) build

build-aggregator:
	@echo "Building aggregator..."
	@cd aggregator && $(MAKE) build

# Clean build artifacts
clean: clean-all

clean-all:
	@echo "Cleaning all services..."
	@cd provider && $(MAKE) clean
	@cd gateway && $(MAKE) clean
	@cd aggregator && $(MAKE) clean
	@rm -rf bootstrap/bootstrap bootstrap/*.log
	@rm -rf bin/ test-results/ coverage/
	@echo "Clean complete!"

# Test all services
test: test-all

test-all:
	@echo "Testing all services..."
	@echo "===================="
	@cd provider && $(MAKE) test
	@cd gateway && $(MAKE) test
	@cd aggregator && $(MAKE) test
	@echo "===================="
	@echo "All tests passed!"

test-unit:
	mkdir -p coverage
	cd provider && go test -v -coverprofile=../coverage/provider.out ./...
	cd gateway && go test -v -coverprofile=../coverage/gateway.out ./...
	cd aggregator && go test -v -coverprofile=../coverage/aggregator.out ./...

test-integration:
	@echo "Integration tests not yet implemented"

test-e2e:
	@echo "E2E tests not yet implemented"

lint:
	cd provider && go vet ./... && go fmt ./...
	cd gateway && go vet ./... && go fmt ./...
	cd aggregator && go vet ./... && go fmt ./...

bench:
	@echo "Benchmarks not yet implemented"

coverage:
	mkdir -p coverage
	$(MAKE) test-unit
	go tool cover -html=coverage/provider.out -o coverage/provider.html
	go tool cover -html=coverage/gateway.out -o coverage/gateway.html
	go tool cover -html=coverage/aggregator.out -o coverage/aggregator.html
	@echo "Coverage reports generated in coverage/"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@cd provider && go mod download
	@cd gateway && go mod download
	@cd aggregator && go mod download
	@cd contracts && npm install
	@echo "Dependencies installed!"

# Run network
run: build-all
	@echo "Starting QUIVer network..."
	@./scripts/start-network.sh

# Stop network
stop:
	@echo "Stopping QUIVer network..."
	@./scripts/stop-network.sh

# Run demo
demo: build-all
	@./scripts/demo.sh

# Deploy contracts to local hardhat network
deploy-local:
	@cd contracts && npm run deploy:local

# Deploy contracts to testnet
deploy-testnet:
	@echo "Deploying to Polygon Amoy testnet..."
	@cd contracts && npm run deploy:amoy

# Show help
help:
	@echo "QUIVer Makefile Commands:"
	@echo "========================"
	@echo "  make build-all      - Build all services"
	@echo "  make test-all       - Run all tests"
	@echo "  make clean-all      - Clean all build artifacts"
	@echo "  make deps           - Install all dependencies"
	@echo "  make run            - Build and start the network"
	@echo "  make stop           - Stop the network"
	@echo "  make demo           - Run the demo"
	@echo "  make deploy-local   - Deploy contracts locally"
	@echo "  make deploy-testnet - Deploy to testnet"
	@echo "  make lint           - Run linters"
	@echo "  make coverage       - Generate coverage reports"
	@echo "  make help           - Show this help message"