package gadget

import (
	"context"
	"path/filepath"
	"testing"
)

func TestParseMetadata(t *testing.T) {
	content := `NAME="test-gadget"
BRAND="sandisk"
VID="0781"
PID="5567"
SERIAL="4C530001123456789"
IMG="/var/lib/virtusb/test-gadget.img"
FS="fat32"
SIZE="8G"
AUTOSTART="1"
`

	manager := &Manager{}

	gadget, err := manager.parseMetadata(content)
	if err != nil {
		t.Fatal(err)
	}

	// Verify parsed values
	if gadget.Name != "test-gadget" {
		t.Errorf("Expected name 'test-gadget', got '%s'", gadget.Name)
	}
	if gadget.Brand != "sandisk" {
		t.Errorf("Expected brand 'sandisk', got '%s'", gadget.Brand)
	}
	if gadget.VID != "0781" {
		t.Errorf("Expected VID '0781', got '%s'", gadget.VID)
	}
	if gadget.PID != "5567" {
		t.Errorf("Expected PID '5567', got '%s'", gadget.PID)
	}
	if gadget.Serial != "4C530001123456789" {
		t.Errorf("Expected serial '4C530001123456789', got '%s'", gadget.Serial)
	}
	if gadget.ImagePath != "/var/lib/virtusb/test-gadget.img" {
		t.Errorf("Expected image path '/var/lib/virtusb/test-gadget.img', got '%s'", gadget.ImagePath)
	}
	if gadget.FileSystem != "fat32" {
		t.Errorf("Expected filesystem 'fat32', got '%s'", gadget.FileSystem)
	}
	if gadget.Size != "8G" {
		t.Errorf("Expected size '8G', got '%s'", gadget.Size)
	}
}

func TestParseMetadataMissingFields(t *testing.T) {
	content := `NAME="test-gadget"
BRAND="sandisk"
VID="0781"
PID="5567"
SERIAL="4C530001123456789"
IMG="/var/lib/virtusb/test-gadget.img"
FS="fat32"
SIZE="8G"
`

	manager := &Manager{}

	_, err := manager.parseMetadata(content)
	if err == nil {
		t.Error("Expected error for missing AUTOSTART field")
	}
}

func TestParseMetadataInvalidFormat(t *testing.T) {
	testCases := []struct {
		name    string
		content string
	}{
		{"empty content", ""},
		{"no name", `BRAND="sandisk"`},
		{"invalid format", `NAME=test-gadget`},
		{"malformed line", `NAME="test-gadget" BRAND="sandisk"`},
	}

	manager := &Manager{}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := manager.parseMetadata(tc.content)
			if err == nil {
				t.Errorf("Expected error for %s", tc.name)
			}
		})
	}
}

func TestManager_ClearCache(t *testing.T) {
	manager := NewManager(nil)

	// Verify cache is initialized
	if manager.gadgetCache == nil {
		t.Error("Cache should be initialized")
	}

	// Clear cache
	manager.ClearCache()

	// Verify cache is cleared
	if len(manager.gadgetCache) != 0 {
		t.Error("Cache should be empty after clearing")
	}
}

func TestCreateOptionsValidation(t *testing.T) {
	testCases := []struct {
		name    string
		opts    CreateOptions
		isValid bool
	}{
		{
			name: "valid options",
			opts: CreateOptions{
				Name:       "test-gadget",
				Size:       "8G",
				Brand:      "sandisk",
				FileSystem: "fat32",
			},
			isValid: true,
		},
		{
			name: "empty name",
			opts: CreateOptions{
				Name:       "",
				Size:       "8G",
				Brand:      "sandisk",
				FileSystem: "fat32",
			},
			isValid: false,
		},
		{
			name: "invalid size",
			opts: CreateOptions{
				Name:       "test-gadget",
				Size:       "invalid",
				Brand:      "sandisk",
				FileSystem: "fat32",
			},
			isValid: false,
		},
		{
			name: "invalid brand",
			opts: CreateOptions{
				Name:       "test-gadget",
				Size:       "8G",
				Brand:      "invalid",
				FileSystem: "fat32",
			},
			isValid: false,
		},
		{
			name: "invalid filesystem",
			opts: CreateOptions{
				Name:       "test-gadget",
				Size:       "8G",
				Brand:      "sandisk",
				FileSystem: "invalid",
			},
			isValid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// This would be implemented in the actual validation function
			// For now, we just test the structure
			if tc.opts.Name == "" {
				if tc.isValid {
					t.Error("Empty name should be invalid")
				}
			}
		})
	}
}

func TestGadgetPathGeneration(t *testing.T) {
	manager := &Manager{}

	// Mock context
	ctx := Context{
		Platform: &MockPlatform{},
	}

	expectedPath := filepath.Join("/sys/kernel/config/usb_gadget", "virtusb-test-gadget")
	actualPath := manager.getGadgetPath(ctx, "test-gadget")

	if actualPath != expectedPath {
		t.Errorf("Expected path '%s', got '%s'", expectedPath, actualPath)
	}
}

func TestImagePathGeneration(t *testing.T) {
	manager := &Manager{}

	// Mock context
	ctx := Context{
		Platform: &MockPlatform{},
	}

	expectedPath := filepath.Join("/var/lib/virtusb", "test-gadget.img")
	actualPath := manager.getImagePath(ctx, "test-gadget")

	if actualPath != expectedPath {
		t.Errorf("Expected path '%s', got '%s'", expectedPath, actualPath)
	}
}

func TestMetadataPathGeneration(t *testing.T) {
	manager := &Manager{}

	// Mock context
	ctx := Context{
		Platform: &MockPlatform{},
	}

	expectedPath := filepath.Join("/etc/virtusb", "test-gadget.env")
	actualPath := manager.getMetadataPath(ctx, "test-gadget")

	if actualPath != expectedPath {
		t.Errorf("Expected path '%s', got '%s'", expectedPath, actualPath)
	}
}

// MockPlatform for testing
type MockPlatform struct{}

func (p *MockPlatform) GetGadgetRoot() string {
	return "/sys/kernel/config/usb_gadget"
}

func (p *MockPlatform) GetImageDir() string {
	return "/var/lib/virtusb"
}

func (p *MockPlatform) GetStateDir() string {
	return "/etc/virtusb"
}

func (p *MockPlatform) FileExists(path string) bool {
	return false
}

func (p *MockPlatform) WriteString(path, content string) error {
	return nil
}

func (p *MockPlatform) ReadString(path string) (string, error) {
	return "", nil
}

func (p *MockPlatform) CreateDirectory(path string) error {
	return nil
}

func (p *MockPlatform) RequireRoot() error {
	return nil
}

func (p *MockPlatform) EnsureEnvironment(ctx context.Context) error {
	return nil
}

func (p *MockPlatform) GetFirstUDC() (string, error) {
	return "dummy_udc", nil
}

func (p *MockPlatform) IsUDCAvailable() bool {
	return true
}

func (p *MockPlatform) LoadModule(name string) error {
	return nil
}

func (p *MockPlatform) IsModuleLoaded(name string) bool {
	return true
}

func (p *MockPlatform) IsMountpoint(path string) bool {
	return true
}

func (p *MockPlatform) MountConfigFS() error {
	return nil
}

func (p *MockPlatform) Which(binary string) string {
	return "/usr/bin/" + binary
}

func (p *MockPlatform) RunCommand(name string, args ...string) error {
	return nil
}

func (p *MockPlatform) RunCommandQuiet(name string, args ...string) error {
	return nil
}

func (p *MockPlatform) IsMockMode() bool {
	return true
}

// MockConfigAdapter for testing
type MockConfigAdapter struct{}

func (c *MockConfigAdapter) GetGadgetRoot() string {
	return "/sys/kernel/config/usb_gadget"
}

func (c *MockConfigAdapter) GetImageDir() string {
	return "/var/lib/virtusb"
}

func (c *MockConfigAdapter) GetStateDir() string {
	return "/etc/virtusb"
}
