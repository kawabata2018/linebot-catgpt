.PHONY: lint lint-fix

## Linting
lint: ## Run golangci-lint
	@echo "Running golangci-lint..."
	golangci-lint run --verbose ./cloudrun

lint-fix: ## Run golangci-lint with auto-fix
	@echo "Running golangci-lint with auto-fix..."
	golangci-lint run --fix --verbose ./cloudrun
