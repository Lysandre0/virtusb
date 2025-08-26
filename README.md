# virtusb - Virtual USB Device Manager

The simple and reliable solution for creating virtual USB drives on Linux systems hosting virtual machines.

## ✨ Features

- **Universal Linux support** - Works on Proxmox, VMware ESXi, KVM/QEMU, VirtualBox, and any Linux system
- **Simple installation** - Easy setup via Makefile
- **Persistent installation** - Survives system updates and reboots
- **Automatic module loading** - Loads required kernel modules automatically
- **Multiple UDC support** - Creates up to 30 USB Device Controllers for maximum flexibility
- **Realistic device simulation** - Supports popular USB drive brands with authentic VID:PID pairs
- **Simple management** - Easy create, enable, disable, and delete operations
- **Clean interface** - Minimal, user-friendly output with status indicators
- **Automatic state persistence** - Seamlessly restores enabled devices after system reboot
- **Robust error handling** - Automatically cleans up orphaned devices and metadata

## 🚀 Installation

```bash
git clone https://github.com/user/virtusb.git
cd virtusb
make install
```

## 📖 Basic Usage

```bash
# Create a virtual USB drive
sudo virtusb create mykey --size 1G --brand sandisk

# Enable the device
sudo virtusb enable mykey

# List all devices
sudo virtusb list

# Disable a device
sudo virtusb disable mykey

# Delete a device
sudo virtusb delete mykey
```

## 🔄 State Persistence

virtusb automatically saves and restores the state of enabled devices across system reboots. The process is completely transparent to the user:

- **Automatic saving**: When you enable a device, its state is automatically saved
- **Automatic restoration**: After system reboot, all previously enabled devices are automatically restored
- **Automatic cleanup**: Orphaned devices and metadata are automatically cleaned up
- **No user intervention required**: Everything happens seamlessly in the background

## �� Commands Reference

| Command | Description |
|---------|-------------|
| `create <name> --size <size> --brand <brand>` | Create a new virtual USB device |
| `enable <name>` | Activate a device |
| `disable <name>` | Deactivate a device |
| `delete <name>` | Remove a device completely |
| `list` | Show all devices with status |
| `purge` | Remove ALL devices |
| `help` | Display help information |

## 📋 Supported Brands

- **sandisk** (VID:PID: 0781:5567)
- **kingston** (VID:PID: 0951:1666)
- **samsung** (VID:PID: 04e8:61f5)
- **toshiba** (VID:PID: 0930:6545)
- **lexar** (VID:PID: 05dc:a4a5)
- **pny** (VID:PID: 154b:0078)
- **verbatim** (VID:PID: 18a5:0243)
- **transcend** (VID:PID: 8564:1000)
- **adata** (VID:PID: 125f:c96a)
- **corsair** (VID:PID: 1b1c:1a0d)

## 🏗️ Installation Structure

```
/usr/local/bin/virtusb              # Main executable
/usr/lib/systemd/system/virtusb.service  # Systemd service
/etc/virtusb/                       # Configuration directory
/opt/virtusb/data/                  # Data directory
/opt/virtusb/logs/                  # Logs directory
```

## 🛠️ Development

```bash
make install    # Install for development
make test       # Test installation
make uninstall  # Remove development installation
make clean      # Clean build artifacts
```

## 🔧 Technical Details

### Kernel Modules
- `libcomposite` - USB Composite Framework
- `dummy_hcd` - Virtual USB Host Controller (loaded with 30 UDC instances)
- `usb_f_mass_storage` - Mass Storage Function

### System Requirements
- Linux kernel with configfs support
- Root privileges for USB gadget operations
- At least 100MB free disk space per device

## ⚠️ Important Notes

### Multiple Device Support
The `dummy_hcd` module is configured with **30 UDC instances** (`num=30`), allowing you to have **up to 30 virtual USB devices active simultaneously**. Each device uses a separate UDC (dummy_udc.0, dummy_udc.1, etc.).

**Example**: You can enable multiple devices at once:
```bash
sudo virtusb create key1 --size 1G --brand sandisk
sudo virtusb create key2 --size 512M --brand kingston
sudo virtusb enable key1
sudo virtusb enable key2
# Both devices are now active and visible in lsusb
```

### Installation Management
- **Easy installation** via `make install`
- **Clean uninstallation** with `make uninstall`
- **Service management** via systemd (`sudo systemctl status virtusb`)

## 🐛 Troubleshooting

### "No UDC available"
- The system automatically loads required modules
- Check if configfs is mounted: `mount | grep configfs`
- Restart the service: `sudo systemctl restart virtusb`

### "Permission denied"
- Run commands with `sudo`
- Ensure you have root privileges

### Device not detected in VM
- Verify the device is enabled: `sudo virtusb list`
- Check if the VID:PID appears in `lsusb`
- Ensure your VM manager supports USB passthrough

### Orphaned devices or metadata
- The system automatically cleans up orphaned devices and metadata
- Use `sudo virtusb purge` to manually clean all devices if needed

## 📝 Examples

```bash
# Create a 2GB Sandisk drive
sudo virtusb create backup --size 2G --brand sandisk
sudo virtusb enable backup

# Create a 512MB Kingston drive
sudo virtusb create small --size 512M --brand kingston
sudo virtusb enable small

# Both devices can be active simultaneously
sudo virtusb list
lsusb | grep -E "(0781:5567|0951:1666)"

# Switch between devices (optional)
sudo virtusb disable backup
sudo virtusb enable new_device
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit pull requests or open issues for bugs and feature requests.

## 📄 License

This project is licensed under the GNU General Public License v3.0 (GPL-3.0)