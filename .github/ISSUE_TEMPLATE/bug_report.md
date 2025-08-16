---
name: Bug report
about: Create a report to help us improve
title: '[BUG] '
labels: ['bug']
assignees: ''
---

## 🐛 Bug Description

A clear and concise description of what the bug is.

## 🔄 Steps to Reproduce

1. Run command `...`
2. Create device with `...`
3. See error

## ✅ Expected Behavior

A clear and concise description of what you expected to happen.

## ❌ Actual Behavior

A clear and concise description of what actually happened.

## 📋 Environment

- **OS**: [e.g., Ubuntu 22.04, Fedora 38]
- **Kernel**: [e.g., 5.15.0-88-generic]
- **virtusb version**: [e.g., 1.0.0]
- **Go version**: [e.g., 1.22.0]

## 🔍 Diagnostic Output

```bash
# Run this command and paste the output
virtusb diagnose
```

## 📝 Additional Information

Add any other context about the problem here, including:
- Error messages
- Log files
- Screenshots
- Related issues

## 🧪 Reproduction Steps

```bash
# Commands to reproduce the issue
virtusb create test-device --size 8G --brand sandisk
virtusb enable test-device
# ... more commands
```
