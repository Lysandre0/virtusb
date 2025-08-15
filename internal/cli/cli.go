package cli

import (
	"fmt"

	"virtusb/internal/config"
	"virtusb/internal/core/gadget"
	"virtusb/internal/core/storage"
	"virtusb/internal/core/usbip"
	"virtusb/internal/platform"
)

// CLI represents the command line interface
type CLI struct {
	config         *config.Config
	platform       platform.Platform
	gadgetManager  gadget.GadgetManager
	storageManager storage.StorageManager
	usbipManager   usbip.USBIPManager
}

// NewCLI creates a new CLI instance
func NewCLI() (*CLI, error) {
	// Load configuration
	cfg := config.LoadFromEnv()
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Create platform
	platformConfig := platform.Config{
		MockMode:   cfg.MockMode,
		GadgetRoot: cfg.GadgetRoot,
		StateDir:   cfg.StateDir,
		ImageDir:   cfg.ImageDir,
		USBIPBin:   cfg.USBIPBin,
		USBIPDBin:  cfg.USBIPDBin,
	}

	platform, err := platform.NewPlatform(platformConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create platform: %w", err)
	}

	// Create managers
	storageManager := storage.NewManager()
	gadgetManager := gadget.NewManager(storageManager)
	usbipManager := usbip.NewManager()

	return &CLI{
		config:         cfg,
		platform:       platform,
		gadgetManager:  gadgetManager,
		storageManager: storageManager,
		usbipManager:   usbipManager,
	}, nil
}

// Run executes the CLI command
func (cli *CLI) Run(args []string) error {
	if len(args) < 1 {
		return cli.showUsage()
	}

	command := args[0]
	commandArgs := args[1:]

	switch command {
	case "help", "-h", "--help":
		return cli.showUsage()
	case "version", "-v", "--version":
		return cli.showVersion()
	case "create":
		return cli.handleCreate(commandArgs)
	case "list":
		return cli.handleList(commandArgs)
	case "enable":
		return cli.handleEnable(commandArgs)
	case "disable":
		return cli.handleDisable(commandArgs)
	case "delete":
		return cli.handleDelete(commandArgs)
	case "export":
		return cli.handleExport(commandArgs)
	case "unexport":
		return cli.handleUnexport(commandArgs)
	case "restore":
		return cli.handleRestore(commandArgs)
	case "diagnose":
		return cli.handleDiagnose(commandArgs)
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

// Méthodes de gestion des commandes

func (cli *CLI) showUsage() error {
	fmt.Println(`virtusb — manage virtual USB flash gadgets (Linux USB Gadget + USB/IP)

USAGE
  virtusb create <name> [--size 8G] [--brand sandisk|kingston|corsair|samsung|generic]
                         [--serial SERIAL] [--fs fat32|exfat|none]
  virtusb enable <name>
  virtusb disable <name>
  virtusb export <name>
  virtusb unexport <busid>
  virtusb list
  virtusb delete <name>
  virtusb restore
  virtusb diagnose

ENV
  MOCK=1                 mock mode (no real system changes)
  VIRTUSB_ROOT=...       override gadget root (default /sys/kernel/config/usb_gadget)
  USBVIRT_ROOT=...       (alias) same as VIRTUSB_ROOT
  USBIP_BIN / USBIPD_BIN override usbip binaries`)
	return nil
}

func (cli *CLI) showVersion() error {
	fmt.Println("virtusb dev")
	return nil
}

func (cli *CLI) createContext() gadget.Context {
	configAdapter := NewConfigAdapter(cli.config)
	return gadget.Context{
		Platform: cli.platform,
		Config:   configAdapter,
	}
}

func (cli *CLI) handleCreate(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: virtusb create <name> [options]")
	}

	name := args[0]
	opts := gadget.CreateOptions{
		Name:       name,
		Size:       cli.config.DefaultSize,
		Brand:      gadget.Brand(cli.config.DefaultBrand),
		FileSystem: gadget.FileSystem(cli.config.DefaultFS),
	}

	ctx := cli.createContext()

	gadget, err := cli.gadgetManager.Create(ctx, opts)
	if err != nil {
		return fmt.Errorf("failed to create gadget: %w", err)
	}

	fmt.Printf("Created '%s'  VID:PID=%s:%s  SN=%s  IMG=%s\n",
		gadget.Name, gadget.VID, gadget.PID, gadget.Serial, gadget.ImagePath)
	return nil
}

func (cli *CLI) handleList(args []string) error {
	ctx := cli.createContext()

	gadgets, err := cli.gadgetManager.List(ctx, gadget.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list gadgets: %w", err)
	}

	fmt.Printf("%-14s %-7s %-11s %s\n", "NAME", "ENABLED", "VID:PID", "IMG")
	for _, g := range gadgets {
		enabled := "no"
		if g.Enabled {
			enabled = "yes"
		}
		fmt.Printf("%-14s %-7s %s:%s %s\n", g.Name, enabled, g.VID, g.PID, g.ImagePath)
	}

	if len(gadgets) == 0 {
		fmt.Println("No gadgets found. Use 'virtusb create <name>' to create one.")
	}

	return nil
}

func (cli *CLI) handleEnable(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: virtusb enable <name>")
	}

	ctx := cli.createContext()

	if err := cli.gadgetManager.Enable(ctx, args[0]); err != nil {
		return fmt.Errorf("failed to enable gadget: %w", err)
	}

	fmt.Printf("Enabled '%s'\n", args[0])
	return nil
}

func (cli *CLI) handleDisable(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: virtusb disable <name>")
	}

	ctx := cli.createContext()

	if err := cli.gadgetManager.Disable(ctx, args[0]); err != nil {
		return fmt.Errorf("failed to disable gadget: %w", err)
	}

	fmt.Printf("Disabled '%s'\n", args[0])
	return nil
}

func (cli *CLI) handleDelete(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: virtusb delete <name>")
	}

	ctx := cli.createContext()

	if err := cli.gadgetManager.Delete(ctx, args[0]); err != nil {
		return fmt.Errorf("failed to delete gadget: %w", err)
	}

	fmt.Printf("Deleted gadget '%s' (image kept)\n", args[0])
	return nil
}

func (cli *CLI) handleExport(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: virtusb export <name>")
	}

	fmt.Printf("Export functionality not yet implemented for '%s'\n", args[0])
	return nil
}

func (cli *CLI) handleUnexport(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: virtusb unexport <busid>")
	}

	fmt.Printf("Unexport functionality not yet implemented for '%s'\n", args[0])
	return nil
}

func (cli *CLI) handleRestore(args []string) error {
	ctx := cli.createContext()

	if err := cli.gadgetManager.Restore(ctx); err != nil {
		return fmt.Errorf("failed to restore gadgets: %w", err)
	}

	fmt.Println("Restore completed successfully")
	return nil
}

func (cli *CLI) handleDiagnose(args []string) error {
	return cli.runDiagnostic()
}

func (cli *CLI) runDiagnostic() error {
	fmt.Println("=== virtusb Diagnostic ===")

	// Check privileges
	if err := cli.platform.RequireRoot(); err != nil {
		fmt.Println("❌ Not running as root (needed for USB gadget operations)")
	} else {
		fmt.Println("✅ Running as root")
	}

	// Check configfs
	if cli.platform.IsMountpoint("/sys/kernel/config") {
		fmt.Println("✅ configfs is mounted")
	} else {
		fmt.Println("❌ configfs is not mounted")
	}

	// Check kernel modules
	modules := []string{"libcomposite", "dummy_hcd", "usbip_core", "usbip_host"}
	for _, module := range modules {
		if cli.platform.IsModuleLoaded(module) {
			fmt.Printf("✅ Module %s is loaded\n", module)
		} else {
			fmt.Printf("❌ Module %s is not loaded\n", module)
		}
	}

	// Check UDCs
	if udc, err := cli.platform.GetFirstUDC(); err != nil {
		fmt.Printf("❌ No UDC found: %v\n", err)
	} else {
		fmt.Printf("✅ UDC found: %s\n", udc)
	}

	// Check usbip binaries
	if cli.platform.Which("usbip") != "" {
		fmt.Println("✅ usbip binary found")
	} else {
		fmt.Println("❌ usbip binary not found")
	}

	if cli.platform.Which("usbipd") != "" {
		fmt.Println("✅ usbipd binary found")
	} else {
		fmt.Println("❌ usbipd binary not found")
	}

	// Check directories
	dirs := []string{cli.platform.GetStateDir(), cli.platform.GetImageDir(), cli.platform.GetGadgetRoot()}
	for _, dir := range dirs {
		if cli.platform.FileExists(dir) {
			fmt.Printf("✅ Directory %s exists\n", dir)
		} else {
			fmt.Printf("❌ Directory %s missing\n", dir)
		}
	}

	return nil
}
