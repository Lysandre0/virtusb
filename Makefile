.PHONY: build install clean test help

# Variables
BINARY_NAME=virtusb
BUILD_DIR=build
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")

# Default targets
all: build

# Build
build:
	@echo "Building $(BINARY_NAME) v$(VERSION)..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags "-X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME) cmd/virtusb/main.go
	@echo "✅ Build successful: $(BUILD_DIR)/$(BINARY_NAME)"

# Install
install: build
	@echo "Installing $(BINARY_NAME)..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	@echo "✅ Installation successful"

# Clean
clean:
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	@echo "✅ Clean completed"

# Tests
test:
	@echo "Running tests..."
	go test ./...
	@echo "✅ Tests completed"

# Test in mock mode
test-mock: build
	@echo "Testing in mock mode..."
	@rm -rf /tmp/virtusb_test
	MOCK=1 $(BUILD_DIR)/$(BINARY_NAME) diagnose
	MOCK=1 $(BUILD_DIR)/$(BINARY_NAME) create test --size 64M
	MOCK=1 $(BUILD_DIR)/$(BINARY_NAME) list
	MOCK=1 $(BUILD_DIR)/$(BINARY_NAME) delete test
	@echo "✅ Mock tests completed"

# Release
release: clean build
	@echo "Creating release for $(VERSION)..."
	@mkdir -p dist
	@cp $(BUILD_DIR)/$(BINARY_NAME) dist/$(BINARY_NAME)_$(shell uname -s | tr '[:upper:]' '[:lower:]')_$(shell uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')
	@echo "✅ Release binary created in dist/"

# Help
help:
	@echo "Available targets:"
	@echo "  build      - Build the project"
	@echo "  install    - Install the binary"
	@echo "  clean      - Clean build files"
	@echo "  test       - Run tests"
	@echo "  test-mock  - Test in mock mode"
	@echo "  release    - Create release binary"
	@echo "  help       - Show this help"
