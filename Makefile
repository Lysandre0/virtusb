# virtusb - Virtual USB Device Manager
# Version: 1.0.0 - Simple Installation

.PHONY: help install uninstall test clean build

help: ## Show this help
	@echo "virtusb - Virtual USB Device Manager"
	@echo "Version: 1.0.0 - Package Native"
	@echo ""
	@echo "Available targets:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-15s - %s\n", $$1, $$2}'

install: ## Install virtusb
	@echo "Installing virtusb..."
	@sudo cp src/virtusb.sh /usr/local/bin/virtusb
	@sudo chmod +x /usr/local/bin/virtusb
	@sudo mkdir -p /etc/virtusb
	@sudo mkdir -p /opt/virtusb/data
	@sudo mkdir -p /opt/virtusb/logs
	@echo '[Unit]' | sudo tee /usr/lib/systemd/system/virtusb.service > /dev/null
	@echo 'Description=Virtual USB Gadget Manager' | sudo tee -a /usr/lib/systemd/system/virtusb.service > /dev/null
	@echo 'After=network.target' | sudo tee -a /usr/lib/systemd/system/virtusb.service > /dev/null
	@echo 'Before=multi-user.target' | sudo tee -a /usr/lib/systemd/system/virtusb.service > /dev/null
	@echo '' | sudo tee -a /usr/lib/systemd/system/virtusb.service > /dev/null
	@echo '[Service]' | sudo tee -a /usr/lib/systemd/system/virtusb.service > /dev/null
	@echo 'Type=oneshot' | sudo tee -a /usr/lib/systemd/system/virtusb.service > /dev/null
	@echo 'RemainAfterExit=yes' | sudo tee -a /usr/lib/systemd/system/virtusb.service > /dev/null
	@echo 'ExecStart=/usr/local/bin/virtusb --load-modules' | sudo tee -a /usr/lib/systemd/system/virtusb.service > /dev/null
	@echo 'ExecStartPost=/bin/bash -c "mkdir -p /opt/virtusb/logs && echo virtusb started at $(date) >> /opt/virtusb/logs/service.log"' | sudo tee -a /usr/lib/systemd/system/virtusb.service > /dev/null
	@echo 'ExecStop=/bin/true' | sudo tee -a /usr/lib/systemd/system/virtusb.service > /dev/null
	@echo 'Restart=on-failure' | sudo tee -a /usr/lib/systemd/system/virtusb.service > /dev/null
	@echo 'RestartSec=10' | sudo tee -a /usr/lib/systemd/system/virtusb.service > /dev/null
	@echo '' | sudo tee -a /usr/lib/systemd/system/virtusb.service > /dev/null
	@echo '[Install]' | sudo tee -a /usr/lib/systemd/system/virtusb.service > /dev/null
	@echo 'WantedBy=multi-user.target' | sudo tee -a /usr/lib/systemd/system/virtusb.service > /dev/null
	@sudo systemctl daemon-reload
	@sudo systemctl enable virtusb.service
	@sudo systemctl start virtusb.service
	@echo "✅ virtusb installed and service started"
	@echo "Usage: sudo virtusb help"

uninstall: ## Remove virtusb
	@echo "Removing virtusb..."
	@sudo systemctl stop virtusb.service 2>/dev/null || true
	@sudo systemctl disable virtusb.service 2>/dev/null || true
	@sudo rm -f /usr/local/bin/virtusb
	@sudo rm -f /usr/lib/systemd/system/virtusb.service
	@sudo systemctl daemon-reload
	@echo "✅ virtusb removed"

test: ## Test virtusb installation
	@echo "Testing virtusb installation..."
	@sudo virtusb help
	@echo "✅ Test completed"

clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	@rm -rf virtusb-1.0.0/
	@rm -rf rpmbuild/
	@rm -rf pkg/
	@rm -f *.deb *.rpm *.pkg.tar.zst
	@rm -f virtusb.service
	@echo "✅ Clean completed"

build: ## Build systemd service (for production)
	@echo "Building systemd service..."
	@mkdir -p /usr/lib/systemd/system
	@sudo cp virtusb.service /usr/lib/systemd/system/ 2>/dev/null || echo "Service file not found, will be created during installation"
	@echo "✅ Service ready"