# Release Guide

## How to create a release

### 1. Prepare the release

```bash
# Ensure everything works
make clean
make test-mock
make test

# Check that code is ready
git status
git diff
```

### 2. Create a tag

```bash
# Create a version tag
git tag v0.1.0

# Push the tag (automatically triggers the release)
git push origin v0.1.0
```

### 3. Verify the release

The release will be automatically created on GitHub with:
- ✅ Binaries for Linux (AMD64, ARM64)
- ✅ Binaries for macOS (AMD64, ARM64)
- ✅ Automatic installation script
- ✅ Usage instructions

## Installation for users

### Quick installation
```bash
curl -sSL https://github.com/[your-username]/virtusb/releases/latest/download/install.sh | bash
```

### Manual installation
1. Go to the [releases page](https://github.com/[your-username]/virtusb/releases)
2. Download the appropriate binary
3. Execute:
```bash
chmod +x virtusb_*
sudo mv virtusb_* /usr/local/bin/virtusb
```

## Development workflow

```bash
# Daily development
make build
make test-mock

# Before a release
make clean
make test
git tag v0.1.0
git push origin v0.1.0
```

## Release system advantages

- ✅ **Simplified installation** for users
- ✅ **Multi-platform** (Linux, macOS)
- ✅ **Multi-architecture** (AMD64, ARM64)
- ✅ **Complete automation**
- ✅ **Smart installation script**
- ✅ **Documentation** included in each release
