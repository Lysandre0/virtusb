# Changelog 📝

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Nothing yet

### Changed
- Nothing yet

### Deprecated
- Nothing yet

### Removed
- Nothing yet

### Fixed
- Nothing yet

### Security
- Nothing yet

## [1.0.0] - 2025-08-16

### Added
- 🚀 **Initial release** of virtusb
- 🔧 **USB gadget creation** with custom sizes and brands
- 🎭 **Multiple brand support** (SanDisk, Kingston, Corsair, Samsung, Generic)
- 💾 **Filesystem support** (FAT32, exFAT, none)
- 🔄 **Automatic device restoration** after system reboot
- 🧪 **Mock mode** for testing without system modifications
- 🔍 **System diagnostics** with comprehensive health checks
- ⚡ **Performance optimizations** with intelligent caching
- 🏗️ **Modular architecture** with clear interfaces
- 📦 **Cross-platform builds** (Linux amd64/arm64, macOS amd64/arm64)
- 🔒 **Security features** with proper error handling
- 📚 **Comprehensive documentation** and examples
- 🧪 **Test suite** with 148 unit tests
- 🔄 **CI/CD pipelines** with GitHub Actions
- 📋 **Issue templates** and contribution guidelines

### Features
- **CLI Interface**: Complete command-line interface with help and version commands
- **Device Management**: Create, list, enable, disable, and delete virtual USB devices
- **Storage Management**: Create and manage storage images with various filesystems
- **Platform Abstraction**: Support for Linux and mock platforms
- **Configuration Management**: Environment-based configuration with validation
- **Error Handling**: Custom error types with meaningful messages
- **Caching System**: Intelligent caching for performance optimization

### Technical Details
- **Go 1.22+** compatibility
- **MIT License** for open source use
- **Modular design** with clear separation of concerns
- **Thread-safe operations** with proper locking
- **Memory-efficient** data structures
- **Optimized builds** with size and performance flags

### Documentation
- **README.md**: Comprehensive project documentation
- **CONTRIBUTING.md**: Contribution guidelines
- **SECURITY.md**: Security policy and best practices
- **CHANGELOG.md**: Version history and changes
- **Issue Templates**: Standardized bug reports and feature requests
- **Pull Request Template**: Guidelines for code contributions

### Infrastructure
- **GitHub Actions**: Automated testing, linting, and releases
- **Cross-compilation**: Support for multiple platforms
- **Release automation**: Automatic binary distribution
- **Code quality**: Linting, formatting, and vetting
- **Test coverage**: Comprehensive test suite

---

## Version History

### Version 1.0.0
- **Release Date**: August 16, 2025
- **Status**: Initial stable release
- **Features**: Complete USB gadget management system
- **Platforms**: Linux (amd64/arm64), macOS (amd64/arm64)

---

## Contributing to the Changelog

When adding entries to the changelog, please follow these guidelines:

1. **Use clear, descriptive language**
2. **Categorize changes** appropriately
3. **Include issue numbers** when relevant
4. **Add emojis** for visual appeal
5. **Keep entries concise** but informative

### Categories

- **Added**: New features
- **Changed**: Changes in existing functionality
- **Deprecated**: Soon-to-be removed features
- **Removed**: Removed features
- **Fixed**: Bug fixes
- **Security**: Security-related changes
