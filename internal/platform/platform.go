package platform

import (
	"context"
)

// Platform defines the interface for system operations
type Platform interface {
	// Privilege management
	RequireRoot() error

	// Environment management
	EnsureEnvironment(ctx context.Context) error
	GetStateDir() string
	GetImageDir() string
	GetGadgetRoot() string

	// UDC management
	GetFirstUDC() (string, error)
	IsUDCAvailable() bool

	// Kernel module management
	LoadModule(name string) error
	IsModuleLoaded(name string) bool

	// Filesystem management
	IsMountpoint(path string) bool
	MountConfigFS() error

	// Binary management
	Which(binary string) string
	RunCommand(name string, args ...string) error
	RunCommandQuiet(name string, args ...string) error

	// File management
	WriteString(path, content string) error
	ReadString(path string) (string, error)
	FileExists(path string) bool
	CreateDirectory(path string) error

	// Mock mode
	IsMockMode() bool
}

// Config contains platform configuration
type Config struct {
	MockMode   bool
	GadgetRoot string
	StateDir   string
	ImageDir   string
	USBIPBin   string
	USBIPDBin  string
}

// NewPlatform creates a new platform instance
func NewPlatform(config Config) (Platform, error) {
	if config.MockMode {
		return NewMockPlatform(config)
	}
	return NewLinuxPlatform(config)
}

// CommonPlatform provides common implementations
type CommonPlatform struct {
	config Config
}

func (p *CommonPlatform) IsMockMode() bool {
	return p.config.MockMode
}

func (p *CommonPlatform) GetStateDir() string {
	if p.config.MockMode {
		return "/tmp/virtusb_test/etc/virtusb"
	}
	return p.config.StateDir
}

func (p *CommonPlatform) GetImageDir() string {
	if p.config.MockMode {
		return "/tmp/virtusb_test/var/lib/virtusb"
	}
	return p.config.ImageDir
}

func (p *CommonPlatform) GetGadgetRoot() string {
	if p.config.MockMode {
		return "/tmp/virtusb_test/sys/kernel/config/usb_gadget"
	}
	return p.config.GadgetRoot
}

// NewMockPlatform creates a mock platform
func NewMockPlatform(config Config) (Platform, error) {
	return &MockPlatform{
		CommonPlatform: CommonPlatform{config: config},
	}, nil
}

// NewLinuxPlatform creates a Linux platform
func NewLinuxPlatform(config Config) (Platform, error) {
	return &LinuxPlatform{
		CommonPlatform: CommonPlatform{config: config},
	}, nil
}
