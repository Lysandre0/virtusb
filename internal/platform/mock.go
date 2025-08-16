package platform

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// MockPlatform implements Platform for testing
type MockPlatform struct {
	CommonPlatform
}

func (p *MockPlatform) RequireRoot() error {
	// In mock mode, we don't need root privileges
	return nil
}

func (p *MockPlatform) EnsureEnvironment(ctx context.Context) error {
	// Create temporary directories for mock mode
	mockStateDir := "/tmp/virtusb_test/etc/virtusb"
	mockImageDir := "/tmp/virtusb_test/var/lib/virtusb"

	if err := p.CreateDirectory(mockStateDir); err != nil {
		return fmt.Errorf("failed to create mock state directory: %w", err)
	}
	if err := p.CreateDirectory(mockImageDir); err != nil {
		return fmt.Errorf("failed to create mock image directory: %w", err)
	}

	// Create mock directories for configfs in temporary directory
	mockGadgetRoot := "/tmp/virtusb_test/sys/kernel/config/usb_gadget"
	if err := p.CreateDirectory(mockGadgetRoot); err != nil {
		return fmt.Errorf("failed to create mock gadget root: %w", err)
	}
	// Update gadget root path for mock mode
	p.config.GadgetRoot = mockGadgetRoot
	if err := p.CreateDirectory("/tmp/virtusb_test/sys/class/udc/dummy_udc.0"); err != nil {
		return fmt.Errorf("failed to create mock UDC: %w", err)
	}

	// Update configuration paths for mock mode
	p.config.StateDir = mockStateDir
	p.config.ImageDir = mockImageDir

	return nil
}

func (p *MockPlatform) GetFirstUDC() (string, error) {
	// Create a dummy UDC in temporary directory
	udcPath := "/tmp/virtusb_test/sys/class/udc/dummy_udc.0"
	if err := p.CreateDirectory(udcPath); err != nil {
		return "", fmt.Errorf("failed to create mock UDC: %w", err)
	}
	return "dummy_udc.0", nil
}

func (p *MockPlatform) IsUDCAvailable() bool {
	_, err := p.GetFirstUDC()
	return err == nil
}

func (p *MockPlatform) LoadModule(name string) error {
	// In mock mode, simulate loading
	return nil
}

func (p *MockPlatform) IsModuleLoaded(name string) bool {
	// In mock mode, simulate all modules loaded
	return true
}

func (p *MockPlatform) IsMountpoint(path string) bool {
	// In mock mode, simulate configfs mounted
	return true
}

func (p *MockPlatform) MountConfigFS() error {
	// In mock mode, simulate mounting
	return nil
}

func (p *MockPlatform) Which(binary string) string {
	// In mock mode, return dummy path
	return "/usr/bin/" + binary
}

func (p *MockPlatform) RunCommand(name string, args ...string) error {
	// In mock mode, simulate successful execution
	return nil
}

func (p *MockPlatform) RunCommandQuiet(name string, args ...string) error {
	// In mock mode, simulate successful execution
	return nil
}

func (p *MockPlatform) WriteString(path, content string) error {
	// Validate path is not empty
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Validate path is not a directory traversal attempt
	if strings.Contains(path, "..") {
		return fmt.Errorf("path traversal not allowed")
	}

	// Create parent directory if necessary
	dir := filepath.Dir(path)
	if err := p.CreateDirectory(dir); err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", path, err)
	}

	// Use more restrictive permissions for security
	return os.WriteFile(path, []byte(content), 0600)
}

func (p *MockPlatform) ReadString(path string) (string, error) {
	// Validate path is not empty
	if path == "" {
		return "", fmt.Errorf("path cannot be empty")
	}

	// Validate path is not a directory traversal attempt
	if strings.Contains(path, "..") {
		return "", fmt.Errorf("path traversal not allowed")
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (p *MockPlatform) FileExists(path string) bool {
	// Validate path is not empty
	if path == "" {
		return false
	}

	// Validate path is not a directory traversal attempt
	if strings.Contains(path, "..") {
		return false
	}

	_, err := os.Stat(path)
	return err == nil
}

func (p *MockPlatform) CreateDirectory(path string) error {
	// Validate path is not empty
	if path == "" {
		return fmt.Errorf("path cannot be empty")
	}

	// Validate path is not a directory traversal attempt
	if strings.Contains(path, "..") {
		return fmt.Errorf("path traversal not allowed")
	}

	// Use more restrictive permissions for security
	return os.MkdirAll(path, 0700)
}
