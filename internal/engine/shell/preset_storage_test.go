package shell_test

import (
	"file_sync/internal/engine/shell"
	"os"
	"path/filepath"
	"testing"
)

func TestPresetStorage_EmptyFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "non_existent.json")

	storage := shell.NewPresetStorage(configPath)
	cfg, err := storage.Load()
	if err != nil {
		t.Fatalf("expected no error for non-existent file, got: %v", err)
	}

	if cfg == nil || len(cfg.Presets) != 0 {
		t.Fatal("expected empty preset list")
	}
}

func TestPresetStorage_SaveAndLoad(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "presets.json")

	storage := shell.NewPresetStorage(configPath)

	orig := &shell.PresetConfig{
		ActivePresetID: "preset-1",
		Presets: []shell.Preset{
			{
				ID:        "preset-1",
				Name:      "Workspace Sync",
				Sources:   []string{"C:/Src1", "C:/Src2"},
				Targets:   []string{"D:/Dst1"},
				AutoWatch: true,
			},
		},
	}

	if err := storage.Save(orig); err != nil {
		t.Fatalf("failed to save preset: %v", err)
	}

	loaded, err := storage.Load()
	if err != nil {
		t.Fatalf("failed to load preset: %v", err)
	}

	if loaded.ActivePresetID != orig.ActivePresetID {
		t.Fatalf("expected active ID %s, got %s", orig.ActivePresetID, loaded.ActivePresetID)
	}

	if len(loaded.Presets) != 1 {
		t.Fatalf("expected 1 preset, got %d", len(loaded.Presets))
	}

	if loaded.Presets[0].Name != "Workspace Sync" {
		t.Fatalf("expected preset name 'Workspace Sync', got %s", loaded.Presets[0].Name)
	}

	// Verify temp file cleaned up
	if _, err := os.Stat(configPath + ".tmp"); !os.IsNotExist(err) {
		t.Fatal("temporary file was not cleaned up after atomic save")
	}
}
