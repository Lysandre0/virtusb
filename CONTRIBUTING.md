# Contributing to virtusb 🤝

Thank you for your interest in contributing to virtusb! This document provides guidelines and information for contributors.

## 🎯 How to Contribute

### Reporting Issues

Before creating an issue, please:

1. **Search existing issues** to avoid duplicates
2. **Check the documentation** for common solutions
3. **Run diagnostics** with `virtusb diagnose`
4. **Provide detailed information** including:
   - Operating system and version
   - Kernel version
   - virtusb version
   - Complete error messages
   - Steps to reproduce

### Feature Requests

When requesting features:

1. **Describe the use case** clearly
2. **Explain the benefits** for the community
3. **Consider implementation complexity**
4. **Check if it aligns** with project goals

### Code Contributions

#### Development Setup

```bash
# Fork and clone the repository
git clone https://github.com/your-username/virtusb.git
cd virtusb

# Install dependencies
go mod download

# Run tests
make test

# Build the project
make build
```

#### Code Style Guidelines

- **Follow Go conventions** and use `gofmt`
- **Write clear comments** for complex logic
- **Add tests** for new functionality
- **Update documentation** when needed
- **Use meaningful commit messages**

#### Testing

```bash
# Run all tests
make test

# Run tests with coverage
go test -v -coverprofile=coverage.out ./...

# Test in mock mode
MOCK=1 ./build/virtusb diagnose
```

#### Pull Request Process

1. **Create a feature branch** from `main`
2. **Make your changes** following the style guide
3. **Add tests** for new functionality
4. **Update documentation** if needed
5. **Run the test suite** and ensure it passes
6. **Submit a pull request** with a clear description

### Pull Request Guidelines

#### Title and Description

- **Use clear, descriptive titles**
- **Explain the problem** and solution
- **Include any relevant issue numbers**
- **Describe testing performed**

#### Code Review

- **Address review comments** promptly
- **Keep commits focused** and logical
- **Squash commits** if requested
- **Update documentation** as needed

## 🏗️ Development Guidelines

### Architecture Principles

- **Modular design** with clear interfaces
- **Platform abstraction** for testability
- **Error handling** with meaningful messages
- **Performance optimization** where appropriate

### Code Organization

```
internal/
├── cli/          # Command-line interface
├── config/       # Configuration management
├── core/         # Core business logic
│   ├── gadget/   # USB gadget management
│   └── storage/  # Storage management
├── platform/     # Platform abstraction
└── utils/        # Utility functions
```

### Testing Strategy

- **Unit tests** for all packages
- **Integration tests** for complex workflows
- **Mock mode** for testing without system access
- **Performance benchmarks** for critical paths

### Error Handling

- **Use custom error types** for specific errors
- **Provide context** in error messages
- **Handle errors gracefully** at boundaries
- **Log errors** appropriately

## 📋 Issue Templates

### Bug Report Template

```markdown
## Bug Description
Brief description of the issue

## Steps to Reproduce
1. Step one
2. Step two
3. Step three

## Expected Behavior
What should happen

## Actual Behavior
What actually happens

## Environment
- OS: [e.g., Ubuntu 22.04]
- Kernel: [e.g., 5.15.0]
- virtusb version: [e.g., 1.0.0]

## Additional Information
Any other relevant information
```

### Feature Request Template

```markdown
## Feature Description
Brief description of the requested feature

## Use Case
How this feature would be used

## Proposed Implementation
Any ideas for implementation

## Alternatives Considered
Other approaches that were considered

## Additional Context
Any other relevant information
```

## 🎖️ Recognition

Contributors will be recognized in:

- **README.md** contributors section
- **Release notes** for significant contributions
- **GitHub contributors** page

## 📞 Getting Help

If you need help with contributions:

- **Open a discussion** on GitHub
- **Ask questions** in issues
- **Review existing code** for examples
- **Check the documentation**

## 📄 License

By contributing to virtusb, you agree that your contributions will be licensed under the MIT License.

---

Thank you for contributing to virtusb! 🚀
