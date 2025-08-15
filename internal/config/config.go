package config

import (
	"os"
	"strings"
)

// Config contains all application configuration
type Config struct {
	// Execution mode
	MockMode bool

	// System paths
	GadgetRoot string
	StateDir   string
	ImageDir   string

	// External binaries
	USBIPBin  string
	USBIPDBin string

	// Gadget configuration
	DefaultSize  string
	DefaultBrand string
	DefaultFS    string

	// USB/IP configuration
	USBIPPort int
}

// LoadFromEnv loads configuration from environment variables
func LoadFromEnv() *Config {
	config := &Config{
		// Execution mode
		MockMode: envOn("MOCK"),

		// System paths
		GadgetRoot: getEnv("VIRTUSB_ROOT", "/sys/kernel/config/usb_gadget"),
		StateDir:   getEnv("VIRTUSB_STATE_DIR", "/etc/virtusb"),
		ImageDir:   getEnv("VIRTUSB_IMAGE_DIR", "/var/lib/virtusb"),

		// External binaries
		USBIPBin:  getEnv("USBIP_BIN", "usbip"),
		USBIPDBin: getEnv("USBIPD_BIN", "usbipd"),

		// Gadget configuration
		DefaultSize:  getEnv("VIRTUSB_DEFAULT_SIZE", "8G"),
		DefaultBrand: getEnv("VIRTUSB_DEFAULT_BRAND", "sandisk"),
		DefaultFS:    getEnv("VIRTUSB_DEFAULT_FS", "fat32"),

		// USB/IP configuration
		USBIPPort: 3240, // Default port
	}

	if config.GadgetRoot == "/sys/kernel/config/usb_gadget" {
		if alt := os.Getenv("USBVIRT_ROOT"); alt != "" {
			config.GadgetRoot = alt
		}
	}

	return config
}

// getEnv retrieves an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// envOn checks if an environment variable is enabled
func envOn(key string) bool {
	value := strings.ToLower(os.Getenv(key))
	return value == "1" || value == "true" || value == "yes"
}

// Validate validates the configuration
func (c *Config) Validate() error {
	// Basic checks
	if c.GadgetRoot == "" {
		return ErrInvalidConfig("GadgetRoot cannot be empty")
	}
	if c.StateDir == "" {
		return ErrInvalidConfig("StateDir cannot be empty")
	}
	if c.ImageDir == "" {
		return ErrInvalidConfig("ImageDir cannot be empty")
	}

	// Default value checks
	if !isValidSize(c.DefaultSize) {
		return ErrInvalidConfig("DefaultSize must be a valid size (e.g., 64M, 2G, 8G)")
	}
	if !isValidBrand(c.DefaultBrand) {
		return ErrInvalidConfig("DefaultBrand must be a valid brand")
	}
	if !isValidFS(c.DefaultFS) {
		return ErrInvalidConfig("DefaultFS must be a valid filesystem type")
	}

	return nil
}

// isValidSize checks if a size is valid
func isValidSize(size string) bool {
	validSizes := []string{"64M", "128M", "256M", "512M", "1G", "2G", "4G", "8G", "16G", "32G", "64G"}
	for _, valid := range validSizes {
		if size == valid {
			return true
		}
	}
	return false
}

// isValidBrand checks if a brand is valid
func isValidBrand(brand string) bool {
	validBrands := []string{"sandisk", "kingston", "corsair", "samsung", "generic"}
	for _, valid := range validBrands {
		if strings.ToLower(brand) == valid {
			return true
		}
	}
	return false
}

// isValidFS checks if a filesystem is valid
func isValidFS(fs string) bool {
	validFS := []string{"fat32", "exfat", "none"}
	for _, valid := range validFS {
		if strings.ToLower(fs) == valid {
			return true
		}
	}
	return false
}

// ErrInvalidConfig represents a configuration error
type ErrInvalidConfig string

func (e ErrInvalidConfig) Error() string {
	return string(e)
}
