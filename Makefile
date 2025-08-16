.PHONY: build install clean test help

BINARY_NAME=virtusb
BUILD_DIR=build
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
GOOS=$(shell go env GOOS)
GOARCH=$(shell go env GOARCH)

LDFLAGS=-ldflags "-X main.version=$(VERSION)"
BUILD_FLAGS=-trimpath -ldflags="-s -w"

all: build

build:
	@echo "Building $(BINARY_NAME) v$(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=0 GOOS=$(GOOS) GOARCH=$(GOARCH) go build $(BUILD_FLAGS) $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) cmd/virtusb/main.go
	@echo "✅ Build successful: $(BUILD_DIR)/$(BINARY_NAME)"
	@ls -lh $(BUILD_DIR)/$(BINARY_NAME)

install: build
	@echo "Installing $(BINARY_NAME)..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "✅ Installation successful"

clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	@echo "✅ Clean completed"

test:
	@echo "Running tests..."
	go test -v ./...
	@echo "✅ Tests completed"

test-mock: build
	@echo "Testing in mock mode..."
	@rm -rf /tmp/virtusb_test
	MOCK=1 $(BUILD_DIR)/$(BINARY_NAME) diagnose
	MOCK=1 $(BUILD_DIR)/$(BINARY_NAME) create test --size 64M
	MOCK=1 $(BUILD_DIR)/$(BINARY_NAME) list
	MOCK=1 $(BUILD_DIR)/$(BINARY_NAME) delete test
	@echo "✅ Mock tests completed"

fmt:
	@echo "Formatting code..."
	go fmt ./...
	@echo "✅ Code formatting completed"

vet:
	@echo "Vetting code..."
	go vet ./...
	@echo "✅ Code vetting completed"

help:
	@echo "Available targets:"
	@echo "  build      - Build the project"
	@echo "  install    - Install the binary"
	@echo "  clean      - Clean build files"
	@echo "  test       - Run tests"
	@echo "  test-mock  - Test in mock mode"
	@echo "  fmt        - Format code"
	@echo "  vet        - Vet code"
	@echo "  help       - Show this help"
