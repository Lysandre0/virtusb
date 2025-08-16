package platform

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type LinuxPlatform struct {
	CommonPlatform
	moduleCache  map[string]bool
	mountCache   map[string]bool
	cacheMutex   sync.RWMutex
	udcCache     []string
	udcCacheTime time.Time
	udcMutex     sync.RWMutex
}

func (p *LinuxPlatform) initLinuxPlatform() {
	if p.moduleCache == nil {
		p.moduleCache = make(map[string]bool, 16)
	}
	if p.mountCache == nil {
		p.mountCache = make(map[string]bool, 8)
	}
}

func (p *LinuxPlatform) RequireRoot() error {
	if os.Geteuid() != 0 {
		return errors.New("must be run as root")
	}
	return nil
}

func (p *LinuxPlatform) EnsureEnvironment(ctx context.Context) error {
	modules := []string{"libcomposite", "dummy_hcd", "usbip_core", "usbip_host"}
	for _, module := range modules {
		if !p.IsModuleLoaded(module) {
			if err := p.LoadModule(module); err != nil {
				return fmt.Errorf("failed to load module %s: %w", module, err)
			}
		}
	}

	if !p.IsMountpoint("/sys/kernel/config") {
		if err := p.MountConfigFS(); err != nil {
			return fmt.Errorf("failed to mount configfs: %w", err)
		}
	}

	return nil
}

func (p *LinuxPlatform) GetFirstUDC() (string, error) {
	p.udcMutex.RLock()
	if len(p.udcCache) > 0 && time.Since(p.udcCacheTime) < 5*time.Second {
		udc := p.udcCache[0]
		p.udcMutex.RUnlock()
		return udc, nil
	}
	p.udcMutex.RUnlock()

	entries, err := os.ReadDir("/sys/class/udc")
	if err != nil {
		return "", fmt.Errorf("failed to read UDC directory: %w", err)
	}

	var udcs []string
	for _, entry := range entries {
		if entry.IsDir() {
			udcs = append(udcs, entry.Name())
		}
	}

	if len(udcs) == 0 {
		return "", errors.New("no UDC available")
	}

	p.udcMutex.Lock()
	p.udcCache = udcs
	p.udcCacheTime = time.Now()
	p.udcMutex.Unlock()

	return udcs[0], nil
}

func (p *LinuxPlatform) IsUDCAvailable() bool {
	_, err := p.GetFirstUDC()
	return err == nil
}

func (p *LinuxPlatform) LoadModule(name string) error {
	return p.RunCommand("modprobe", name)
}

func (p *LinuxPlatform) IsModuleLoaded(name string) bool {
	p.cacheMutex.RLock()
	if p.moduleCache == nil {
		p.cacheMutex.RUnlock()
		p.cacheMutex.Lock()
		defer p.cacheMutex.Unlock()
		if p.moduleCache == nil {
			p.moduleCache = make(map[string]bool)
		}
		// Check cache again after initialization
		if loaded, exists := p.moduleCache[name]; exists {
			return loaded
		}
	} else if loaded, exists := p.moduleCache[name]; exists {
		p.cacheMutex.RUnlock()
		return loaded
	} else {
		p.cacheMutex.RUnlock()
	}

	loaded := p.checkModuleLoaded(name)

	p.cacheMutex.Lock()
	if p.moduleCache == nil {
		p.moduleCache = make(map[string]bool)
	}
	p.moduleCache[name] = loaded
	p.cacheMutex.Unlock()

	return loaded
}

func (p *LinuxPlatform) checkModuleLoaded(name string) bool {
	data, err := os.ReadFile("/proc/modules")
	if err != nil {
		return false
	}
	return strings.Contains(string(data), name+" ")
}

func (p *LinuxPlatform) IsMountpoint(path string) bool {
	p.cacheMutex.RLock()
	if p.mountCache == nil {
		p.cacheMutex.RUnlock()
		p.cacheMutex.Lock()
		defer p.cacheMutex.Unlock()
		if p.mountCache == nil {
			p.mountCache = make(map[string]bool)
		}
		// Check cache again after initialization
		if mounted, exists := p.mountCache[path]; exists {
			return mounted
		}
	} else if mounted, exists := p.mountCache[path]; exists {
		p.cacheMutex.RUnlock()
		return mounted
	} else {
		p.cacheMutex.RUnlock()
	}

	mounted := p.checkMountpoint(path)

	p.cacheMutex.Lock()
	if p.mountCache == nil {
		p.mountCache = make(map[string]bool)
	}
	p.mountCache[path] = mounted
	p.cacheMutex.Unlock()

	return mounted
}

func (p *LinuxPlatform) checkMountpoint(path string) bool {
	data, err := os.ReadFile("/proc/self/mounts")
	if err != nil {
		return false
	}
	return strings.Contains(string(data), " "+path+" ")
}

func (p *LinuxPlatform) MountConfigFS() error {
	return p.RunCommand("mount", "-t", "configfs", "none", "/sys/kernel/config")
}

func (p *LinuxPlatform) Which(binary string) string {
	path, err := exec.LookPath(binary)
	if err != nil {
		return ""
	}
	return path
}

func (p *LinuxPlatform) RunCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (p *LinuxPlatform) RunCommandQuiet(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

func (p *LinuxPlatform) WriteString(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

func (p *LinuxPlatform) ReadString(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (p *LinuxPlatform) FileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (p *LinuxPlatform) CreateDirectory(path string) error {
	return os.MkdirAll(path, 0755)
}
