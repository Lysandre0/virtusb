package storage

import (
	"fmt"
	"sync"
)

type StorageManager interface {
	CreateImage(ctx Context, path, size, fsType string) error
	DeleteImage(ctx Context, path string) error
	ImageExists(ctx Context, path string) bool
}

type Context struct {
	Platform Platform
}

type Platform interface {
	IsMockMode() bool
	RunCommand(name string, args ...string) error
	RunCommandQuiet(name string, args ...string) error
	FileExists(path string) bool
}

type Manager struct {
	sizeCache map[string]bool
	sizeMutex sync.RWMutex
}

func NewManager() *Manager {
	return &Manager{
		sizeCache: make(map[string]bool, 16),
	}
}

func (m *Manager) CreateImage(ctx Context, path, size, fsType string) error {
	if ctx.Platform.IsMockMode() {
		return m.createMockImage(ctx, path)
	}

	if err := m.validateSizeCached(size); err != nil {
		return fmt.Errorf("invalid size format: %w", err)
	}

	if err := m.createImageFileOptimized(ctx, path, size); err != nil {
		return fmt.Errorf("image file creation failed: %w", err)
	}

	if err := m.formatFilesystem(ctx, path, fsType); err != nil {
		return fmt.Errorf("filesystem formatting failed: %w", err)
	}

	return nil
}

func (m *Manager) DeleteImage(ctx Context, path string) error {
	if !ctx.Platform.FileExists(path) {
		return nil
	}

	m.unmountImage(ctx, path)

	if err := ctx.Platform.RunCommand("rm", "-f", path); err != nil {
		return fmt.Errorf("image deletion failed: %w", err)
	}

	return nil
}

func (m *Manager) ImageExists(ctx Context, path string) bool {
	return ctx.Platform.FileExists(path)
}

func (m *Manager) validateSizeCached(size string) error {
	m.sizeMutex.RLock()
	if valid, exists := m.sizeCache[size]; exists {
		m.sizeMutex.RUnlock()
		if !valid {
			return fmt.Errorf("invalid size format: %s", size)
		}
		return nil
	}
	m.sizeMutex.RUnlock()

	err := m.validateSize(size)
	valid := err == nil

	m.sizeMutex.Lock()
	if len(m.sizeCache) >= 50 {
		for k := range m.sizeCache {
			delete(m.sizeCache, k)
			break
		}
	}
	m.sizeCache[size] = valid
	m.sizeMutex.Unlock()

	return err
}

func (m *Manager) createImageFileOptimized(ctx Context, path, size string) error {
	if err := ctx.Platform.RunCommand("fallocate", "-l", size, path); err == nil {
		return nil
	}

	if err := ctx.Platform.RunCommand("truncate", "-s", size, path); err == nil {
		return nil
	}

	if err := ctx.Platform.RunCommand("dd", "if=/dev/zero", "of="+path, "bs=1M", "count=64"); err != nil {
		return fmt.Errorf("failed to create image file with fallocate, truncate, and dd: %w", err)
	}

	return nil
}

func (m *Manager) validateSize(size string) error {
	if len(size) < 2 {
		return fmt.Errorf("size too short")
	}

	unit := size[len(size)-1]
	if unit != 'M' && unit != 'G' {
		return fmt.Errorf("invalid size unit, must be M or G")
	}

	return nil
}

func (m *Manager) formatFilesystem(ctx Context, path, fsType string) error {
	switch fsType {
	case "none":
		return nil
	case "exfat":
		if err := ctx.Platform.RunCommand("mkfs.exfat", path); err == nil {
			return nil
		}
		fallthrough
	default:
		if err := ctx.Platform.RunCommand("mkfs.vfat", "-F", "32", path); err != nil {
			return fmt.Errorf("FAT32 filesystem creation failed: %w", err)
		}
		return nil
	}
}

func (m *Manager) createMockImage(ctx Context, path string) error {
	return ctx.Platform.RunCommand("dd", "if=/dev/zero", "of="+path, "bs=1K", "count=1")
}

func (m *Manager) unmountImage(ctx Context, path string) {
	ctx.Platform.RunCommandQuiet("umount", path)
}
