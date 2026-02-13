GO ?= go

WORK_DIR   := $(shell pwd)

# Test configuration
FORGEJO_SDK_TEST_URL ?= http://localhost:3000
FORGEJO_SDK_TEST_USERNAME ?= test01
FORGEJO_SDK_TEST_PASSWORD ?= test01
FORGEJO_SDK_TEST_EMAIL ?= test01@forgejo.org

# Common test instance configuration
FORGEJO_SECRET_KEY := 2crAW4UANgvLipDS6U5obRcFosjSJHQANll6MNfX7P0G3se3fKcCwwK3szPyGcbo
FORGEJO_INTERNAL_TOKEN := eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYmYiOjE1NTg4MzY4ODB9.LoKQyK5TN_0kMJFVHWUW0uDAyoGjDP6Mkup4ps2VJN4

PACKAGE := codeberg.org/mvdkleijn/forgejo-sdk/forgejo/v2

GOFUMPT_PACKAGE ?= mvdan.cc/gofumpt@v0.7.0
GOLANGCI_LINT_VERSION ?= v2.9.0

FORGEJO_VERSION := 9.0.3
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
all: clean test build

.PHONY: help
help:
	@echo "Make Routines:"
	@echo " - \"\"                      run \"make clean test build\""
	@echo " - build                   build sdk"
	@echo " - clean                   clean build artifacts and test instances"
	@echo " - fmt                     format the code"
	@echo " - lint / ci-lint          run golint"
	@echo " - vet                     examines Go source code and reports suspicious constructs"
	@echo " - test                    run unit tests (requires a running forgejo instance)"
	@echo " - test-instance           start a forgejo instance for test (auto-detects method)"
	@echo " - test-instance-native    start a native forgejo instance (Linux only)"
	@echo " - test-instance-docker    start a forgejo instance using Docker"
	@echo " - test-instance-stop      stop the forgejo test instance"
	@echo " - bench                   run benchmarks"
	@echo ""
	@echo "Docker available: $(DOCKER_AVAILABLE)"


.PHONY: clean
clean:
	rm -r -f test test-cache
	cd forgejo && $(GO) clean -i ./...

.PHONY: fmt
fmt:
	find . -name "*.go" -type f | xargs gofmt -s -w; \
	$(GO) run $(GOFUMPT_PACKAGE) -extra -w ./forgejo

.PHONY: vet
vet:
	# Default vet
	cd forgejo && $(GO) vet $(PACKAGE)

.PHONY: ci-lint
ci-lint:
	@cd forgejo/; echo -n "gofumpt ...";\
	diff=$$($(GO) run $(GOFUMPT_PACKAGE) -extra -l .); \
	if [ -n "$$diff" ]; then \
		echo; echo "Not gofumpt-ed"; \
		exit 1; \
	fi; echo " done"; echo -n "golangci-lint ...";\
	curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b $$($(GO) env GOPATH)/bin $(GOLANGCI_LINT_VERSION) \
	$$($(GO) env GOPATH)/bin/golangci-lint run --timeout 5m; \
	if [ $$? -eq 1 ]; then \
		echo; echo "Doesn't pass golangci-lint"; \
		exit 1; \
	fi; echo " done"; \
	cd -; \

.PHONY: test
test:
	@export FORGEJO_SDK_TEST_URL=${FORGEJO_SDK_TEST_URL}; export FORGEJO_SDK_TEST_USERNAME=${FORGEJO_SDK_TEST_USERNAME}; export FORGEJO_SDK_TEST_PASSWORD=${FORGEJO_SDK_TEST_PASSWORD}; \
	if [ -z "$(shell curl --noproxy "*" "${FORGEJO_SDK_TEST_URL}/api/v1/version" 2> /dev/null)" ]; then \echo "No test-instance detected!"; exit 1; else \
	    cd forgejo && $(GO) test -race -cover -coverprofile coverage.out; \
	fi

.PHONY: test-instance
test-instance:
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
test-instance-native:
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
	echo "[security]" > ${WORK_DIR}/test/conf/app.ini; \
	echo "INTERNAL_TOKEN = $(FORGEJO_INTERNAL_TOKEN)" >> ${WORK_DIR}/test/conf/app.ini; \
	echo "INSTALL_LOCK   = true" >> ${WORK_DIR}/test/conf/app.ini; \
	echo "SECRET_KEY     = $(FORGEJO_SECRET_KEY)" >> ${WORK_DIR}/test/conf/app.ini; \
	echo "PASSWORD_COMPLEXITY = off" >> ${WORK_DIR}/test/conf/app.ini; \
	echo "[database]" >> ${WORK_DIR}/test/conf/app.ini; \
	echo "DB_TYPE = sqlite3" >> ${WORK_DIR}/test/conf/app.ini; \
	echo "[repository]" >> ${WORK_DIR}/test/conf/app.ini; \
	echo "ROOT = ${WORK_DIR}/test/data/" >> ${WORK_DIR}/test/conf/app.ini; \
	echo "[server]" >> ${WORK_DIR}/test/conf/app.ini; \
	echo "ROOT_URL = ${FORGEJO_SDK_TEST_URL}" >> ${WORK_DIR}/test/conf/app.ini; \
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
test-instance-docker:
ifeq ($(DOCKER_AVAILABLE),no)
	@echo "Error: Docker is not available in PATH"
	@echo "Please install Docker from https://www.docker.com/products/docker-desktop"
	@exit 1
endif
	@echo "Starting Forgejo test instance in Docker..."
	@docker volume create forgejo-test-data > /dev/null 2>&1 || true
	@docker run -d --name forgejo-test \
		-p 3000:3000 \
		-v forgejo-test-data:/data \
		-e FORGEJO__security__INSTALL_LOCK=true \
		-e FORGEJO__security__SECRET_KEY=$(FORGEJO_SECRET_KEY) \
		-e FORGEJO__security__INTERNAL_TOKEN=$(FORGEJO_INTERNAL_TOKEN) \
		-e FORGEJO__security__PASSWORD_COMPLEXITY=off \
		-e FORGEJO__database__DB_TYPE=sqlite3 \
		-e FORGEJO__server__ROOT_URL=${FORGEJO_SDK_TEST_URL} \
		-e FORGEJO__service__DISABLE_REGISTRATION=false \
		-e FORGEJO__admin__DISABLE_REGULAR_ORG_CREATION=false \
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
test-instance-stop:
	@echo "Stopping Forgejo test instance..."
ifeq ($(DOCKER_AVAILABLE),yes)
	@docker stop forgejo-test > /dev/null 2>&1 || true
	@docker rm forgejo-test > /dev/null 2>&1 || true
	@docker volume rm forgejo-test-data > /dev/null 2>&1 || true
endif
	@pkill -f "forgejo-main web" > /dev/null 2>&1 || true
	@echo "Test instance stopped"

.PHONY: bench
bench:
	cd forgejo && $(GO) test -run=XXXXXX -benchtime=10s -bench=. || exit 1

.PHONY: build
build:
	cd forgejo && $(GO) build
