GO ?= go

WORK_DIR   := $(shell pwd)

# Test configuration
FORGEJO_SDK_TEST_URL ?= http://localhost:3000
FORGEJO_SDK_TEST_USERNAME ?= test01
FORGEJO_SDK_TEST_PASSWORD ?= test01
FORGEJO_SDK_TEST_EMAIL ?= test01@forgejo.org

# Every top-level integration test calls t.Parallel(). Pin the degree here
# instead of letting it default to GOMAXPROCS: the CI runner has 4 CPUs and
# shares them with the Forgejo service container, so leaving it implicit
# makes the suite's behavior depend on whatever the runner exposes that day.
# Measured: 4 and 8 are within noise of each other -- the bottleneck is the
# Forgejo instance, not Go. See docs/PLANO-PARALELIZACAO-TESTES.md.
TEST_PARALLEL ?= 4

PACKAGE := codeberg.org/MatheusAlves96/forgejo-sdk/forgejo/v3

GOFUMPT_PACKAGE ?= mvdan.cc/gofumpt@v0.7.0
GOLANGCI_LINT_VERSION ?= v2.9.0

# renovate: datasource=docker depName=codeberg.org/forgejo/forgejo
FORGEJO_VERSION := 16.0.5
FORGEJO_DL := https://codeberg.org/forgejo/forgejo/releases/download/v$(FORGEJO_VERSION)/forgejo-$(FORGEJO_VERSION)-

# Detect OS and architecture
UNAME_S := $(shell uname -s)
UNAME_M := $(shell uname -m)
NATIVE_BINARY_AVAILABLE := no

ifeq ($(UNAME_S),Linux)
  FORGEJO_DL := $(FORGEJO_DL)linux-
  ifeq ($(UNAME_M),x86_64)
    FORGEJO_DL := $(FORGEJO_DL)amd64
    NATIVE_BINARY_AVAILABLE := yes
  else ifeq ($(UNAME_M),aarch64)
    FORGEJO_DL := $(FORGEJO_DL)arm64
    NATIVE_BINARY_AVAILABLE := yes
  else ifneq ($(filter armv6% armv7% armhf arm,$(UNAME_M)),)
    FORGEJO_DL := $(FORGEJO_DL)arm-6
    NATIVE_BINARY_AVAILABLE := yes
  endif
endif
# macOS: Native binaries not available, Docker required
# Linux: Only amd64, arm64, and arm-6 binaries available for Forgejo 9.0.3+

# Check if Docker is available
HAS_DOCKER := $(shell command -v docker 2> /dev/null)
DOCKER_AVAILABLE := $(if $(HAS_DOCKER),yes,no)

.PHONY: all
all: clean test build ## Run "make clean test build".

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk command is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: clean
clean: ## Clean build artifacts and test instances.
	rm -r -f test test-cache
	cd forgejo && $(GO) clean -i ./...

##@ Development

.PHONY: fmt
fmt: ## Format the code.
	find . -name "*.go" -type f | xargs gofmt -s -w; \
	$(GO) run $(GOFUMPT_PACKAGE) -extra -w ./forgejo

.PHONY: vet
vet: ## Examines Go source code and reports suspicious constructs.
	# Default vet
	cd forgejo && $(GO) vet $(PACKAGE)

.PHONY: ci-lint
ci-lint: ## Run the linter.
	@cd forgejo/; echo -n "gofumpt ...";\
	diff=$$($(GO) run $(GOFUMPT_PACKAGE) -extra -l .); \
	if [ -n "$$diff" ]; then \
		echo; echo "Not gofumpt-ed"; \
		exit 1; \
	fi; echo " done"; echo -n "golangci-lint ...";\
	bin=$$($(GO) env GOPATH)/bin; \
	if ! "$$bin/golangci-lint" --version 2>/dev/null | grep -q "$(GOLANGCI_LINT_VERSION:v%=%)"; then \
		curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b "$$bin" $(GOLANGCI_LINT_VERSION); \
	fi; \
	"$$bin/golangci-lint" run --timeout 5m; \
	if [ $$? -ne 0 ]; then \
		echo; echo "Doesn't pass golangci-lint"; \
		exit 1; \
	fi; echo " done"; \
	cd -; \

##@ Testing

.PHONY: test-unit
test-unit: ## Run the client's own unit tests (no Forgejo instance needed).
	cd forgejo && $(GO) test -race -run '^TestUnit_' -v ./...

.PHONY: test
test: ## Run the integration test suite (requires a running Forgejo instance).
	@export FORGEJO_SDK_TEST_URL=${FORGEJO_SDK_TEST_URL}; export FORGEJO_SDK_TEST_USERNAME=${FORGEJO_SDK_TEST_USERNAME}; export FORGEJO_SDK_TEST_PASSWORD=${FORGEJO_SDK_TEST_PASSWORD}; \
	if [ -z "$(shell curl --noproxy "*" "${FORGEJO_SDK_TEST_URL}/api/v1/version" 2> /dev/null)" ]; then \echo "No test-instance detected! See Make targets test-instance*"; exit 1; else \
	    cd forgejo && $(GO) test -race -timeout 20m -parallel $(TEST_PARALLEL) -cover -coverprofile coverage.out; \
	fi

.PHONY: check-forgejo-version
check-forgejo-version: ## Verify every pinned Forgejo test-instance version matches.
	@bash scripts/check-forgejo-version.sh

.PHONY: check-test-instance-settings
check-test-instance-settings: ## Verify the extra app.ini settings match across Makefile/CI/main_test.go.
	@bash scripts/check-test-instance-settings.sh

.PHONY: check-route-coverage
check-route-coverage: ## Verify every *Client method (route) is referenced by a test.
	@bash scripts/check-route-coverage.sh

.PHONY: check-route-coverage-strict
check-route-coverage-strict: ## Verify every route has both a unit test and an integration test.
	@bash scripts/check-route-coverage.sh --strict

.PHONY: route-report
route-report: test ## Run the suite and write the route evidence (forgejo/route-report.json).
	@echo "route evidence: forgejo/route-report.json (label: $${FORGEJO_SDK_TEST_LABEL:-local})"

# ROUTE_REPORTS lets CI pass one report per matrix leg; empty means "the local
# forgejo/route-report.json".
ROUTE_REPORTS ?=

.PHONY: route-matrix
route-matrix: ## Generate ROUTES.md and the README route-matrix block from the route reports.
	@bash scripts/gen-route-matrix.sh $(ROUTE_REPORTS)

.PHONY: check-routes
check-routes: ## Verify ROUTES.md and the README block match the route reports.
	@bash scripts/gen-route-matrix.sh --check $(ROUTE_REPORTS)

COVERAGE_MIN ?= 70

.PHONY: coverage-check
coverage-check: ## Fail if forgejo/coverage.out total coverage is below COVERAGE_MIN.
	@if [ ! -f forgejo/coverage.out ]; then echo "forgejo/coverage.out missing; run 'make test' first"; exit 1; fi
	@pct=$$(cd forgejo && $(GO) tool cover -func=coverage.out | tail -1 | grep -oE '[0-9]+\.[0-9]+'); \
	echo "total coverage: $${pct}% (minimum: $(COVERAGE_MIN)%)"; \
	awk -v p="$$pct" -v min="$(COVERAGE_MIN)" 'BEGIN { exit !(p+0 >= min+0) }' || { \
		echo "coverage $${pct}% is below the $(COVERAGE_MIN)% floor"; exit 1; \
	}

.PHONY: test-instance
test-instance: ## Start a forgejo instance for test (auto-detects method).
ifeq ($(NATIVE_BINARY_AVAILABLE),no)
	@echo "Native binary not available for $(UNAME_S)/$(UNAME_M): using Docker"
	@$(MAKE) test-instance-docker
else
ifeq ($(DOCKER_AVAILABLE),yes)
	@$(MAKE) test-instance-docker
else
	@$(MAKE) test-instance-native
endif
endif

.PHONY: test-instance-native
test-instance-native: ## Start a native forgejo instance (Linux only).
ifeq ($(NATIVE_BINARY_AVAILABLE),no)
	@echo "Error: Native binary not available for $(UNAME_S)/$(UNAME_M)"
	@echo "Please use 'make test-instance' to automatically use Docker, or install Docker and run 'make test-instance-docker'"
	@exit 1
endif
	@echo "Starting native Forgejo test instance..."
	rm -f -r ${WORK_DIR}/test 2> /dev/null; \
	mkdir -p ${WORK_DIR}/test/conf/ ${WORK_DIR}/test/data/ ${WORK_DIR}/test-cache
	[ -f ${WORK_DIR}/test-cache/forgejo-main ] || { wget ${FORGEJO_DL} -O ${WORK_DIR}/test-cache/forgejo-main; }
	cp ${WORK_DIR}/test-cache/forgejo-main ${WORK_DIR}/test/forgejo-main; \
	chmod +x ${WORK_DIR}/test/forgejo-main; \
	secret_key=$$(${WORK_DIR}/test/forgejo-main generate secret SECRET_KEY); \
	internal_token=$$(${WORK_DIR}/test/forgejo-main generate secret INTERNAL_TOKEN); \
	echo "[security]" > ${WORK_DIR}/test/conf/app.ini; \
	echo "INTERNAL_TOKEN = $$internal_token" >> ${WORK_DIR}/test/conf/app.ini; \
	echo "INSTALL_LOCK   = true" >> ${WORK_DIR}/test/conf/app.ini; \
	echo "SECRET_KEY     = $$secret_key" >> ${WORK_DIR}/test/conf/app.ini; \
	echo "PASSWORD_COMPLEXITY = off" >> ${WORK_DIR}/test/conf/app.ini; \
	echo "DISABLE_GIT_HOOKS = false" >> ${WORK_DIR}/test/conf/app.ini; \
	echo "[database]" >> ${WORK_DIR}/test/conf/app.ini; \
	echo "DB_TYPE = sqlite3" >> ${WORK_DIR}/test/conf/app.ini; \
	echo "[repository]" >> ${WORK_DIR}/test/conf/app.ini; \
	echo "ROOT = ${WORK_DIR}/test/data/" >> ${WORK_DIR}/test/conf/app.ini; \
	echo "ENABLE_FLAGS = true" >> ${WORK_DIR}/test/conf/app.ini; \
	echo "[server]" >> ${WORK_DIR}/test/conf/app.ini; \
	echo "ROOT_URL = ${FORGEJO_SDK_TEST_URL}" >> ${WORK_DIR}/test/conf/app.ini; \
	echo "[quota]" >> ${WORK_DIR}/test/conf/app.ini; \
	echo "ENABLED = true" >> ${WORK_DIR}/test/conf/app.ini; \
	echo "[migrations]" >> ${WORK_DIR}/test/conf/app.ini; \
	echo "ALLOW_LOCALNETWORKS = true" >> ${WORK_DIR}/test/conf/app.ini; \
	echo "[federation]" >> ${WORK_DIR}/test/conf/app.ini; \
	echo "ENABLED = true" >> ${WORK_DIR}/test/conf/app.ini; \
	${WORK_DIR}/test/forgejo-main migrate -c ${WORK_DIR}/test/conf/app.ini; \
	${WORK_DIR}/test/forgejo-main admin user create \
		--username=${FORGEJO_SDK_TEST_USERNAME} \
		--password=${FORGEJO_SDK_TEST_PASSWORD} \
		--email=${FORGEJO_SDK_TEST_EMAIL} \
		--admin=true \
		--must-change-password=false \
		--access-token \
		-c ${WORK_DIR}/test/conf/app.ini; \
	${WORK_DIR}/test/forgejo-main web -c ${WORK_DIR}/test/conf/app.ini

.PHONY: test-instance-docker
test-instance-docker: ## Start a forgejo instance using Docker.
ifeq ($(DOCKER_AVAILABLE),no)
	@echo "Error: Docker is not available in PATH"
	@echo "Please install Docker from https://www.docker.com/products/docker-desktop"
	@exit 1
endif
	@echo "Starting Forgejo test instance in Docker..."
	@docker volume create forgejo-test-data > /dev/null 2>&1 || true
	@secret_key=$$(docker run --rm codeberg.org/forgejo/forgejo:${FORGEJO_VERSION} forgejo generate secret SECRET_KEY); \
	internal_token=$$(docker run --rm codeberg.org/forgejo/forgejo:${FORGEJO_VERSION} forgejo generate secret INTERNAL_TOKEN); \
	docker run -d --name forgejo-test \
		-p 3000:3000 \
		-v forgejo-test-data:/data \
		-e FORGEJO__security__INSTALL_LOCK=true \
		-e FORGEJO__security__SECRET_KEY=$$secret_key \
		-e FORGEJO__security__INTERNAL_TOKEN=$$internal_token \
		-e FORGEJO__security__PASSWORD_COMPLEXITY=off \
		-e FORGEJO__security__DISABLE_GIT_HOOKS=false \
		-e FORGEJO__repository__ENABLE_FLAGS=true \
		-e FORGEJO__database__DB_TYPE=sqlite3 \
		-e FORGEJO__server__ROOT_URL=${FORGEJO_SDK_TEST_URL} \
		-e FORGEJO__service__DISABLE_REGISTRATION=false \
		-e FORGEJO__admin__DISABLE_REGULAR_ORG_CREATION=false \
		-e FORGEJO__quota__ENABLED=true \
		-e FORGEJO__migrations__ALLOW_LOCALNETWORKS=true \
		-e FORGEJO__federation__ENABLED=true \
		codeberg.org/forgejo/forgejo:${FORGEJO_VERSION} > /dev/null 2>&1 || true
	@echo "Waiting for Forgejo to start..."
	@for i in 1 2 3 4 5 6 7 8 9 10 11 12 13 14 15 16 17 18 19 20; do \
		if curl --noproxy "*" -s "${FORGEJO_SDK_TEST_URL}/api/v1/version" > /dev/null 2>&1; then \
			echo "Forgejo is ready!"; \
			break; \
		fi; \
		echo "Waiting... ($$i/20)"; \
		sleep 2; \
	done
	@echo "Creating test user..."
	@docker exec -u git forgejo-test forgejo admin user create \
		--username=${FORGEJO_SDK_TEST_USERNAME} \
		--password=${FORGEJO_SDK_TEST_PASSWORD} \
		--email=${FORGEJO_SDK_TEST_EMAIL} \
		--admin=true \
		--must-change-password=false \
		--access-token > /dev/null 2>&1 || echo "User might already exist"
	@echo "Test instance is ready at ${FORGEJO_SDK_TEST_URL}"

.PHONY: test-instance-stop
test-instance-stop: ## Stop the forgejo test instance.
	@echo "Stopping Forgejo test instance..."
ifeq ($(DOCKER_AVAILABLE),yes)
	@docker stop forgejo-test > /dev/null 2>&1 || true
	@docker rm forgejo-test > /dev/null 2>&1 || true
	@docker volume rm forgejo-test-data > /dev/null 2>&1 || true
endif
	@pkill -f "forgejo-main web" > /dev/null 2>&1 || true
	@echo "Test instance stopped"

.PHONY: bench
bench: ## Run benchmarks.
	cd forgejo && $(GO) test -run=XXXXXX -benchtime=10s -bench=. || exit 1

##@ Building

.PHONY: build
build: ## Build the SDK.
	cd forgejo && $(GO) build
