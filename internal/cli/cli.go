package cli

import (
	"fmt"
	"strings"
	"sync"

	"virtusb/internal/config"
	"virtusb/internal/core/gadget"
	"virtusb/internal/core/storage"
	"virtusb/internal/platform"
)

type CLI struct {
	config         *config.Config
	platform       platform.Platform
	gadgetManager  gadget.GadgetManager
	storageManager storage.StorageManager
	validCommands  map[string]bool
	commandMutex   sync.RWMutex
}

func NewCLI() (*CLI, error) {
	cfg := config.LoadFromEnv()
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

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
		return nil, fmt.Errorf("platform initialization failed: %w", err)
	}

	storageManager := storage.NewManager()
	gadgetManager := gadget.NewManager(storageManager)

	cli := &CLI{
		config:         cfg,
		platform:       platform,
		gadgetManager:  gadgetManager,
		storageManager: storageManager,
		validCommands:  make(map[string]bool, 16),
	}

	cli.initValidCommands()
	return cli, nil
}

func (cli *CLI) initValidCommands() {
	cli.commandMutex.Lock()
	defer cli.commandMutex.Unlock()

	commands := []string{
		"help", "-h", "--help",
		"version", "-v", "--version",
		"create", "list", "enable", "disable",
		"delete", "restore", "diagnose",
	}

	for _, cmd := range commands {
		cli.validCommands[cmd] = true
	}
}

func (cli *CLI) Run(args []string) error {
	if len(args) < 1 {
		return cli.showUsage()
	}

	command := args[0]
	commandArgs := args[1:]

	if !cli.isValidCommand(command) {
		return fmt.Errorf("unknown command: %s. Use 'virtusb help' for usage", command)
	}

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
	case "restore":
		return cli.handleRestore(commandArgs)
	case "diagnose":
		return cli.handleDiagnose(commandArgs)
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

func (cli *CLI) isValidCommand(command string) bool {
	cli.commandMutex.RLock()
	defer cli.commandMutex.RUnlock()
	return cli.validCommands[command]
}

func (cli *CLI) showUsage() error {
	fmt.Println(`virtusb — Virtual USB Gadget Manager (Linux USB Gadget + USB/IP)

USAGE
  virtusb create <name> [--size 8G] [--brand sandisk|kingston|corsair|samsung|generic]
                       [--serial SERIAL] [--fs fat32|exfat|none]
  virtusb enable <name>
  virtusb disable <name>
  virtusb list
  virtusb delete <name>
  virtusb restore
  virtusb diagnose

ENVIRONMENT VARIABLES
  MOCK=1                 mock mode (no system modifications)
  VIRTUSB_ROOT=...       gadget root directory (default: /sys/kernel/config/usb_gadget)
  USBVIRT_ROOT=...       (alias) same as VIRTUSB_ROOT
  USBIP_BIN / USBIPD_BIN custom usbip binaries

EXAMPLES
  virtusb create my-gadget --size 8G --brand sandisk
  virtusb enable my-gadget
  virtusb list
  virtusb diagnose`)
	return nil
}

func (cli *CLI) showVersion() error {
	fmt.Println("virtusb version 1.0.0")
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
		return fmt.Errorf("usage: virtusb create <name> [--size 8G] [--brand sandisk|kingston|corsair|samsung|generic] [--serial SERIAL] [--fs fat32|exfat|none]")
	}

	name := args[0]
	if err := cli.validateGadgetName(name); err != nil {
		return fmt.Errorf("invalid gadget name: %w", err)
	}

	opts := cli.parseCreateOptions(args[1:])
	opts.Name = name

	ctx := cli.createContext()
	gadget, err := cli.gadgetManager.Create(ctx, opts)
	if err != nil {
		return fmt.Errorf("gadget creation failed: %w", err)
	}

	fmt.Printf("✅ Gadget '%s' created successfully\n", gadget.Name)
	fmt.Printf("   VID:PID=%s:%s  SN=%s  IMG=%s\n",
		gadget.VID, gadget.PID, gadget.Serial, gadget.ImagePath)
	return nil
}

func (cli *CLI) handleList(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("usage: virtusb list")
	}

	ctx := cli.createContext()
	gadgets, err := cli.gadgetManager.List(ctx, gadget.ListOptions{})
	if err != nil {
		return fmt.Errorf("gadget listing failed: %w", err)
	}

	if len(gadgets) == 0 {
		fmt.Println("No gadgets found. Use 'virtusb create <name>' to create one.")
		return nil
	}

	fmt.Printf("%-20s %-8s %-12s %-20s %s\n", "NAME", "ENABLED", "VID:PID", "BRAND", "IMAGE")
	fmt.Println(strings.Repeat("-", 80))

	for _, g := range gadgets {
		enabled := "❌"
		if g.Enabled {
			enabled = "✅"
		}
		fmt.Printf("%-20s %-8s %s:%s %-20s %s\n",
			g.Name, enabled, g.VID, g.PID, g.Brand, g.ImagePath)
	}

	return nil
}

func (cli *CLI) handleEnable(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: virtusb enable <name>")
	}

	name := args[0]
	if err := cli.validateGadgetName(name); err != nil {
		return fmt.Errorf("invalid gadget name: %w", err)
	}

	ctx := cli.createContext()

	if err := cli.gadgetManager.Enable(ctx, name); err != nil {
		return fmt.Errorf("gadget enable failed: %w", err)
	}

	fmt.Printf("✅ Gadget '%s' enabled successfully\n", name)
	return nil
}

func (cli *CLI) handleDisable(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: virtusb disable <name>")
	}

	name := args[0]
	if err := cli.validateGadgetName(name); err != nil {
		return fmt.Errorf("invalid gadget name: %w", err)
	}

	ctx := cli.createContext()

	if err := cli.gadgetManager.Disable(ctx, name); err != nil {
		return fmt.Errorf("gadget disable failed: %w", err)
	}

	fmt.Printf("✅ Gadget '%s' disabled successfully\n", name)
	return nil
}

func (cli *CLI) handleDelete(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: virtusb delete <name>")
	}

	name := args[0]
	if err := cli.validateGadgetName(name); err != nil {
		return fmt.Errorf("invalid gadget name: %w", err)
	}

	ctx := cli.createContext()

	if err := cli.gadgetManager.Delete(ctx, name); err != nil {
		return fmt.Errorf("gadget deletion failed: %w", err)
	}

	fmt.Printf("✅ Gadget '%s' deleted successfully (image preserved)\n", name)
	return nil
}

func (cli *CLI) handleRestore(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("usage: virtusb restore")
	}

	ctx := cli.createContext()
	if err := cli.gadgetManager.Restore(ctx); err != nil {
		return fmt.Errorf("gadget restoration failed: %w", err)
	}

	fmt.Println("✅ Gadget restoration completed successfully")
	return nil
}

func (cli *CLI) handleDiagnose(args []string) error {
	if len(args) > 0 {
		return fmt.Errorf("usage: virtusb diagnose")
	}

	return cli.runDiagnostic()
}

func (cli *CLI) validateGadgetName(name string) error {
	if name == "" {
		return fmt.Errorf("gadget name required")
	}

	if len(name) < 1 {
		return fmt.Errorf("gadget name too short (min 1 character)")
	}

	if len(name) > 50 {
		return fmt.Errorf("name too long (maximum 50 characters)")
	}

	// Check for leading/trailing whitespace
	if strings.TrimSpace(name) != name {
		return fmt.Errorf("gadget name cannot have leading or trailing whitespace")
	}

	// Check for control characters
	for _, char := range name {
		if char < 32 || char == 127 {
			return fmt.Errorf("gadget name cannot contain control characters")
		}
	}

	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", ".."}
	for _, char := range invalidChars {
		if strings.Contains(name, char) {
			return fmt.Errorf("name cannot contain '%s'", char)
		}
	}

	// Check for reserved names
	reservedNames := []string{".", "..", "null", "zero", "random", "urandom"}
	for _, reserved := range reservedNames {
		if strings.ToLower(name) == reserved {
			return fmt.Errorf("gadget name cannot be reserved name: %s", reserved)
		}
	}

	return nil
}

func (cli *CLI) parseCreateOptions(args []string) gadget.CreateOptions {
	opts := gadget.CreateOptions{
		Size:       cli.config.DefaultSize,
		Brand:      gadget.Brand(cli.config.DefaultBrand),
		FileSystem: gadget.FileSystem(cli.config.DefaultFS),
	}

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--size":
			if i+1 < len(args) {
				opts.Size = args[i+1]
				i++
			}
		case "--brand":
			if i+1 < len(args) {
				opts.Brand = gadget.Brand(args[i+1])
				i++
			}
		case "--serial":
			if i+1 < len(args) {
				opts.Serial = args[i+1]
				i++
			}
		case "--fs":
			if i+1 < len(args) {
				opts.FileSystem = gadget.FileSystem(args[i+1])
				i++
			}
		}
	}

	return opts
}

func (cli *CLI) runDiagnostic() error {
	fmt.Println("🔍 virtusb diagnostic")
	fmt.Println(strings.Repeat("=", 50))

	if err := cli.platform.RequireRoot(); err != nil {
		fmt.Println("❌ Not running as root (required for USB gadget operations)")
	} else {
		fmt.Println("✅ Running as root")
	}

	if cli.platform.IsMountpoint("/sys/kernel/config") {
		fmt.Println("✅ configfs is mounted")
	} else {
		fmt.Println("❌ configfs is not mounted")
	}

	modules := []string{"libcomposite", "dummy_hcd", "usbip_core", "usbip_host"}
	for _, module := range modules {
		if cli.platform.IsModuleLoaded(module) {
			fmt.Printf("✅ Module %s loaded\n", module)
		} else {
			fmt.Printf("❌ Module %s not loaded\n", module)
		}
	}

	if udc, err := cli.platform.GetFirstUDC(); err != nil {
		fmt.Printf("❌ No UDC found: %v\n", err)
	} else {
		fmt.Printf("✅ UDC found: %s\n", udc)
	}

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

	dirs := []string{cli.platform.GetStateDir(), cli.platform.GetImageDir(), cli.platform.GetGadgetRoot()}
	for _, dir := range dirs {
		if cli.platform.FileExists(dir) {
			fmt.Printf("✅ Directory %s exists\n", dir)
		} else {
			fmt.Printf("❌ Directory %s missing\n", dir)
		}
	}

	fmt.Println(strings.Repeat("=", 50))
	return nil
}
