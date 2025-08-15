package gadget

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"virtusb/internal/core/storage"
	"virtusb/internal/utils"
)

// Manager implémente GadgetManager
type Manager struct {
	storageManager storage.StorageManager
}

// NewManager crée un nouveau gestionnaire de gadgets
func NewManager(storageManager storage.StorageManager) *Manager {
	return &Manager{
		storageManager: storageManager,
	}
}

// Create crée un nouveau gadget
func (m *Manager) Create(ctx Context, opts CreateOptions) (*Gadget, error) {
	// Vérifier les privilèges
	if err := ctx.Platform.RequireRoot(); err != nil {
		return nil, err
	}

	// Assurer l'environnement
	if err := ctx.Platform.EnsureEnvironment(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ensure environment: %w", err)
	}

	// Vérifier si le gadget existe déjà
	gadgetPath := m.getGadgetPath(ctx, opts.Name)
	if ctx.Platform.FileExists(gadgetPath) {
		return nil, fmt.Errorf("gadget '%s' already exists", opts.Name)
	}

	// Générer les identifiants
	vid, pid := utils.VidPid(string(opts.Brand))
	serial := opts.Serial
	if serial == "" {
		serial = utils.SerialFor(string(opts.Brand))
	}

	// Créer l'image si elle n'existe pas
	imagePath := m.getImagePath(ctx, opts.Name)
	if !ctx.Platform.FileExists(imagePath) {
		storageCtx := storage.Context{
			Platform: ctx.Platform,
		}
		if err := m.storageManager.CreateImage(storageCtx, imagePath, opts.Size, string(opts.FileSystem)); err != nil {
			return nil, fmt.Errorf("failed to create image: %w", err)
		}
	}

	// Créer le gadget
	if err := m.createGadgetStructure(ctx, opts.Name, vid, pid, serial, string(opts.Brand), imagePath); err != nil {
		return nil, fmt.Errorf("failed to create gadget structure: %w", err)
	}

	// Sauvegarder les métadonnées
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
		return nil, fmt.Errorf("failed to save metadata: %w", err)
	}

	return gadget, nil
}

// Get récupère un gadget par son nom
func (m *Manager) Get(ctx Context, name string) (*Gadget, error) {
	metadataPath := m.getMetadataPath(ctx, name)
	if !ctx.Platform.FileExists(metadataPath) {
		return nil, fmt.Errorf("gadget '%s' not found", name)
	}

	content, err := ctx.Platform.ReadString(metadataPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read metadata: %w", err)
	}

	gadget, err := m.parseMetadata(content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse metadata: %w", err)
	}

	// Vérifier si le gadget est activé
	gadget.Enabled = m.isGadgetEnabled(ctx, name)

	return gadget, nil
}

// List liste les gadgets
func (m *Manager) List(ctx Context, opts ListOptions) ([]*Gadget, error) {
	stateDir := ctx.Platform.GetStateDir()
	entries, err := os.ReadDir(stateDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read state directory: %w", err)
	}

	var gadgets []*Gadget
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".env") {
			continue
		}

		name := strings.TrimSuffix(entry.Name(), ".env")
		gadget, err := m.Get(ctx, name)
		if err != nil {
			// Ignorer les gadgets corrompus
			continue
		}

		// Appliquer les filtres
		if opts.Enabled != nil && gadget.Enabled != *opts.Enabled {
			continue
		}
		if opts.Brand != "" && gadget.Brand != opts.Brand {
			continue
		}

		gadgets = append(gadgets, gadget)
	}

	return gadgets, nil
}

// Enable active un gadget
func (m *Manager) Enable(ctx Context, name string) error {
	// Vérifier les privilèges
	if err := ctx.Platform.RequireRoot(); err != nil {
		return err
	}

	// Vérifier que le gadget existe
	gadgetPath := m.getGadgetPath(ctx, name)
	if !ctx.Platform.FileExists(gadgetPath) {
		return fmt.Errorf("gadget '%s' not found", name)
	}

	// Vérifier qu'au moins une fonction est liée
	if err := m.ensureFunctionsLinked(ctx, name); err != nil {
		return err
	}

	// Obtenir un UDC
	udc, err := ctx.Platform.GetFirstUDC()
	if err != nil {
		return fmt.Errorf("no UDC available: %w", err)
	}

	// Activer le gadget
	udcFile := filepath.Join(gadgetPath, "UDC")
	if err := ctx.Platform.WriteString(udcFile, udc+"\n"); err != nil {
		return fmt.Errorf("failed to write UDC: %w", err)
	}

	// Vérifier l'activation
	if !m.isGadgetEnabled(ctx, name) {
		return fmt.Errorf("failed to bind gadget '%s' to UDC %s", name, udc)
	}

	return nil
}

// Disable désactive un gadget
func (m *Manager) Disable(ctx Context, name string) error {
	// Vérifier les privilèges
	if err := ctx.Platform.RequireRoot(); err != nil {
		return err
	}

	// Vérifier que le gadget existe
	gadgetPath := m.getGadgetPath(ctx, name)
	if !ctx.Platform.FileExists(gadgetPath) {
		return fmt.Errorf("gadget '%s' not found", name)
	}

	// Désactiver le gadget
	udcFile := filepath.Join(gadgetPath, "UDC")
	if err := ctx.Platform.WriteString(udcFile, "\n"); err != nil {
		return fmt.Errorf("failed to clear UDC: %w", err)
	}

	return nil
}

// Delete supprime un gadget
func (m *Manager) Delete(ctx Context, name string) error {
	// Désactiver d'abord
	if err := m.Disable(ctx, name); err != nil {
		return err
	}

	// Supprimer la structure du gadget
	gadgetPath := m.getGadgetPath(ctx, name)
	if err := os.RemoveAll(gadgetPath); err != nil {
		return fmt.Errorf("failed to remove gadget directory: %w", err)
	}

	// Supprimer les métadonnées
	metadataPath := m.getMetadataPath(ctx, name)
	if err := os.Remove(metadataPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove metadata: %w", err)
	}

	return nil
}

// Restore restaure tous les gadgets
func (m *Manager) Restore(ctx Context) error {
	// Vérifier les privilèges
	if err := ctx.Platform.RequireRoot(); err != nil {
		return err
	}

	// Assurer l'environnement
	if err := ctx.Platform.EnsureEnvironment(context.Background()); err != nil {
		return fmt.Errorf("failed to ensure environment: %w", err)
	}

	// Lister tous les gadgets
	gadgets, err := m.List(ctx, ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list gadgets: %w", err)
	}

	// Restaurer chaque gadget
	for _, gadget := range gadgets {
		gadgetPath := m.getGadgetPath(ctx, gadget.Name)
		if !ctx.Platform.FileExists(gadgetPath) {
			// Recréer la structure si elle n'existe pas
			if err := m.createGadgetStructure(ctx, gadget.Name, gadget.VID, gadget.PID, gadget.Serial, string(gadget.Brand), gadget.ImagePath); err != nil {
				return fmt.Errorf("failed to recreate gadget '%s': %w", gadget.Name, err)
			}
		}

		// Activer le gadget
		if err := m.Enable(ctx, gadget.Name); err != nil {
			return fmt.Errorf("failed to enable gadget '%s': %w", gadget.Name, err)
		}
	}

	return nil
}

// Méthodes utilitaires privées

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
		return fmt.Errorf("failed to read config directory: %w", err)
	}

	for _, entry := range entries {
		if entry.Type()&os.ModeSymlink != 0 {
			return nil // Au moins une fonction est liée
		}
	}

	return fmt.Errorf("gadget '%s' has no functions linked to configs/c.1", name)
}

func (m *Manager) createGadgetStructure(ctx Context, name, vid, pid, serial, brand, imagePath string) error {
	gadgetPath := m.getGadgetPath(ctx, name)

	// Créer les dossiers de base
	dirs := []string{
		filepath.Join(gadgetPath, "strings/0x409"),
		filepath.Join(gadgetPath, "configs/c.1/strings/0x409"),
		filepath.Join(gadgetPath, "functions/mass_storage.0/lun.0"),
	}

	for _, dir := range dirs {
		if err := ctx.Platform.CreateDirectory(dir); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Écrire les identifiants et chaînes
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
			return fmt.Errorf("failed to write %s: %w", path, err)
		}
	}

	// Lier la fonction à la configuration
	linkDst := filepath.Join(gadgetPath, "configs/c.1", "mass_storage.0")
	linkSrc := filepath.Join(gadgetPath, "functions/mass_storage.0")

	// Supprimer le lien existant s'il existe
	os.Remove(linkDst)

	if err := os.Symlink(linkSrc, linkDst); err != nil {
		return fmt.Errorf("failed to link mass_storage function: %w", err)
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
	// Implémentation simple du parsing des métadonnées
	// Pour une implémentation complète, utiliser un parser plus robuste
	lines := strings.Split(content, "\n")
	metadata := make(map[string]string)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := parts[0]
			value := strings.Trim(parts[1], "\"")
			metadata[key] = value
		}
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
		CreatedAt:  time.Now(), // À améliorer avec un vrai timestamp
		UpdatedAt:  time.Now(),
	}

	return gadget, nil
}
