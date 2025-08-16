package gadget

import (
	"context"
	"time"
)

// Brand represents a USB device brand
type Brand string

const (
	BrandSanDisk  Brand = "sandisk"
	BrandKingston Brand = "kingston"
	BrandCorsair  Brand = "corsair"
	BrandSamsung  Brand = "samsung"
	BrandGeneric  Brand = "generic"
)

// FileSystem represents a filesystem
type FileSystem string

const (
	FSFAT32 FileSystem = "fat32"
	FSExFAT FileSystem = "exfat"
	FSNone  FileSystem = "none"
)

// Gadget represents a virtual USB gadget
type Gadget struct {
	Name       string     `json:"name"`
	Brand      Brand      `json:"brand"`
	VID        string     `json:"vid"`
	PID        string     `json:"pid"`
	Serial     string     `json:"serial"`
	ImagePath  string     `json:"image_path"`
	FileSystem FileSystem `json:"filesystem"`
	Size       string     `json:"size"`
	Enabled    bool       `json:"enabled"`
	UDC        string     `json:"udc,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// CreateOptions contains options for creating a gadget
type CreateOptions struct {
	Name       string     `json:"name"`
	Size       string     `json:"size"`
	Brand      Brand      `json:"brand"`
	Serial     string     `json:"serial,omitempty"`
	FileSystem FileSystem `json:"filesystem"`
}

// ListOptions contains options for listing gadgets
type ListOptions struct {
	Enabled *bool `json:"enabled,omitempty"`
	Brand   Brand `json:"brand,omitempty"`
}

// GadgetManager defines the interface for managing gadgets
type GadgetManager interface {
	Create(ctx Context, opts CreateOptions) (*Gadget, error)
	Get(ctx Context, name string) (*Gadget, error)
	List(ctx Context, opts ListOptions) ([]*Gadget, error)
	Enable(ctx Context, name string) error
	Disable(ctx Context, name string) error
	Delete(ctx Context, name string, keepImage bool) error
	Restore(ctx Context) error
}

// Context contains execution context
type Context struct {
	Platform Platform
	Config   Config
}

// Platform defines the platform interface (imported from platform)
type Platform interface {
	RequireRoot() error
	EnsureEnvironment(ctx context.Context) error
	GetStateDir() string
	GetImageDir() string
	GetGadgetRoot() string
	GetFirstUDC() (string, error)
	IsUDCAvailable() bool
	LoadModule(name string) error
	IsModuleLoaded(name string) bool
	IsMountpoint(path string) bool
	MountConfigFS() error
	Which(binary string) string
	RunCommand(name string, args ...string) error
	RunCommandQuiet(name string, args ...string) error
	WriteString(path, content string) error
	ReadString(path string) (string, error)
	FileExists(path string) bool
	CreateDirectory(path string) error
	IsMockMode() bool
}

// Config defines the configuration interface (imported from config)
type Config interface {
	GetGadgetRoot() string
	GetStateDir() string
	GetImageDir() string
	GetUSBIPBin() string
	GetUSBIPDBin() string
	IsMockMode() bool
}
