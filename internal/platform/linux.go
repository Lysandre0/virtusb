package platform

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// LinuxPlatform implements Platform for Linux
type LinuxPlatform struct {
	CommonPlatform
}

func (p *LinuxPlatform) RequireRoot() error {
	if os.Geteuid() != 0 {
		return errors.New("must be run as root (sudo)")
	}
	return nil
}

func (p *LinuxPlatform) EnsureEnvironment(ctx context.Context) error {
	// Create persistent directories
	if err := p.CreateDirectory(p.config.StateDir); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}
	if err := p.CreateDirectory(p.config.ImageDir); err != nil {
		return fmt.Errorf("failed to create image directory: %w", err)
	}

	// Mount configfs if necessary
	if !p.IsMountpoint("/sys/kernel/config") {
		if err := p.MountConfigFS(); err != nil {
			return fmt.Errorf("failed to mount configfs: %w", err)
		}
	}

	// Load kernel modules
	modules := []string{"libcomposite", "dummy_hcd", "usbip_core", "usbip_host"}
	for _, module := range modules {
		if err := p.LoadModule(module); err != nil {
			// Warning only, not fatal error
			fmt.Printf("Warning: failed to load module %s: %v\n", module, err)
		}
	}

	return nil
}

func (p *LinuxPlatform) GetFirstUDC() (string, error) {
	entries, err := os.ReadDir("/sys/class/udc")
	if err != nil || len(entries) == 0 {
		return "", errors.New("no UDC available")
	}
	return entries[0].Name(), nil
}

func (p *LinuxPlatform) IsUDCAvailable() bool {
	_, err := p.GetFirstUDC()
	return err == nil
}

func (p *LinuxPlatform) LoadModule(name string) error {
	return p.RunCommandQuiet("modprobe", name)
}

func (p *LinuxPlatform) IsModuleLoaded(name string) bool {
	f, err := os.Open("/proc/modules")
	if err != nil {
		return false
	}
	defer f.Close()

	// Simple read to check if module is loaded
	data, err := os.ReadFile("/proc/modules")
	if err != nil {
		return false
	}
	return strings.Contains(string(data), name+" ")
}

func (p *LinuxPlatform) IsMountpoint(path string) bool {
	f, err := os.Open("/proc/self/mounts")
	if err != nil {
		return false
	}
	defer f.Close()

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
	path, _ := exec.LookPath(binary)
	return path
}

func (p *LinuxPlatform) RunCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %s: %v: %s", name, strings.Join(args, " "), err, stderr.String())
	}
	return nil
}

func (p *LinuxPlatform) RunCommandQuiet(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	return cmd.Run()
}

func (p *LinuxPlatform) WriteString(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o644)
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
	return os.MkdirAll(path, 0o755)
}
