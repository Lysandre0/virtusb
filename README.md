# virtusb

Virtual USB gadget manager for Linux using the Linux USB Gadget framework and USB/IP.

## Features

- Create virtual USB keys with different brands (SanDisk, Kingston, etc.)
- Support for different filesystems (FAT32, exFAT)
- Export via USB/IP for remote access
- Configuration persistence
- Mock mode for testing

## Prerequisites

- Linux with kernel 4.0+
- Root privileges (sudo)
- Kernel modules: `libcomposite`, `dummy_hcd`, `usbip_core`, `usbip_host`
- Tools: `usbip`, `dosfstools`, `exfat-utils`

## Installation

### Quick Install (Recommended)

```bash
# Automatic installation
curl -sSL https://github.com/[your-username]/virtusb/releases/latest/download/install.sh | bash
```

### Manual Installation

```bash
# Build from source
go build -o virtusb cmd/virtusb/main.go

# Install
sudo cp virtusb /usr/local/bin/
```

### Direct Download

Download the appropriate binary from [GitHub releases](https://github.com/[your-username]/virtusb/releases):

- **Linux AMD64**: `virtusb_linux_amd64`
- **Linux ARM64**: `virtusb_linux_arm64`
- **macOS AMD64**: `virtusb_darwin_amd64`
- **macOS ARM64**: `virtusb_darwin_arm64`

Then:
```bash
chmod +x virtusb_*
sudo mv virtusb_* /usr/local/bin/virtusb
```

## Usage

```bash
# Create a virtual USB key
sudo virtusb create my_key --size 8G --brand sandisk

# List gadgets
sudo virtusb list

# Enable a gadget
sudo virtusb enable my_key

# Export via USB/IP
sudo virtusb export my_key

# Diagnostic
sudo virtusb diagnose
```

## Environment Variables

- `MOCK=1` : Mock mode (no system changes)
- `VIRTUSB_ROOT` : Gadget root directory
- `USBIP_BIN` / `USBIPD_BIN` : Paths to usbip binaries

## Architecture

The project follows a modular architecture with separation of concerns:

- `internal/config/` : Centralized configuration
- `internal/core/` : Business logic (gadget, storage, usbip)
- `internal/platform/` : Platform abstraction
- `internal/cli/` : Command line interface
- `internal/utils/` : Generic utilities

## License

[Insert your license here]
