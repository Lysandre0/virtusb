# virtusb 🚀

**Virtual USB Gadget Manager for Linux** - Create and manage virtual USB storage devices using the Linux USB Gadget framework.

[![Go Version](https://img.shields.io/badge/Go-1.22+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Build Status](https://github.com/your-username/virtusb/workflows/Test/badge.svg)](https://github.com/your-username/virtusb/actions)
[![Release](https://img.shields.io/github/v/release/your-username/virtusb)](https://github.com/your-username/virtusb/releases)

## 🎯 What is virtusb?

virtusb is a powerful command-line tool that allows you to create and manage virtual USB storage devices on Linux systems. It leverages the Linux USB Gadget framework to create realistic USB storage devices that appear as genuine hardware to the system.

### ✨ Key Features

- 🔧 **Easy USB Device Creation** - Create virtual USB storage devices with custom sizes and brands
- 🎭 **Multiple Brand Support** - SanDisk, Kingston, Corsair, Samsung, and Generic devices
- 💾 **Flexible Storage Options** - FAT32, exFAT, or raw storage
- 🔄 **Automatic Restoration** - Restore devices after system reboot
- 🧪 **Mock Mode** - Test without system modifications
- ⚡ **High Performance** - Optimized with intelligent caching
- 🔍 **System Diagnostics** - Comprehensive system health checks

## 🚀 Quick Start

### Prerequisites

- Linux kernel with USB Gadget support
- Root privileges (required for USB operations)
- Kernel modules: `libcomposite`, `dummy_hcd`, `usbip_core`, `usbip_host`

### Installation

```bash
# Clone the repository
git clone https://github.com/your-username/virtusb.git
cd virtusb

# Build the project
make build

# Install (requires sudo)
sudo make install
```

### Basic Usage

```bash
# Create a virtual USB device
sudo virtusb create my-device --size 8G --brand sandisk

# List all devices
sudo virtusb list

# Enable a device
sudo virtusb enable my-device

# Check system status
sudo virtusb diagnose
```

## 📖 Usage Guide

### Creating Virtual USB Devices

```bash
# Basic device creation
virtusb create my-gadget --size 8G --brand sandisk

# Custom filesystem
virtusb create exfat-device --size 16G --brand kingston --fs exfat

# Custom serial number
virtusb create custom-device --size 4G --brand corsair --serial MY_SERIAL_123

# Raw storage (no filesystem)
virtusb create raw-device --size 2G --fs none
```

### Managing Devices

```bash
# List all devices
virtusb list

# Enable a device (makes it visible to the system)
virtusb enable my-gadget

# Disable a device
virtusb disable my-gadget

# Delete a device (preserves the image file)
virtusb delete my-gadget

# Restore all devices after reboot
virtusb restore
```

### System Diagnostics

```bash
# Comprehensive system check
virtusb diagnose
```

This command checks:
- Root privileges
- Kernel modules status
- configfs mounting
- UDC availability
- USB/IP binaries
- Directory permissions

### Environment Variables

```bash
# Mock mode (for testing)
export MOCK=1

# Custom paths
export VIRTUSB_ROOT=/sys/kernel/config/usb_gadget
export VIRTUSB_STATE_DIR=/etc/virtusb
export VIRTUSB_IMAGE_DIR=/var/lib/virtusb

# Custom binaries
export USBIP_BIN=/usr/bin/usbip
export USBIPD_BIN=/usr/bin/usbipd

# Default values
export VIRTUSB_DEFAULT_SIZE=8G
export VIRTUSB_DEFAULT_BRAND=sandisk
export VIRTUSB_DEFAULT_FS=fat32
```

## 🏗️ Architecture

```
virtusb/
├── cmd/virtusb/          # Main application entry point
├── internal/
│   ├── cli/             # Command-line interface
│   ├── config/          # Configuration management
│   ├── core/
│   │   ├── gadget/      # USB gadget management
│   │   └── storage/     # Storage image management
│   ├── platform/        # Platform abstraction layer
│   └── utils/           # Utility functions
└── .github/workflows/   # CI/CD pipelines
```

### Supported Brands

| Brand | VID | PID | Description |
|-------|-----|-----|-------------|
| SanDisk | 0781 | 5567 | SanDisk USB devices |
| Kingston | 0951 | 1666 | Kingston USB devices |
| Corsair | 1b1c | 1a0a | Corsair USB devices |
| Samsung | 04e8 | 61b6 | Samsung USB devices |
| Generic | 13fe | 4200 | Generic USB devices |

### Supported Filesystems

- **FAT32** - Universal compatibility
- **exFAT** - Large file support
- **none** - Raw storage (no formatting)

### Supported Sizes

- **64M, 128M, 256M, 512M**
- **1G, 2G, 4G, 8G, 16G, 32G, 64G**

## 🧪 Testing

### Unit Tests

```bash
# Run all tests
make test

# Run tests with coverage
go test -v -coverprofile=coverage.out ./...
```

### Mock Mode Testing

```bash
# Test without system modifications
MOCK=1 virtusb create test-device --size 64M
MOCK=1 virtusb list
MOCK=1 virtusb diagnose
```

## 🔧 Development

### Building

```bash
# Standard build
make build

# Clean build artifacts
make clean

# Format code
make fmt

# Lint code
make vet
```

### Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 🐛 Troubleshooting

### Common Issues

#### "must be run as root"
```bash
# Solution: Run with sudo
sudo virtusb create my-device
```

#### "no UDC available"
```bash
# Check kernel modules
lsmod | grep -E "(libcomposite|dummy_hcd)"

# Load modules manually
sudo modprobe libcomposite
sudo modprobe dummy_hcd
```

#### "configfs is not mounted"
```bash
# Mount configfs
sudo mount -t configfs none /sys/kernel/config

# Verify mounting
mount | grep configfs
```

#### "usbip binary not found"
```bash
# Install usbip (Ubuntu/Debian)
sudo apt install usbip

# Install usbip (Fedora/RHEL)
sudo dnf install usbip

# Install usbip (Arch)
sudo pacman -S usbip
```

### Diagnostic Commands

```bash
# Full system diagnostic
virtusb diagnose

# Check kernel modules
lsmod | grep -E "(libcomposite|dummy_hcd|usbip)"

# Check configfs
mount | grep configfs

# Check UDC availability
ls /sys/class/udc/
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

### Development Setup

```bash
# Clone and setup
git clone https://github.com/your-username/virtusb.git
cd virtusb

# Install dependencies
go mod download

# Run tests
make test

# Build
make build
```

## 📊 Performance

virtusb is optimized for performance with:
- **Intelligent caching** for gadget metadata
- **Optimized file operations** with fallocate/truncate
- **Concurrent operations** with proper locking
- **Memory-efficient** data structures

## 🔗 Related Projects

- [Linux USB Gadget Framework](https://www.kernel.org/doc/html/latest/driver-api/usb/gadget.html)
- [USB/IP Project](http://usbip.sourceforge.net/)
- [ConfigFS Documentation](https://www.kernel.org/doc/html/latest/filesystems/configfs.html)

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/your-username/virtusb/issues)
- **Discussions**: [GitHub Discussions](https://github.com/your-username/virtusb/discussions)
- **Documentation**: [Wiki](https://github.com/your-username/virtusb/wiki)

---

**Made with ❤️ for the Linux community**

[![GitHub stars](https://img.shields.io/github/stars/your-username/virtusb?style=social)](https://github.com/your-username/virtusb/stargazers)
[![GitHub forks](https://img.shields.io/github/forks/your-username/virtusb?style=social)](https://github.com/your-username/virtusb/network)
[![GitHub issues](https://img.shields.io/github/issues/your-username/virtusb)](https://github.com/your-username/virtusb/issues)
