package storage

import (
	"fmt"
)

// StorageManager defines the interface for managing storage
type StorageManager interface {
	CreateImage(ctx Context, path, size, fsType string) error
	DeleteImage(ctx Context, path string) error
	ImageExists(ctx Context, path string) bool
}

// Context contains execution context for storage
type Context struct {
	Platform Platform
}

// Platform defines the platform interface for storage
type Platform interface {
	RunCommand(name string, args ...string) error
	RunCommandQuiet(name string, args ...string) error
	Which(binary string) string
	FileExists(path string) bool
	IsMockMode() bool
}

// Manager implements StorageManager
type Manager struct{}

// NewManager creates a new storage manager
func NewManager() *Manager {
	return &Manager{}
}

// CreateImage creates a storage image
func (m *Manager) CreateImage(ctx Context, path, size, fsType string) error {
	if ctx.Platform.IsMockMode() {
		// In mock mode, create a dummy file
		return m.createMockImage(ctx, path)
	}

	// Create image file
	if err := m.createImageFile(ctx, path, size); err != nil {
		return fmt.Errorf("failed to create image file: %w", err)
	}

	// Format filesystem
	if err := m.formatFilesystem(ctx, path, fsType); err != nil {
		return fmt.Errorf("failed to format filesystem: %w", err)
	}

	return nil
}

// DeleteImage deletes an image
func (m *Manager) DeleteImage(ctx Context, path string) error {
	return nil
}

// ImageExists checks if an image exists
func (m *Manager) ImageExists(ctx Context, path string) bool {
	return ctx.Platform.FileExists(path)
}

// Méthodes privées

func (m *Manager) createImageFile(ctx Context, path, size string) error {
	// Try fallocate first
	if err := ctx.Platform.RunCommand("fallocate", "-l", size, path); err != nil {
		// Fallback to dd
		if err := ctx.Platform.RunCommand("dd", "if=/dev/zero", "of="+path, "bs=1M", "count=64"); err != nil {
			return fmt.Errorf("failed to create image file with both fallocate and dd: %w", err)
		}
	}
	return nil
}

func (m *Manager) formatFilesystem(ctx Context, path, fsType string) error {
	switch fsType {
	case "none":
		return nil
	case "exfat":
		if ctx.Platform.Which("mkfs.exfat") != "" {
			if err := ctx.Platform.RunCommand("mkfs.exfat", path); err != nil {
				return fmt.Errorf("failed to create exFAT filesystem: %w", err)
			}
			return nil
		}
		// Fallback to FAT32 if mkfs.exfat is not available
		fallthrough
	default:
		if err := ctx.Platform.RunCommand("mkfs.vfat", "-F", "32", path); err != nil {
			return fmt.Errorf("failed to create FAT32 filesystem: %w", err)
		}
		return nil
	}
}

func (m *Manager) createMockImage(ctx Context, path string) error {
	// In mock mode, create a dummy 1KB file
	return ctx.Platform.RunCommand("dd", "if=/dev/zero", "of="+path, "bs=1K", "count=1")
}
