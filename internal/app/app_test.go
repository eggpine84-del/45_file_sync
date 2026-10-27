package app_test

import (
	"bytes"
	"file_sync/internal/app"
	"file_sync/internal/engine/shell"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func TestApp_PresetManagement(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "presets.json")

	application := app.NewApp(configPath)

	// Save Preset
	p1 := shell.Preset{
		ID:      "p1",
		Name:    "Preset 1",
		Sources: []string{"C:/src"},
		Targets: []string{"D:/dst"},
	}

	res := application.SavePreset(p1)
	if !res.Success {
		t.Fatalf("expected SavePreset success, got error: %s", res.Error)
	}

	// Get Config
	res = application.GetConfig()
	if !res.Success {
		t.Fatalf("expected GetConfig success, got error: %s", res.Error)
	}
	cfg, ok := res.Data.(*shell.PresetConfig)
	if !ok || len(cfg.Presets) != 1 {
		t.Fatalf("expected 1 preset in config, got %+v", res.Data)
	}

	// Delete Preset
	res = application.DeletePreset("p1")
	if !res.Success {
		t.Fatalf("expected DeletePreset success, got error: %s", res.Error)
	}

	res = application.GetConfig()
	cfg = res.Data.(*shell.PresetConfig)
	if len(cfg.Presets) != 0 {
		t.Fatalf("expected 0 presets after delete, got %d", len(cfg.Presets))
	}
}

func TestGenerateDefaultIcon(t *testing.T) {
	iconBytes := app.GenerateDefaultIcon()
	if len(iconBytes) < 22 {
		t.Fatalf("expected icon bytes length >= 22, got %d", len(iconBytes))
	}
	// Verify ICO header: 0, 0, 1, 0, 1, 0
	if iconBytes[0] != 0 || iconBytes[1] != 0 || iconBytes[2] != 1 || iconBytes[3] != 0 || iconBytes[4] != 1 || iconBytes[5] != 0 {
		t.Fatalf("invalid ICO header: %v", iconBytes[:6])
	}

	// Verify PNG payload
	pngBytes := iconBytes[22:]
	img, err := png.Decode(bytes.NewReader(pngBytes))
	if err != nil {
		t.Fatalf("failed to decode icon PNG payload: %v", err)
	}

	// Check that emerald arrow pixels exist (R: 16, G: 185, B: 129)
	bounds := img.Bounds()
	emeraldCount := 0
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			// 16-bit color values scaled
			if r>>8 == 16 && g>>8 == 185 && b>>8 == 129 {
				emeraldCount++
			}
		}
	}
	if emeraldCount < 2 {
		t.Fatalf("expected at least 2 emerald arrow pixels, got %d", emeraldCount)
	}
}

func TestExportIcon(t *testing.T) {
	_ = os.MkdirAll(filepath.Join("..", "..", "build"), 0755)
	if err := app.ExportDefaultIcon(filepath.Join("..", "..", "build", "app.ico")); err != nil {
		t.Fatalf("failed to export build/app.ico: %v", err)
	}
	if err := app.ExportDefaultIcon(filepath.Join("..", "..", "cmd", "file_sync", "app.ico")); err != nil {
		t.Fatalf("failed to export cmd/file_sync/app.ico: %v", err)
	}
}

func TestWindowsResource_IconAndSyso(t *testing.T) {
	icoPath := filepath.Join("..", "..", "cmd", "file_sync", "app.ico")
	icoInfo, err := os.Stat(icoPath)
	if err != nil || icoInfo.Size() == 0 {
		t.Fatalf("expected app.ico in cmd/file_sync with size > 0, err: %v", err)
	}

	sysoPath := filepath.Join("..", "..", "cmd", "file_sync", "app.syso")
	sysoInfo, err := os.Stat(sysoPath)
	if err != nil || sysoInfo.Size() == 0 {
		t.Fatalf("expected app.syso in cmd/file_sync with size > 0, err: %v", err)
	}
}
