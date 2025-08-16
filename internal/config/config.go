package config

import (
	"fmt"
	"os"
	"strings"
	"sync"
)

const (
	DefaultGadgetRoot = "/sys/kernel/config/usb_gadget"
	DefaultStateDir   = "/etc/virtusb"
	DefaultImageDir   = "/var/lib/virtusb"
	DefaultUSBIPBin   = "usbip"
	DefaultUSBIPDBin  = "usbipd"
	DefaultSize       = "8G"
	DefaultBrand      = "sandisk"
	DefaultFS         = "fat32"
	DefaultUSBIPPort  = 3240

	Size64M  = "64M"
	Size128M = "128M"
	Size256M = "256M"
	Size512M = "512M"
	Size1G   = "1G"
	Size2G   = "2G"
	Size4G   = "4G"
	Size8G   = "8G"
	Size16G  = "16G"
	Size32G  = "32G"
	Size64G  = "64G"

	BrandSandisk  = "sandisk"
	BrandKingston = "kingston"
	BrandCorsair  = "corsair"
	BrandSamsung  = "samsung"
	BrandGeneric  = "generic"

	FSFat32 = "fat32"
	FSExFat = "exfat"
	FSNone  = "none"
)

var (
	validSizes  map[string]bool
	validBrands map[string]bool
	validFS     map[string]bool
	initOnce    sync.Once
)

func initValidMaps() {
	validSizes = map[string]bool{
		Size64M: true, Size128M: true, Size256M: true, Size512M: true,
		Size1G: true, Size2G: true, Size4G: true, Size8G: true,
		Size16G: true, Size32G: true, Size64G: true,
	}

	validBrands = map[string]bool{
		BrandSandisk: true, BrandKingston: true, BrandCorsair: true,
		BrandSamsung: true, BrandGeneric: true,
	}

	validFS = map[string]bool{
		FSFat32: true, FSExFat: true, FSNone: true,
	}
}

type Config struct {
	MockMode bool

	GadgetRoot string
	StateDir   string
	ImageDir   string

	USBIPBin  string
	USBIPDBin string

	DefaultSize  string
	DefaultBrand string
	DefaultFS    string

	USBIPPort int
}

func LoadFromEnv() *Config {
	config := &Config{
		MockMode: envOn("MOCK"),

		GadgetRoot: getEnv("VIRTUSB_ROOT", DefaultGadgetRoot),
		StateDir:   getEnv("VIRTUSB_STATE_DIR", DefaultStateDir),
		ImageDir:   getEnv("VIRTUSB_IMAGE_DIR", DefaultImageDir),

		USBIPBin:  getEnv("USBIP_BIN", DefaultUSBIPBin),
		USBIPDBin: getEnv("USBIPD_BIN", DefaultUSBIPDBin),

		DefaultSize:  getEnv("VIRTUSB_DEFAULT_SIZE", DefaultSize),
		DefaultBrand: getEnv("VIRTUSB_DEFAULT_BRAND", DefaultBrand),
		DefaultFS:    getEnv("VIRTUSB_DEFAULT_FS", DefaultFS),

		USBIPPort: DefaultUSBIPPort,
	}

	if config.GadgetRoot == DefaultGadgetRoot {
		if alt := os.Getenv("USBVIRT_ROOT"); alt != "" {
			config.GadgetRoot = alt
		}
	}

	return config
}

func (c *Config) Validate() error {
	if c.GadgetRoot == "" {
		return fmt.Errorf("gadget root path is required")
	}
	if c.StateDir == "" {
		return fmt.Errorf("state directory path is required")
	}
	if c.ImageDir == "" {
		return fmt.Errorf("image directory path is required")
	}

	if !isValidSize(c.DefaultSize) {
		return fmt.Errorf("invalid default size: %s", c.DefaultSize)
	}
	if !isValidBrand(c.DefaultBrand) {
		return fmt.Errorf("invalid default brand: %s", c.DefaultBrand)
	}
	if !isValidFS(c.DefaultFS) {
		return fmt.Errorf("invalid default filesystem: %s", c.DefaultFS)
	}

	return nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func envOn(key string) bool {
	value := strings.ToLower(os.Getenv(key))
	return value == "1" || value == "true" || value == "yes" || value == "on"
}

func isValidSize(size string) bool {
	initOnce.Do(initValidMaps)
	return validSizes[size]
}

func isValidBrand(brand string) bool {
	initOnce.Do(initValidMaps)
	return validBrands[strings.ToLower(brand)]
}

func isValidFS(fs string) bool {
	initOnce.Do(initValidMaps)
	return validFS[strings.ToLower(fs)]
}

func GetValidSizes() []string {
	initOnce.Do(initValidMaps)
	sizes := make([]string, 0, len(validSizes))
	for size := range validSizes {
		sizes = append(sizes, size)
	}
	return sizes
}

func GetValidBrands() []string {
	initOnce.Do(initValidMaps)
	brands := make([]string, 0, len(validBrands))
	for brand := range validBrands {
		brands = append(brands, brand)
	}
	return brands
}

func GetValidFilesystems() []string {
	initOnce.Do(initValidMaps)
	fs := make([]string, 0, len(validFS))
	for filesystem := range validFS {
		fs = append(fs, filesystem)
	}
	return fs
}

func IsValidSize(size string) bool {
	initOnce.Do(initValidMaps)
	return isValidSize(size)
}

func IsValidBrand(brand string) bool {
	initOnce.Do(initValidMaps)
	return isValidBrand(brand)
}

func IsValidFS(fs string) bool {
	initOnce.Do(initValidMaps)
	return isValidFS(fs)
}
