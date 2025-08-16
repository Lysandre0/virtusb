package gadget

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"virtusb/internal/core/storage"
	"virtusb/internal/utils"
)

type Manager struct {
	storageManager  storage.StorageManager
	gadgetCache     map[string]*Gadget
	cacheMutex      sync.RWMutex
	validationCache map[string]bool
	validMutex      sync.RWMutex
}

func NewManager(storageManager storage.StorageManager) *Manager {
	return &Manager{
		storageManager:  storageManager,
		gadgetCache:     make(map[string]*Gadget, 16),
		validationCache: make(map[string]bool, 32),
	}
}

func (m *Manager) Create(ctx Context, opts CreateOptions) (*Gadget, error) {
	if err := ctx.Platform.RequireRoot(); err != nil {
		return nil, fmt.Errorf("privileges required: %w", err)
	}

	if err := m.validateCreateOptions(opts); err != nil {
		return nil, fmt.Errorf("invalid options: %w", err)
	}

	if err := ctx.Platform.EnsureEnvironment(context.Background()); err != nil {
		return nil, fmt.Errorf("environment setup failed: %w", err)
	}

	gadgetPath := m.getGadgetPath(ctx, opts.Name)
	if ctx.Platform.FileExists(gadgetPath) {
		return nil, ErrGadgetAlreadyExists{Name: opts.Name}
	}

	vid, pid := utils.VidPid(string(opts.Brand))
	serial := opts.Serial
	if serial == "" {
		serial = utils.SerialFor(string(opts.Brand))
	}

	imagePath := m.getImagePath(ctx, opts.Name)
	if !ctx.Platform.FileExists(imagePath) {
		storageCtx := storage.Context{
			Platform: ctx.Platform,
		}
		if err := m.storageManager.CreateImage(storageCtx, imagePath, opts.Size, string(opts.FileSystem)); err != nil {
			return nil, fmt.Errorf("image creation failed: %w", err)
		}
	}

	if err := m.createGadgetStructure(ctx, opts.Name, vid, pid, serial, string(opts.Brand), imagePath); err != nil {
		return nil, fmt.Errorf("gadget structure creation failed: %w", err)
	}

	gadget := &Gadget{
		Name:       opts.Name,
		Brand:      opts.Brand,
		VID:        vid,
		PID:        pid,
		Serial:     serial,
		ImagePath:  imagePath,
		FileSystem: opts.FileSystem,
		Size:       opts.Size,
		Enabled:    false,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := m.saveMetadata(ctx, gadget); err != nil {
		return nil, fmt.Errorf("metadata save failed: %w", err)
	}

	m.invalidateCache()
	return gadget, nil
}

func (m *Manager) Get(ctx Context, name string) (*Gadget, error) {
	m.cacheMutex.RLock()
	if gadget, exists := m.gadgetCache[name]; exists {
		m.cacheMutex.RUnlock()
		gadget.Enabled = m.isGadgetEnabledCached(ctx, name)
		return gadget, nil
	}
	m.cacheMutex.RUnlock()

	metadataPath := m.getMetadataPath(ctx, name)
	if !ctx.Platform.FileExists(metadataPath) {
		return nil, ErrGadgetNotFound{Name: name}
	}

	content, err := ctx.Platform.ReadString(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("metadata read failed: %w", err)
	}

	gadget, err := m.parseMetadata(content)
	if err != nil {
		return nil, fmt.Errorf("metadata parse failed: %w", err)
	}

	// Verify that the gadget structure still exists
	gadgetPath := m.getGadgetPath(ctx, name)
	if !ctx.Platform.FileExists(gadgetPath) {
		return nil, ErrGadgetNotFound{Name: name}
	}

	gadget.Enabled = m.isGadgetEnabled(ctx, name)
	m.cacheGadgetOptimized(name, gadget)

	return gadget, nil
}

func (m *Manager) List(ctx Context, opts ListOptions) ([]*Gadget, error) {
	stateDir := ctx.Platform.GetStateDir()
	entries, err := os.ReadDir(stateDir)
	if err != nil {
		return nil, fmt.Errorf("state directory read failed: %w", err)
	}

	var gadgets []*Gadget
	var errors []string

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".env") {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".env")
		gadget, err := m.Get(ctx, name)
		if err != nil {
			// Log the error but continue processing other gadgets
			errors = append(errors, fmt.Sprintf("gadget %s: %v", name, err))
			continue
		}

		if !m.matchesFilters(gadget, opts) {
			continue
		}

		gadgets = append(gadgets, gadget)
	}

	// If there were errors, log them for debugging
	if len(errors) > 0 {
		fmt.Fprintf(os.Stderr, "Warning: Some gadgets could not be loaded: %s\n", strings.Join(errors, "; "))
	}

	return gadgets, nil
}

func (m *Manager) Enable(ctx Context, name string) error {
	if err := ctx.Platform.RequireRoot(); err != nil {
		return fmt.Errorf("privileges required: %w", err)
	}

	gadgetPath := m.getGadgetPath(ctx, name)
	if !ctx.Platform.FileExists(gadgetPath) {
		return ErrGadgetNotFound{Name: name}
	}

	if m.isGadgetEnabled(ctx, name) {
		return ErrGadgetAlreadyEnabled{Name: name}
	}

	if err := m.ensureFunctionsLinked(ctx, name); err != nil {
		return fmt.Errorf("function linking failed: %w", err)
	}

	udc, err := ctx.Platform.GetFirstUDC()
	if err != nil {
		return fmt.Errorf("no UDC available: %w", err)
	}

	udcFile := filepath.Join(gadgetPath, "UDC")
	if err := ctx.Platform.WriteString(udcFile, udc+"\n"); err != nil {
		return fmt.Errorf("UDC write failed: %w", err)
	}

	if !m.isGadgetEnabled(ctx, name) {
		return fmt.Errorf("gadget activation failed for %s with UDC %s", name, udc)
	}

	m.invalidateCache()
	return nil
}

func (m *Manager) Disable(ctx Context, name string) error {
	if err := ctx.Platform.RequireRoot(); err != nil {
		return fmt.Errorf("privileges required: %w", err)
	}

	gadgetPath := m.getGadgetPath(ctx, name)
	if !ctx.Platform.FileExists(gadgetPath) {
		return ErrGadgetNotFound{Name: name}
	}

	if !m.isGadgetEnabled(ctx, name) {
		return ErrGadgetNotEnabled{Name: name}
	}

	udcFile := filepath.Join(gadgetPath, "UDC")
	if err := ctx.Platform.WriteString(udcFile, "\n"); err != nil {
		return fmt.Errorf("UDC clear failed: %w", err)
	}

	m.invalidateCache()
	return nil
}

func (m *Manager) Delete(ctx Context, name string, keepImage bool) error {
	if m.isGadgetEnabled(ctx, name) {
		if err := m.Disable(ctx, name); err != nil {
			return fmt.Errorf("disable before delete failed: %w", err)
		}
	}

	// Get gadget info before deletion to access image path
	gadget, err := m.Get(ctx, name)
	if err != nil {
		return fmt.Errorf("failed to get gadget info: %w", err)
	}

	gadgetPath := m.getGadgetPath(ctx, name)
	if err := os.RemoveAll(gadgetPath); err != nil {
		return fmt.Errorf("gadget directory removal failed: %w", err)
	}

	metadataPath := m.getMetadataPath(ctx, name)
	if err := os.Remove(metadataPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("metadata removal failed: %w", err)
	}

	// Delete the image file unless keepImage is true
	if !keepImage && gadget.ImagePath != "" {
		if err := os.Remove(gadget.ImagePath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("image file removal failed: %w", err)
		}
	}

	m.removeFromCache(name)
	return nil
}

func (m *Manager) Restore(ctx Context) error {
	if err := ctx.Platform.RequireRoot(); err != nil {
		return fmt.Errorf("privileges required: %w", err)
	}

	if err := ctx.Platform.EnsureEnvironment(context.Background()); err != nil {
		return fmt.Errorf("environment setup failed: %w", err)
	}

	gadgets, err := m.List(ctx, ListOptions{})
	if err != nil {
		return fmt.Errorf("gadget listing failed: %w", err)
	}

	var errors []string
	for _, gadget := range gadgets {
		if err := m.restoreGadget(ctx, gadget); err != nil {
			errors = append(errors, fmt.Sprintf("gadget %s: %v", gadget.Name, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("restore errors: %s", strings.Join(errors, "; "))
	}

	return nil
}

func (m *Manager) validateCreateOptions(opts CreateOptions) error {
	if opts.Name == "" {
		return fmt.Errorf("gadget name is required")
	}

	if err := m.validateGadgetName(opts.Name); err != nil {
		return err
	}

	if !m.isValidSize(opts.Size) {
		return fmt.Errorf("invalid size: %s", opts.Size)
	}

	if !m.isValidBrand(string(opts.Brand)) {
		return fmt.Errorf("invalid brand: %s", opts.Brand)
	}

	if !m.isValidFS(string(opts.FileSystem)) {
		return fmt.Errorf("invalid filesystem: %s", opts.FileSystem)
	}

	return nil
}

func (m *Manager) validateGadgetName(name string) error {
	if name == "" {
		return fmt.Errorf("gadget name cannot be empty")
	}

	if len(name) < 1 {
		return fmt.Errorf("gadget name too short (min 1 character)")
	}

	if len(name) > 50 {
		return fmt.Errorf("gadget name too long (max 50 characters)")
	}

	// Check for leading/trailing whitespace
	if strings.TrimSpace(name) != name {
		return fmt.Errorf("gadget name cannot have leading or trailing whitespace")
	}

	// Check for control characters
	for _, char := range name {
		if char < 32 || char == 127 {
			return fmt.Errorf("gadget name cannot contain control characters")
		}
	}

	invalidChars := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|", ".."}
	for _, char := range invalidChars {
		if strings.Contains(name, char) {
			return fmt.Errorf("gadget name cannot contain '%s'", char)
		}
	}

	// Check for reserved names
	reservedNames := []string{".", "..", "null", "zero", "random", "urandom"}
	for _, reserved := range reservedNames {
		if strings.ToLower(name) == reserved {
			return fmt.Errorf("gadget name cannot be reserved name: %s", reserved)
		}
	}

	return nil
}

func (m *Manager) matchesFilters(gadget *Gadget, opts ListOptions) bool {
	if opts.Enabled != nil && gadget.Enabled != *opts.Enabled {
		return false
	}
	if opts.Brand != "" && gadget.Brand != opts.Brand {
		return false
	}
	return true
}

func (m *Manager) restoreGadget(ctx Context, gadget *Gadget) error {
	gadgetPath := m.getGadgetPath(ctx, gadget.Name)

	if !ctx.Platform.FileExists(gadgetPath) {
		if err := m.createGadgetStructure(ctx, gadget.Name, gadget.VID, gadget.PID, gadget.Serial, string(gadget.Brand), gadget.ImagePath); err != nil {
			return fmt.Errorf("structure recreation failed: %w", err)
		}
	}

	if err := m.Enable(ctx, gadget.Name); err != nil {
		return fmt.Errorf("enable failed: %w", err)
	}

	return nil
}

func (m *Manager) cacheGadgetOptimized(name string, gadget *Gadget) {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()

	if len(m.gadgetCache) >= 100 {
		for k := range m.gadgetCache {
			delete(m.gadgetCache, k)
			break
		}
	}

	m.gadgetCache[name] = gadget
}

func (m *Manager) isGadgetEnabledCached(ctx Context, name string) bool {
	m.validMutex.RLock()
	if enabled, exists := m.validationCache[name]; exists {
		m.validMutex.RUnlock()
		return enabled
	}
	m.validMutex.RUnlock()

	enabled := m.isGadgetEnabled(ctx, name)

	m.validMutex.Lock()
	if len(m.validationCache) >= 200 {
		for k := range m.validationCache {
			delete(m.validationCache, k)
			break
		}
	}
	m.validationCache[name] = enabled
	m.validMutex.Unlock()

	return enabled
}

func (m *Manager) cacheGadget(name string, gadget *Gadget) {
	m.cacheGadgetOptimized(name, gadget)
}

func (m *Manager) removeFromCache(name string) {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()
	delete(m.gadgetCache, name)
}

func (m *Manager) invalidateCache() {
	m.cacheMutex.Lock()
	defer m.cacheMutex.Unlock()
	m.gadgetCache = make(map[string]*Gadget)
}

func (m *Manager) ClearCache() {
	m.invalidateCache()
}

func (m *Manager) getGadgetPath(ctx Context, name string) string {
	return filepath.Join(ctx.Platform.GetGadgetRoot(), "virtusb-"+name)
}

func (m *Manager) getImagePath(ctx Context, name string) string {
	return filepath.Join(ctx.Platform.GetImageDir(), name+".img")
}

func (m *Manager) getMetadataPath(ctx Context, name string) string {
	return filepath.Join(ctx.Platform.GetStateDir(), name+".env")
}

func (m *Manager) isGadgetEnabled(ctx Context, name string) bool {
	udcFile := filepath.Join(m.getGadgetPath(ctx, name), "UDC")
	content, err := ctx.Platform.ReadString(udcFile)
	if err != nil {
		return false
	}
	return strings.TrimSpace(content) != ""
}

func (m *Manager) ensureFunctionsLinked(ctx Context, name string) error {
	configPath := filepath.Join(m.getGadgetPath(ctx, name), "configs/c.1")
	entries, err := os.ReadDir(configPath)
	if err != nil {
		return fmt.Errorf("config directory read failed: %w", err)
	}

	for _, entry := range entries {
		if entry.Type()&os.ModeSymlink != 0 {
			return nil
		}
	}

	return fmt.Errorf("no functions linked to configs/c.1")
}

func (m *Manager) createGadgetStructure(ctx Context, name, vid, pid, serial, brand, imagePath string) error {
	gadgetPath := m.getGadgetPath(ctx, name)

	dirs := []string{
		filepath.Join(gadgetPath, "strings/0x409"),
		filepath.Join(gadgetPath, "configs/c.1/strings/0x409"),
		filepath.Join(gadgetPath, "functions/mass_storage.0/lun.0"),
	}

	for _, dir := range dirs {
		if err := ctx.Platform.CreateDirectory(dir); err != nil {
			return fmt.Errorf("directory creation failed for %s: %w", dir, err)
		}
	}

	strings := map[string]string{
		filepath.Join(gadgetPath, "idVendor"):                                 "0x" + vid + "\n",
		filepath.Join(gadgetPath, "idProduct"):                                "0x" + pid + "\n",
		filepath.Join(gadgetPath, "strings/0x409/serialnumber"):               serial + "\n",
		filepath.Join(gadgetPath, "strings/0x409/manufacturer"):               strings.Title(brand) + "\n",
		filepath.Join(gadgetPath, "strings/0x409/product"):                    strings.Title(brand) + " USB Flash\n",
		filepath.Join(gadgetPath, "configs/c.1/strings/0x409/configuration"):  "Config 1\n",
		filepath.Join(gadgetPath, "functions/mass_storage.0/lun.0/file"):      imagePath + "\n",
		filepath.Join(gadgetPath, "functions/mass_storage.0/lun.0/removable"): "1\n",
		filepath.Join(gadgetPath, "functions/mass_storage.0/lun.0/ro"):        "0\n",
	}

	for path, content := range strings {
		if err := ctx.Platform.WriteString(path, content); err != nil {
			return fmt.Errorf("file write failed for %s: %w", path, err)
		}
	}

	linkDst := filepath.Join(gadgetPath, "configs/c.1", "mass_storage.0")
	linkSrc := filepath.Join(gadgetPath, "functions/mass_storage.0")

	// Remove existing symlink if it exists
	if err := os.Remove(linkDst); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove existing symlink: %w", err)
	}

	if err := os.Symlink(linkSrc, linkDst); err != nil {
		return fmt.Errorf("function linking failed: %w", err)
	}

	return nil
}

func (m *Manager) saveMetadata(ctx Context, gadget *Gadget) error {
	metadata := fmt.Sprintf(
		"NAME=%q\nBRAND=%q\nVID=%q\nPID=%q\nSERIAL=%q\nIMG=%q\nFS=%q\nSIZE=%q\nAUTOSTART=1\n",
		gadget.Name, gadget.Brand, gadget.VID, gadget.PID, gadget.Serial, gadget.ImagePath, gadget.FileSystem, gadget.Size,
	)

	metadataPath := m.getMetadataPath(ctx, gadget.Name)
	return ctx.Platform.WriteString(metadataPath, metadata)
}

func (m *Manager) parseMetadata(content string) (*Gadget, error) {
	if content == "" {
		return nil, fmt.Errorf("empty metadata content")
	}

	lines := strings.Split(content, "\n")
	metadata := make(map[string]string, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if idx := strings.IndexByte(line, '='); idx > 0 {
			key := strings.TrimSpace(line[:idx])
			value := strings.Trim(strings.Trim(line[idx+1:], "\""), " ")
			if key != "" && value != "" {
				metadata[key] = value
			}
		}
	}

	required := []string{"NAME", "BRAND", "VID", "PID", "SERIAL", "IMG", "FS", "SIZE", "AUTOSTART"}
	for _, field := range required {
		if metadata[field] == "" {
			return nil, fmt.Errorf("missing required field: %s", field)
		}
	}

	// Validate gadget name
	if err := m.validateGadgetName(metadata["NAME"]); err != nil {
		return nil, fmt.Errorf("invalid gadget name in metadata: %w", err)
	}

	// Validate brand
	if !m.isValidBrand(metadata["BRAND"]) {
		return nil, fmt.Errorf("invalid brand in metadata: %s", metadata["BRAND"])
	}

	// Validate filesystem
	if !m.isValidFS(metadata["FS"]) {
		return nil, fmt.Errorf("invalid filesystem in metadata: %s", metadata["FS"])
	}

	// Validate size
	if !m.isValidSize(metadata["SIZE"]) {
		return nil, fmt.Errorf("invalid size in metadata: %s", metadata["SIZE"])
	}

	gadget := &Gadget{
		Name:       metadata["NAME"],
		Brand:      Brand(metadata["BRAND"]),
		VID:        metadata["VID"],
		PID:        metadata["PID"],
		Serial:     metadata["SERIAL"],
		ImagePath:  metadata["IMG"],
		FileSystem: FileSystem(metadata["FS"]),
		Size:       metadata["SIZE"],
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	return gadget, nil
}

func (m *Manager) isValidSize(size string) bool {
	validSizes := []string{"64M", "128M", "256M", "512M", "1G", "2G", "4G", "8G", "16G", "32G", "64G"}
	for _, valid := range validSizes {
		if size == valid {
			return true
		}
	}
	return false
}

func (m *Manager) isValidBrand(brand string) bool {
	validBrands := []string{"sandisk", "kingston", "corsair", "samsung", "generic"}
	for _, valid := range validBrands {
		if strings.ToLower(brand) == valid {
			return true
		}
	}
	return false
}

func (m *Manager) isValidFS(fs string) bool {
	validFS := []string{"fat32", "exfat", "none"}
	for _, valid := range validFS {
		if strings.ToLower(fs) == valid {
			return true
		}
	}
	return false
}
