MODULE     := github.com/oaswrap/gswag
COVER_OUT  := coverage.out
COVER_HTML := coverage.html
EXAMPLES   := stdlib gin echo gorilla chi fiber todo parallel

.PHONY: all build test cover lint vet tidy clean fmt \
        examples validate-examples install help update-golden

all: tidy vet fmt test

# ---------------------------------------------------------------------------
# Test
# ---------------------------------------------------------------------------

test:
	go test ./... 

update-golden:
	UPDATE_GOLDEN=true go test ./...

test-verbose:
	go test -v ./...

test-race:
	go test -race ./...

cover:
	GOCOVERDIR=coverage go test -coverprofile=$(COVER_OUT) -covermode=atomic -coverpkg=./... ./...
	go tool cover -func=$(COVER_OUT)

cover-html: cover
	go tool cover -html=$(COVER_OUT) -o $(COVER_HTML)
	@echo "Coverage report: $(COVER_HTML)"
	@which open > /dev/null && open $(COVER_HTML) || echo "Please open $(COVER_HTML) in your browser"

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

help:
	@echo "Usage: make <target>"
	@echo ""
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
	@echo "  clean              Remove build artefacts"
	@echo "  update-golden      Regenerate golden test fixtures"
