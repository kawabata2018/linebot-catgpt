.PHONY: lint lint-fix

## Linting
lint: ## Run golangci-lint
	@echo "Running golangci-lint..."
	cd cloudrun && golangci-lint run

lint-fix: ## Run golangci-lint with auto-fix
	@echo "Running golangci-lint with auto-fix..."
	cd cloudrun && golangci-lint run --fix
