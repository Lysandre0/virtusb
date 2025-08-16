# virtusb 🚀

**Virtual USB Gadget Manager for Linux** - Create and manage virtual USB storage devices.

[![Go Version](https://img.shields.io/badge/Go-1.22+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-GPL%20v3-green.svg)](LICENSE)

## What is virtusb?

virtusb creates virtual USB storage devices on Linux using the USB Gadget framework. Perfect for testing, development, and creating virtual storage devices.

## Quick Start

### Install
```bash
git clone https://github.com/your-username/virtusb.git
cd virtusb
make build
sudo make install
```

### Use
```bash
# Create a virtual USB device
sudo virtusb create my-device --size 8G --brand sandisk

# List devices
sudo virtusb list

# Enable device
sudo virtusb enable my-device

# Delete device
sudo virtusb delete my-device
```

## Commands

| Command | Description |
|---------|-------------|
| `create <name> --size <size> --brand <brand>` | Create virtual USB device |
| `list` | List all devices |
| `enable <name>` | Enable device |
| `disable <name>` | Disable device |
| `delete <name> [--keep-image]` | Delete device (remove image by default) |
| `restore` | Restore all devices after reboot |
| `diagnose` | Check system status |

## Supported Options

### Brands
- `sandisk`, `kingston`, `corsair`, `samsung`, `generic`

### Sizes
- `64M`, `128M`, `256M`, `512M`, `1G`, `2G`, `4G`, `8G`, `16G`, `32G`, `64G`

### Filesystems
- `fat32` (default), `exfat`, `none`

## Examples

```bash
# Create different types of devices
virtusb create usb1 --size 4G --brand kingston
virtusb create usb2 --size 16G --brand corsair --fs exfat
virtusb create usb3 --size 2G --fs none

# Manage devices
virtusb enable usb1
virtusb list
virtusb disable usb1
virtusb delete usb1

# Keep image when deleting
virtusb delete usb2 --keep-image
```

## Development

```bash
# Build
make build

# Test
make test

# Build for Linux
make build-linux
```

## Requirements

- Linux kernel with USB Gadget support
- Root privileges
- Kernel modules: `libcomposite`, `dummy_hcd`

## License

GNU General Public License v3 - see [LICENSE](LICENSE) file.
