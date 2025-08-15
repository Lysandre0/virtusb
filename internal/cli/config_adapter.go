package cli

import (
	"virtusb/internal/config"
)

// ConfigAdapter adapts config.Config to gadget.Config
type ConfigAdapter struct {
	config *config.Config
}

// NewConfigAdapter creates a new configuration adapter
func NewConfigAdapter(cfg *config.Config) *ConfigAdapter {
	return &ConfigAdapter{config: cfg}
}

// GetGadgetRoot returns the gadget root
func (ca *ConfigAdapter) GetGadgetRoot() string {
	return ca.config.GadgetRoot
}

// GetStateDir returns the state directory
func (ca *ConfigAdapter) GetStateDir() string {
	return ca.config.StateDir
}

// GetImageDir returns the image directory
func (ca *ConfigAdapter) GetImageDir() string {
	return ca.config.ImageDir
}

// GetUSBIPBin returns the usbip binary
func (ca *ConfigAdapter) GetUSBIPBin() string {
	return ca.config.USBIPBin
}

// GetUSBIPDBin returns the usbipd binary
func (ca *ConfigAdapter) GetUSBIPDBin() string {
	return ca.config.USBIPDBin
}

// IsMockMode returns whether mock mode is enabled
func (ca *ConfigAdapter) IsMockMode() bool {
	return ca.config.MockMode
}
