TEMP_DIR = ./.tmp

# Command templates #################################
LINT_CMD = $(TEMP_DIR)/golangci-lint run --tests=false --config .golangci.yaml
GOIMPORTS_CMD := $(TEMP_DIR)/gosimports -local github.com/anchore

# Tool versions #################################
GOLANGCILINT_VERSION := v1.63.4
GOSIMPORTS_VERSION := v0.3.8
BOUNCER_VERSION := v0.4.0
CHRONICLE_VERSION := v0.8.0

# Formatting variables #################################
BOLD := $(shell tput -T linux bold)
PURPLE := $(shell tput -T linux setaf 5)
GREEN := $(shell tput -T linux setaf 2)
CYAN := $(shell tput -T linux setaf 6)
RED := $(shell tput -T linux setaf 1)
RESET := $(shell tput -T linux sgr0)
TITLE := $(BOLD)$(PURPLE)
SUCCESS := $(BOLD)$(GREEN)

# Test variables #################################
COVERAGE_THRESHOLD := 55  # the quality gate lower threshold for unit test total % coverage (by function statements)

## Build variables #################################
SNAPSHOT_DIR := ./snapshot
VERSION := $(shell git describe --dirty --always --tags)

ifndef VERSION
	$(error VERSION is not set)
endif

ifndef TEMP_DIR
    $(error TEMP_DIR is not set)
endif

define title
    @printf '$(TITLE)$(1)$(RESET)\n'
endef

define safe_rm_rf
	bash -c 'test -z "$(1)" && false || rm -rf $(1)'
endef

define safe_rm_rf_children
	bash -c 'test -z "$(1)" && false || rm -rf $(1)/*'
endef

.PHONY: all
all: static-analysis test ## Run all linux-based checks (linting, license check, unit, integration, and linux compare tests)
	@printf '$(SUCCESS)All checks pass!$(RESET)\n'

.PHONY: static-analysis
static-analysis: check-go-mod-tidy lint check-licenses ## Run all static analysis checks

.PHONY: test
test: unit  ## Run all tests (currently unit and integrations)


## Bootstrapping targets #################################

.PHONY: ci-bootstrap
ci-bootstrap: bootstrap
	curl -sLO https://github.com/sylabs/singularity/releases/download/v3.10.0/singularity-ce_3.10.0-focal_amd64.deb && sudo apt-get install -y -f ./singularity-ce_3.10.0-focal_amd64.deb

.PHONY: bootstrap
bootstrap: $(TEMP_DIR) bootstrap-go bootstrap-tools ## Download and install all tooling dependencies (+ prep tooling in the ./tmp dir)
	$(call title,Bootstrapping dependencies)

.PHONY: bootstrap-tools
bootstrap-tools: $(TEMP_DIR)
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(TEMP_DIR)/ $(GOLANGCILINT_VERSION)
	curl -sSfL https://raw.githubusercontent.com/wagoodman/go-bouncer/master/bouncer.sh | sh -s -- -b $(TEMP_DIR)/ $(BOUNCER_VERSION)
	curl -sSfL https://raw.githubusercontent.com/anchore/chronicle/main/install.sh | sh -s -- -b $(TEMP_DIR)/ $(CHRONICLE_VERSION)
	# the only difference between goimports and gosimports is that gosimports removes extra whitespace between import blocks (see https://github.com/golang/go/issues/20818)
	GOBIN="$(realpath $(TEMP_DIR))" go install github.com/rinchsan/gosimports/cmd/gosimports@$(GOSIMPORTS_VERSION)

.PHONY: bootstrap-go
bootstrap-go:
	go mod download

$(TEMP_DIR):
	mkdir -p $(TEMP_DIR)

## Static analysis targets #################################

.PHONY: static-analysis
static-analysis: check-licenses lint

.PHONY: lint
lint: ## Run gofmt + golangci lint checks
	$(call title,Running linters)
	# ensure there are no go fmt differences
	@printf "files with gofmt issues: [$(shell gofmt -l -s .)]\n"
	@test -z "$(shell gofmt -l -s .)"

	# run all golangci-lint rules
	$(LINT_CMD)
	@[ -z "$(shell $(GOIMPORTS_CMD) -d .)" ] || (echo "goimports needs to be fixed" && false)

	# go tooling does not play well with certain filename characters, ensure the common cases don't result in future "go get" failures
	$(eval MALFORMED_FILENAMES := $(shell find . | grep -v tar-cache | grep -e ':'))
	@bash -c "[[ '$(MALFORMED_FILENAMES)' == '' ]] || (printf '\nfound unsupported filename characters:\n$(MALFORMED_FILENAMES)\n\n' && false)"


.PHONY: format
format: ## Auto-format all source code
	$(call title,Running formatters)
	gofmt -w -s .
	$(GOIMPORTS_CMD) -w .
	go mod tidy

.PHONY: lint-fix
lint-fix: format  ## Auto-format all source code + run golangci lint fixers
	$(call title,Running lint fixers)
	$(LINT_CMD) --fix

.PHONY: check-licenses
check-licenses:
	$(call title,Validating licenses for go dependencies)
	$(TEMP_DIR)/bouncer check

check-go-mod-tidy:
	@ .github/scripts/go-mod-tidy-check.sh && echo "go.mod and go.sum are tidy!"

## Testing targets #################################

.PHONY: unit
unit: $(TEMP_DIR) ## Run unit tests (with coverage)
	$(call title,Running unit tests)
	go test -race -coverprofile $(TEMP_DIR)/unit-coverage-details.txt ./...
	@.github/scripts/coverage.py $(COVERAGE_THRESHOLD) $(TEMP_DIR)/unit-coverage-details.txt

## Halp! #################################

.PHONY: help
help:  ## Display this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(BOLD)$(CYAN)%-25s$(RESET)%s\n", $$1, $$2}'
