package usbip

// USBIPManager defines the interface for managing USB/IP
type USBIPManager interface {
	ExportGadget(name string) error
	UnexportGadget(busid string) error
	ListExported() ([]string, error)
}

// Manager implements USBIPManager
type Manager struct{}

// NewManager creates a new USB/IP manager
func NewManager() *Manager {
	return &Manager{}
}

// ExportGadget exports a gadget via USB/IP
func (m *Manager) ExportGadget(name string) error {
	return nil
}

// UnexportGadget unexports a gadget
func (m *Manager) UnexportGadget(busid string) error {
	return nil
}

// ListExported lists exported gadgets
func (m *Manager) ListExported() ([]string, error) {
	return []string{}, nil
}
