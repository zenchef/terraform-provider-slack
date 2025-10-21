TEST?=$$(go list ./... |grep -v 'vendor')
SWEEP_DIR?=./slack
SWEEP?=eu-west-1
GOFMT_FILES?=$$(find . -name '*.go' |grep -v vendor)

default: testacc bin

bin:
	go install
	scripts/install_plugin.sh

download:
	@echo "==> Download go.mod dependencies"
	@go mod download

lint: vet docs-lint
	@echo "==> Checking source code against linters..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run -v; \
	else \
		echo "golangci-lint not found, skipping (runs separately in CI)"; \
	fi

sweep:
	@echo "WARNING: This will destroy infrastructure. Use only in development accounts."
	go test $(SWEEP_DIR) -v -sweep=$(SWEEP) $(SWEEPARGS) -timeout 60m

test: vet
	go test -v $(TEST)

testacc: vet
	@echo "==> Running acceptance tests..."
	TF_ACC=1 go test -v $(TEST)

vet:
	@echo "go vet ."
	@go vet $$(go list ./... | grep -v vendor/)

depscheck:
	@echo "==> Checking source code with go mod tidy..."
	@go mod tidy
	@git diff --exit-code -- go.mod go.sum || \
		(echo; echo "Unexpected difference in go.mod/go.sum files. Run 'go mod tidy' command or revert any go.mod/go.sum changes and commit."; exit 1)

docs: docs-lint

docs-lint:
	@echo "==> Checking docs against linters..."
	@docker run -v $(PWD):/markdown 06kellyjac/markdownlint-cli docs/ || (echo; \
		echo "Unexpected issues found in docs Markdown files."; \
		echo "To apply any automatic fixes, run 'make docs-lint-fix' and commit the changes."; \
		exit 1)

docs-lint-fix:
	@echo "==> Applying automatic docs linter fixes..."
	@docker run -v $(PWD):/markdown 06kellyjac/markdownlint-cli --fix docs/

.PHONY: build download lint sweep test testacc vet depscheck docs docs-lint docs-lint-fix
