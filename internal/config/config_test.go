package config

import (
	"os"
	"testing"
)

func TestLoadFromEnv(t *testing.T) {
	// Save original environment
	originalEnv := make(map[string]string)
	for _, key := range []string{"MOCK", "VIRTUSB_ROOT", "VIRTUSB_STATE_DIR", "VIRTUSB_IMAGE_DIR", "VIRTUSB_DEFAULT_SIZE", "VIRTUSB_DEFAULT_BRAND", "VIRTUSB_DEFAULT_FS"} {
		originalEnv[key] = os.Getenv(key)
	}

	// Clean up after test
	defer func() {
		for key, value := range originalEnv {
			if value != "" {
				os.Setenv(key, value)
			} else {
				os.Unsetenv(key)
			}
		}
	}()

	// Test default values
	config := LoadFromEnv()

	if config.GadgetRoot != DefaultGadgetRoot {
		t.Errorf("Expected GadgetRoot %s, got %s", DefaultGadgetRoot, config.GadgetRoot)
	}
	if config.StateDir != DefaultStateDir {
		t.Errorf("Expected StateDir %s, got %s", DefaultStateDir, config.StateDir)
	}
	if config.ImageDir != DefaultImageDir {
		t.Errorf("Expected ImageDir %s, got %s", DefaultImageDir, config.ImageDir)
	}
	if config.USBIPBin != DefaultUSBIPBin {
		t.Errorf("Expected USBIPBin %s, got %s", DefaultUSBIPBin, config.USBIPBin)
	}
	if config.USBIPDBin != DefaultUSBIPDBin {
		t.Errorf("Expected USBIPDBin %s, got %s", DefaultUSBIPDBin, config.USBIPDBin)
	}
	if config.DefaultSize != DefaultSize {
		t.Errorf("Expected DefaultSize %s, got %s", DefaultSize, config.DefaultSize)
	}
	if config.DefaultBrand != DefaultBrand {
		t.Errorf("Expected DefaultBrand %s, got %s", DefaultBrand, config.DefaultBrand)
	}
	if config.DefaultFS != DefaultFS {
		t.Errorf("Expected DefaultFS %s, got %s", DefaultFS, config.DefaultFS)
	}
	if config.USBIPPort != DefaultUSBIPPort {
		t.Errorf("Expected USBIPPort %d, got %d", DefaultUSBIPPort, config.USBIPPort)
	}
}

func TestLoadFromEnvWithCustomValues(t *testing.T) {
	// Save original environment
	originalEnv := make(map[string]string)
	for _, key := range []string{"MOCK", "VIRTUSB_ROOT", "VIRTUSB_STATE_DIR", "VIRTUSB_IMAGE_DIR", "VIRTUSB_DEFAULT_SIZE", "VIRTUSB_DEFAULT_BRAND", "VIRTUSB_DEFAULT_FS"} {
		originalEnv[key] = os.Getenv(key)
	}

	// Clean up after test
	defer func() {
		for key, value := range originalEnv {
			if value != "" {
				os.Setenv(key, value)
			} else {
				os.Unsetenv(key)
			}
		}
	}()

	// Set custom values
	os.Setenv("MOCK", "1")
	os.Setenv("VIRTUSB_ROOT", "/custom/gadget/root")
	os.Setenv("VIRTUSB_STATE_DIR", "/custom/state")
	os.Setenv("VIRTUSB_IMAGE_DIR", "/custom/images")
	os.Setenv("VIRTUSB_DEFAULT_SIZE", "16G")
	os.Setenv("VIRTUSB_DEFAULT_BRAND", "kingston")
	os.Setenv("VIRTUSB_DEFAULT_FS", "exfat")

	config := LoadFromEnv()

	if !config.MockMode {
		t.Error("Expected MockMode to be true")
	}
	if config.GadgetRoot != "/custom/gadget/root" {
		t.Errorf("Expected GadgetRoot /custom/gadget/root, got %s", config.GadgetRoot)
	}
	if config.StateDir != "/custom/state" {
		t.Errorf("Expected StateDir /custom/state, got %s", config.StateDir)
	}
	if config.ImageDir != "/custom/images" {
		t.Errorf("Expected ImageDir /custom/images, got %s", config.ImageDir)
	}
	if config.DefaultSize != "16G" {
		t.Errorf("Expected DefaultSize 16G, got %s", config.DefaultSize)
	}
	if config.DefaultBrand != "kingston" {
		t.Errorf("Expected DefaultBrand kingston, got %s", config.DefaultBrand)
	}
	if config.DefaultFS != "exfat" {
		t.Errorf("Expected DefaultFS exfat, got %s", config.DefaultFS)
	}
}

func TestEnvOn(t *testing.T) {
	testCases := []struct {
		value    string
		expected bool
	}{
		{"1", true},
		{"true", true},
		{"yes", true},
		{"TRUE", true},
		{"YES", true},
		{"True", true},
		{"Yes", true},
		{"0", false},
		{"false", false},
		{"no", false},
		{"", false},
		{"invalid", false},
	}

	for _, tc := range testCases {
		t.Run(tc.value, func(t *testing.T) {
			os.Setenv("TEST_VAR", tc.value)
			defer os.Unsetenv("TEST_VAR")

			result := envOn("TEST_VAR")
			if result != tc.expected {
				t.Errorf("Expected %v for value '%s', got %v", tc.expected, tc.value, result)
			}
		})
	}
}

func TestValidation(t *testing.T) {
	testCases := []struct {
		name    string
		config  *Config
		isValid bool
	}{
		{
			name: "valid config",
			config: &Config{
				GadgetRoot:   "/sys/kernel/config/usb_gadget",
				StateDir:     "/etc/virtusb",
				ImageDir:     "/var/lib/virtusb",
				DefaultSize:  "8G",
				DefaultBrand: "sandisk",
				DefaultFS:    "fat32",
			},
			isValid: true,
		},
		{
			name: "empty gadget root",
			config: &Config{
				GadgetRoot:   "",
				StateDir:     "/etc/virtusb",
				ImageDir:     "/var/lib/virtusb",
				DefaultSize:  "8G",
				DefaultBrand: "sandisk",
				DefaultFS:    "fat32",
			},
			isValid: false,
		},
		{
			name: "empty state dir",
			config: &Config{
				GadgetRoot:   "/sys/kernel/config/usb_gadget",
				StateDir:     "",
				ImageDir:     "/var/lib/virtusb",
				DefaultSize:  "8G",
				DefaultBrand: "sandisk",
				DefaultFS:    "fat32",
			},
			isValid: false,
		},
		{
			name: "empty image dir",
			config: &Config{
				GadgetRoot:   "/sys/kernel/config/usb_gadget",
				StateDir:     "/etc/virtusb",
				ImageDir:     "",
				DefaultSize:  "8G",
				DefaultBrand: "sandisk",
				DefaultFS:    "fat32",
			},
			isValid: false,
		},
		{
			name: "invalid size",
			config: &Config{
				GadgetRoot:   "/sys/kernel/config/usb_gadget",
				StateDir:     "/etc/virtusb",
				ImageDir:     "/var/lib/virtusb",
				DefaultSize:  "invalid",
				DefaultBrand: "sandisk",
				DefaultFS:    "fat32",
			},
			isValid: false,
		},
		{
			name: "invalid brand",
			config: &Config{
				GadgetRoot:   "/sys/kernel/config/usb_gadget",
				StateDir:     "/etc/virtusb",
				ImageDir:     "/var/lib/virtusb",
				DefaultSize:  "8G",
				DefaultBrand: "invalid",
				DefaultFS:    "fat32",
			},
			isValid: false,
		},
		{
			name: "invalid filesystem",
			config: &Config{
				GadgetRoot:   "/sys/kernel/config/usb_gadget",
				StateDir:     "/etc/virtusb",
				ImageDir:     "/var/lib/virtusb",
				DefaultSize:  "8G",
				DefaultBrand: "sandisk",
				DefaultFS:    "invalid",
			},
			isValid: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.config.Validate()
			if tc.isValid && err != nil {
				t.Errorf("Expected valid config, got error: %v", err)
			}
			if !tc.isValid && err == nil {
				t.Error("Expected invalid config, but validation passed")
			}
		})
	}
}

func TestIsValidSize(t *testing.T) {
	testCases := []struct {
		size     string
		expected bool
	}{
		{"64M", true},
		{"128M", true},
		{"256M", true},
		{"512M", true},
		{"1G", true},
		{"2G", true},
		{"4G", true},
		{"8G", true},
		{"16G", true},
		{"32G", true},
		{"64G", true},
		{"32M", false},
		{"128G", false},
		{"invalid", false},
		{"", false},
		{"8GB", false},
		{"8g", false},
	}

	for _, tc := range testCases {
		t.Run(tc.size, func(t *testing.T) {
			result := isValidSize(tc.size)
			if result != tc.expected {
				t.Errorf("Expected %v for size '%s', got %v", tc.expected, tc.size, result)
			}
		})
	}
}

func TestIsValidBrand(t *testing.T) {
	testCases := []struct {
		brand    string
		expected bool
	}{
		{"sandisk", true},
		{"kingston", true},
		{"corsair", true},
		{"samsung", true},
		{"generic", true},
		{"Sandisk", true},
		{"SANDISK", true},
		{"invalid", false},
		{"", false},
	}

	for _, tc := range testCases {
		t.Run(tc.brand, func(t *testing.T) {
			result := isValidBrand(tc.brand)
			if result != tc.expected {
				t.Errorf("Expected %v for brand '%s', got %v", tc.expected, tc.brand, result)
			}
		})
	}
}

func TestIsValidFS(t *testing.T) {
	testCases := []struct {
		fs       string
		expected bool
	}{
		{"fat32", true},
		{"exfat", true},
		{"none", true},
		{"Fat32", true},
		{"EXFAT", true},
		{"invalid", false},
		{"", false},
	}

	for _, tc := range testCases {
		t.Run(tc.fs, func(t *testing.T) {
			result := isValidFS(tc.fs)
			if result != tc.expected {
				t.Errorf("Expected %v for filesystem '%s', got %v", tc.expected, tc.fs, result)
			}
		})
	}
}

func TestGetValidSizes(t *testing.T) {
	sizes := GetValidSizes()
	expectedCount := 11

	if len(sizes) != expectedCount {
		t.Errorf("Expected %d valid sizes, got %d", expectedCount, len(sizes))
	}

	// Check that all expected sizes are present
	expectedSizes := []string{"64M", "128M", "256M", "512M", "1G", "2G", "4G", "8G", "16G", "32G", "64G"}
	for _, expected := range expectedSizes {
		found := false
		for _, size := range sizes {
			if size == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected size %s not found in valid sizes", expected)
		}
	}
}

func TestGetValidBrands(t *testing.T) {
	brands := GetValidBrands()
	expectedCount := 5

	if len(brands) != expectedCount {
		t.Errorf("Expected %d valid brands, got %d", expectedCount, len(brands))
	}

	// Check that all expected brands are present
	expectedBrands := []string{"sandisk", "kingston", "corsair", "samsung", "generic"}
	for _, expected := range expectedBrands {
		found := false
		for _, brand := range brands {
			if brand == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected brand %s not found in valid brands", expected)
		}
	}
}

func TestGetValidFilesystems(t *testing.T) {
	fs := GetValidFilesystems()
	expectedCount := 3

	if len(fs) != expectedCount {
		t.Errorf("Expected %d valid filesystems, got %d", expectedCount, len(fs))
	}

	// Check that all expected filesystems are present
	expectedFS := []string{"fat32", "exfat", "none"}
	for _, expected := range expectedFS {
		found := false
		for _, filesystem := range fs {
			if filesystem == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected filesystem %s not found in valid filesystems", expected)
		}
	}
}
