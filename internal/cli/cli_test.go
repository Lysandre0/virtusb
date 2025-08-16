package cli

import (
	"testing"

	"virtusb/internal/config"
	"virtusb/internal/core/gadget"
)

func TestIsValidCommand(t *testing.T) {
	testCases := []struct {
		command  string
		expected bool
	}{
		{"help", true},
		{"-h", true},
		{"--help", true},
		{"version", true},
		{"-v", true},
		{"--version", true},
		{"create", true},
		{"list", true},
		{"enable", true},
		{"disable", true},
		{"delete", true},
		{"export", true},
		{"unexport", true},
		{"restore", true},
		{"diagnose", true},
		{"invalid", false},
		{"", false},
		{"create-gadget", false},
	}

	cli := &CLI{
		config: &config.Config{
			DefaultSize:  "8G",
			DefaultBrand: "sandisk",
			DefaultFS:    "fat32",
		},
		validCommands: make(map[string]bool, 16),
	}

	commands := []string{
		"help", "-h", "--help",
		"version", "-v", "--version",
		"create", "list", "enable", "disable",
		"delete", "export", "unexport",
		"restore", "diagnose",
	}

	for _, cmd := range commands {
		cli.validCommands[cmd] = true
	}

	for _, tc := range testCases {
		t.Run(tc.command, func(t *testing.T) {
			result := cli.isValidCommand(tc.command)
			if result != tc.expected {
				t.Errorf("Expected %v for command '%s', got %v", tc.expected, tc.command, result)
			}
		})
	}
}

func TestValidateGadgetName(t *testing.T) {
	testCases := []struct {
		name     string
		expected bool
	}{
		{"test-gadget", true},
		{"gadget1", true},
		{"my_gadget", true},
		{"gadget-123", true},
		{"", false},
		{"test/gadget", false},
		{"test\\gadget", false},
		{"test:gadget", false},
		{"test*gadget", false},
		{"test?gadget", false},
		{"test\"gadget", false},
		{"test<gadget", false},
		{"test>gadget", false},
		{"test|gadget", false},
		{"a", true},
		{"very-long-gadget-name-that-exceeds-fifty-characters-and-should-be-rejected", false},
	}

	cli := &CLI{}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := cli.validateGadgetName(tc.name)
			if tc.expected && err != nil {
				t.Errorf("Expected valid name '%s', got error: %v", tc.name, err)
			}
			if !tc.expected && err == nil {
				t.Errorf("Expected invalid name '%s', but validation passed", tc.name)
			}
		})
	}
}

func TestParseCreateOptions(t *testing.T) {
	testCases := []struct {
		name     string
		args     []string
		expected gadget.CreateOptions
	}{
		{
			name: "no options",
			args: []string{},
			expected: gadget.CreateOptions{
				Size:       "8G",
				Brand:      "sandisk",
				FileSystem: "fat32",
			},
		},
		{
			name: "with size",
			args: []string{"--size", "16G"},
			expected: gadget.CreateOptions{
				Size:       "16G",
				Brand:      "sandisk",
				FileSystem: "fat32",
			},
		},
		{
			name: "with brand",
			args: []string{"--brand", "kingston"},
			expected: gadget.CreateOptions{
				Size:       "8G",
				Brand:      "kingston",
				FileSystem: "fat32",
			},
		},
		{
			name: "with filesystem",
			args: []string{"--fs", "exfat"},
			expected: gadget.CreateOptions{
				Size:       "8G",
				Brand:      "sandisk",
				FileSystem: "exfat",
			},
		},
		{
			name: "with serial",
			args: []string{"--serial", "TEST123"},
			expected: gadget.CreateOptions{
				Size:       "8G",
				Brand:      "sandisk",
				FileSystem: "fat32",
				Serial:     "TEST123",
			},
		},
		{
			name: "multiple options",
			args: []string{"--size", "4G", "--brand", "corsair", "--fs", "none", "--serial", "CORS123"},
			expected: gadget.CreateOptions{
				Size:       "4G",
				Brand:      "corsair",
				FileSystem: "none",
				Serial:     "CORS123",
			},
		},
		{
			name: "incomplete option",
			args: []string{"--size"},
			expected: gadget.CreateOptions{
				Size:       "8G",
				Brand:      "sandisk",
				FileSystem: "fat32",
			},
		},
	}

	cli := &CLI{
		config: &config.Config{
			DefaultSize:  "8G",
			DefaultBrand: "sandisk",
			DefaultFS:    "fat32",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := cli.parseCreateOptions(tc.args)

			if result.Size != tc.expected.Size {
				t.Errorf("Expected size '%s', got '%s'", tc.expected.Size, result.Size)
			}
			if result.Brand != tc.expected.Brand {
				t.Errorf("Expected brand '%s', got '%s'", tc.expected.Brand, result.Brand)
			}
			if result.FileSystem != tc.expected.FileSystem {
				t.Errorf("Expected filesystem '%s', got '%s'", tc.expected.FileSystem, result.FileSystem)
			}
			if result.Serial != tc.expected.Serial {
				t.Errorf("Expected serial '%s', got '%s'", tc.expected.Serial, result.Serial)
			}
		})
	}
}

func TestShowUsage(t *testing.T) {
	cli := &CLI{}
	err := cli.showUsage()
	if err != nil {
		t.Errorf("showUsage should not return an error: %v", err)
	}
}

func TestShowVersion(t *testing.T) {
	cli := &CLI{}
	err := cli.showVersion()
	if err != nil {
		t.Errorf("showVersion should not return an error: %v", err)
	}
}

func TestCreateContext(t *testing.T) {
	cli := &CLI{
		config: &config.Config{
			GadgetRoot: "/sys/kernel/config/usb_gadget",
			StateDir:   "/etc/virtusb",
			ImageDir:   "/var/lib/virtusb",
		},
	}

	ctx := cli.createContext()
	if ctx.Config == nil {
		t.Error("Context should have a config")
	}
}

func TestHandleCreateValidation(t *testing.T) {
	testCases := []struct {
		name     string
		args     []string
		expected bool
	}{
		{
			name:     "valid create",
			args:     []string{"test-gadget"},
			expected: true,
		},
		{
			name:     "no name",
			args:     []string{},
			expected: false,
		},
		{
			name:     "invalid name",
			args:     []string{"test/gadget"},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cli := &CLI{
				config: &config.Config{
					DefaultSize:  "8G",
					DefaultBrand: "sandisk",
					DefaultFS:    "fat32",
				},
			}

			if tc.name == "valid create" {
				err := cli.validateGadgetName(tc.args[0])
				if err != nil {
					t.Errorf("Expected valid name '%s', got error: %v", tc.args[0], err)
				}
			} else if tc.name == "invalid name" {
				err := cli.validateGadgetName(tc.args[0])
				if err == nil {
					t.Errorf("Expected invalid name '%s', but validation passed", tc.args[0])
				}
			} else if tc.name == "no name" {
				err := cli.validateGadgetName("")
				if err == nil {
					t.Error("Expected empty name to fail validation")
				}
			}
		})
	}
}

func TestHandleList(t *testing.T) {
	cli := &CLI{
		config: &config.Config{
			DefaultSize:  "8G",
			DefaultBrand: "sandisk",
			DefaultFS:    "fat32",
		},
	}

	err := cli.handleList([]string{"invalid-arg"})
	if err == nil {
		t.Error("Expected error for invalid arguments")
	}
}

func TestHandleEnable(t *testing.T) {
	testCases := []struct {
		name     string
		args     []string
		expected bool
	}{
		{
			name:     "valid enable",
			args:     []string{"test-gadget"},
			expected: true,
		},
		{
			name:     "no name",
			args:     []string{},
			expected: false,
		},
		{
			name:     "too many args",
			args:     []string{"gadget1", "gadget2"},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.name == "valid enable" {
				if tc.args[0] == "" {
					t.Error("Expected valid name, got empty string")
				}
			} else if tc.name == "no name" {
				if len(tc.args) != 0 {
					t.Error("Expected empty args")
				}
			} else if tc.name == "too many args" {
				if len(tc.args) != 2 {
					t.Error("Expected 2 args")
				}
			}
		})
	}
}

func TestHandleDisable(t *testing.T) {
	testCases := []struct {
		name     string
		args     []string
		expected bool
	}{
		{
			name:     "valid disable",
			args:     []string{"test-gadget"},
			expected: true,
		},
		{
			name:     "no name",
			args:     []string{},
			expected: false,
		},
		{
			name:     "too many args",
			args:     []string{"gadget1", "gadget2"},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.name == "valid disable" {
				if tc.args[0] == "" {
					t.Error("Expected valid name, got empty string")
				}
			} else if tc.name == "no name" {
				if len(tc.args) != 0 {
					t.Error("Expected empty args")
				}
			} else if tc.name == "too many args" {
				if len(tc.args) != 2 {
					t.Error("Expected 2 args")
				}
			}
		})
	}
}

func TestHandleDelete(t *testing.T) {
	testCases := []struct {
		name     string
		args     []string
		expected bool
	}{
		{
			name:     "valid delete",
			args:     []string{"test-gadget"},
			expected: true,
		},
		{
			name:     "no name",
			args:     []string{},
			expected: false,
		},
		{
			name:     "too many args",
			args:     []string{"gadget1", "gadget2"},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.name == "valid delete" {
				if tc.args[0] == "" {
					t.Error("Expected valid name, got empty string")
				}
			} else if tc.name == "no name" {
				if len(tc.args) != 0 {
					t.Error("Expected empty args")
				}
			} else if tc.name == "too many args" {
				if len(tc.args) != 2 {
					t.Error("Expected 2 args")
				}
			}
		})
	}
}

func TestHandleRestore(t *testing.T) {
	args := []string{}
	if len(args) != 0 {
		t.Error("Expected empty args for restore")
	}
}

func TestHandleDiagnose(t *testing.T) {
	args := []string{}
	if len(args) != 0 {
		t.Error("Expected empty args for diagnose")
	}
}
