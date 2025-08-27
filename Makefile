.PHONY: test coverage race fmt lint staticcheck gosec benchmark clean help

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

test: ## Run tests
	go test -v ./...

coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

race: ## Run tests with race detector
	CGO_ENABLED=1 go test -race ./...

fmt: ## Check code formatting
	@if [ "$$(gofmt -s -l . | wc -l)" -gt 0 ]; then \
		echo "Code not formatted:"; \
		gofmt -s -l .; \
		exit 1; \
	fi
	@echo "Code is properly formatted"

lint: ## Run golint
	@golint_output=$$(~/go/bin/golint ./...); \
	if [ -n "$$golint_output" ]; then \
		echo "Linting issues found:"; \
		echo "$$golint_output"; \
		exit 1; \
	fi
	@echo "No linting issues found"

staticcheck: ## Run staticcheck
	~/go/bin/staticcheck ./...

gosec: ## Run security scan (core library only)
	~/go/bin/gosec --exclude-dir=demo ./...

quality: fmt lint staticcheck gosec ## Run all quality checks

ci: test race quality coverage ## Run all CI checks

clean: ## Clean generated files
	rm -f coverage.out coverage.html

install-tools: ## Install development tools
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install golang.org/x/lint/golint@latest
	go install github.com/securego/gosec/v2/cmd/gosec@latest
