MODULE     := github.com/oaswrap/gswag
CMD        := ./cmd/gswag
BIN        := bin/gswag
COVER_OUT  := coverage.out
COVER_HTML := coverage.html
EXAMPLES   := stdlib init-example gin echo chi fiber petstore

.PHONY: all build test cover lint vet tidy clean fmt \
        examples validate-examples install help

all: tidy vet fmt test build

# ---------------------------------------------------------------------------
# Build
# ---------------------------------------------------------------------------

build:
	go build -o $(BIN) $(CMD)

install:
	go install $(CMD)

# ---------------------------------------------------------------------------
# Test
# ---------------------------------------------------------------------------

test:
	go test ./...

test-verbose:
	go test -v ./...

test-race:
	go test -race ./...

cover:
	go test ./... -coverprofile=$(COVER_OUT) -covermode=atomic
	go tool cover -func=$(COVER_OUT)

cover-html: cover
	go tool cover -html=$(COVER_OUT) -o $(COVER_HTML)
	@echo "Coverage report: $(COVER_HTML)"

# ---------------------------------------------------------------------------
# Code quality
# ---------------------------------------------------------------------------

vet:
	go vet ./...

fmt:
	go fmt ./...

lint:
	@which golangci-lint > /dev/null || (echo "golangci-lint not found; run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" && exit 1)
	golangci-lint run ./...

tidy:
	go mod tidy
	@for name in $(EXAMPLES); do \
		echo "==> examples/$$name"; \
		(cd examples/$$name && go mod tidy) || exit 1; \
	done

# ---------------------------------------------------------------------------
# Examples
# ---------------------------------------------------------------------------

examples:
	@for name in $(EXAMPLES); do \
		echo "==> example/$$name"; \
		(cd examples/$$name && go mod tidy && go test ./... -v) || exit 1; \
	done

validate-examples: examples
	@which $(BIN) > /dev/null || $(MAKE) build
	@for name in $(EXAMPLES); do \
		spec=examples/$$name/docs/openapi.yaml; \
		if [ -f "$$spec" ]; then \
			echo "==> validating $$spec"; \
			$(BIN) validate "$$spec" || exit 1; \
		fi; \
	done

# ---------------------------------------------------------------------------
# CLI helpers
# ---------------------------------------------------------------------------

validate: build
	$(BIN) validate $(SPEC)

diff: build
	$(BIN) diff $(BASE) $(HEAD)

# ---------------------------------------------------------------------------
# Maintenance
# ---------------------------------------------------------------------------

clean:
	rm -f $(BIN) $(COVER_OUT) $(COVER_HTML)
	rm -rf bin/

help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "  build              Build the gswag CLI binary to $(BIN)"
	@echo "  install            Install the gswag CLI via go install"
	@echo "  test               Run all unit tests"
	@echo "  test-verbose       Run tests with -v"
	@echo "  test-race          Run tests with race detector"
	@echo "  cover              Run tests and print coverage summary"
	@echo "  cover-html         Run tests and open HTML coverage report"
	@echo "  vet                Run go vet"
	@echo "  fmt                Run go fmt"
	@echo "  lint               Run golangci-lint (must be installed)"
	@echo "  tidy               Run go mod tidy"
	@echo "  examples           Build and test all examples"
	@echo "  validate-examples  Test examples and validate their generated specs"
	@echo "  validate SPEC=<f>  Validate a spec file (requires SPEC=)"
	@echo "  diff BASE=<f> HEAD=<f>  Diff two spec files"
	@echo "  clean              Remove build artefacts"
