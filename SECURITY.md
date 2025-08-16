# Security Policy 🔒

## Supported Versions

We release patches for security vulnerabilities. Which versions are eligible for receiving such patches depends on the CVSS v3.0 Rating:

| Version | Supported          |
| ------- | ------------------ |
| 1.0.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

We take the security of virtusb seriously. If you believe you have found a security vulnerability, please report it to us as described below.

### Reporting Process

1. **Do not create a public GitHub issue** for the vulnerability
2. **Email us** at security@yourdomain.com with the subject line `[SECURITY] virtusb vulnerability report`
3. **Include detailed information** about the vulnerability:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if any)

### What to Expect

- **Acknowledgment**: You will receive an acknowledgment within 48 hours
- **Assessment**: We will assess the vulnerability and determine its severity
- **Timeline**: We will provide a timeline for fixing the issue
- **Updates**: You will receive regular updates on the progress

### Responsible Disclosure

We follow responsible disclosure practices:
- **No public disclosure** until a fix is available
- **Credit acknowledgment** for security researchers
- **Coordinated disclosure** with affected parties

## Security Considerations

### Root Privileges

virtusb requires root privileges to:
- Access USB gadget framework
- Load kernel modules
- Mount configfs
- Create device files

**Security Best Practices:**
- Only run virtusb when necessary
- Use mock mode for testing
- Review created devices regularly
- Keep system updated

### File System Access

virtusb creates and manages:
- Storage image files
- Configuration files
- Device metadata

**Security Considerations:**
- Images are stored in `/var/lib/virtusb/`
- Metadata in `/etc/virtusb/`
- Ensure proper file permissions
- Regular security audits

### Network Exposure

virtusb may interact with:
- USB/IP functionality (future)
- Network device sharing

**Security Guidelines:**
- Use firewall rules appropriately
- Monitor network access
- Validate device sharing permissions

## Security Updates

### Automatic Updates

- **Dependencies**: Keep Go and system dependencies updated
- **Kernel modules**: Update kernel modules regularly
- **System packages**: Maintain updated system packages

### Manual Updates

```bash
# Update virtusb
git pull origin main
make build
sudo make install

# Verify installation
virtusb version
virtusb diagnose
```

## Security Contacts

- **Security Email**: security@yourdomain.com
- **PGP Key**: [Available on request]
- **Response Time**: 48 hours for acknowledgment

## Security Acknowledgments

We thank security researchers who responsibly disclose vulnerabilities to us. Contributors will be acknowledged in:
- Release notes
- Security advisories
- Project documentation

---

**Thank you for helping keep virtusb secure!** 🛡️
