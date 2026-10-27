package app

import (
	"errors"
	"file_sync/internal/engine/shell"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)


// Envelope standardizes all IPC communication between Go and JavaScript.
type Envelope struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

func Ok(data any) Envelope {
	return Envelope{Success: true, Data: data}
}

func Fail(err error) Envelope {
	if err == nil {
		return Envelope{Success: false, Error: "unknown error"}
	}
	return Envelope{Success: false, Error: err.Error()}
}

// SelectFiles opens native OS dialog to select multiple files.
func (a *App) SelectFiles() Envelope {
	if a.ctx == nil {
		return Fail(errors.New("runtime context not ready"))
	}

	files, err := wailsRuntime.OpenMultipleFilesDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "동기화할 파일 다중 선택",
	})
	if err != nil {
		return Fail(err)
	}
	return Ok(files)
}

// SelectDirectory opens native OS dialog to select a single directory.
func (a *App) SelectDirectory() Envelope {
	if a.ctx == nil {
		return Fail(errors.New("runtime context not ready"))
	}

	dir, err := wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
		Title: "동기화 대상 폴더 선택",
	})
	if err != nil {
		return Fail(err)
	}
	return Ok(dir)
}

// GetConfig returns the full presets configuration.
func (a *App) GetConfig() Envelope {
	cfg, err := a.storage.Load()
	if err != nil {
		return Fail(err)
	}
	return Ok(cfg)
}

// SavePreset creates or updates a preset in the storage.
func (a *App) SavePreset(preset shell.Preset) Envelope {
	cfg, err := a.storage.Load()
	if err != nil {
		return Fail(err)
	}

	found := false
	for i, p := range cfg.Presets {
		if p.ID == preset.ID {
			cfg.Presets[i] = preset
			found = true
			break
		}
	}
	if !found {
		cfg.Presets = append(cfg.Presets, preset)
	}

	if cfg.ActivePresetID == "" {
		cfg.ActivePresetID = preset.ID
	}

	if err := a.storage.Save(cfg); err != nil {
		return Fail(err)
	}

	return Ok(cfg)
}

// DeletePreset removes a preset by ID.
func (a *App) DeletePreset(presetID string) Envelope {
	cfg, err := a.storage.Load()
	if err != nil {
		return Fail(err)
	}

	newPresets := make([]shell.Preset, 0, len(cfg.Presets))
	for _, p := range cfg.Presets {
		if p.ID != presetID {
			newPresets = append(newPresets, p)
		}
	}
	cfg.Presets = newPresets

	if cfg.ActivePresetID == presetID {
		if len(cfg.Presets) > 0 {
			cfg.ActivePresetID = cfg.Presets[0].ID
		} else {
			cfg.ActivePresetID = ""
		}
	}

	if err := a.storage.Save(cfg); err != nil {
		return Fail(err)
	}

	return Ok(cfg)
}

// SetActivePreset changes the active preset ID.
func (a *App) SetActivePreset(presetID string) Envelope {
	cfg, err := a.storage.Load()
	if err != nil {
		return Fail(err)
	}

	cfg.ActivePresetID = presetID
	if err := a.storage.Save(cfg); err != nil {
		return Fail(err)
	}

	return Ok(cfg)
}

// RunSync triggers full synchronization for the specified preset.
func (a *App) RunSync(presetID string) Envelope {
	cfg, err := a.storage.Load()
	if err != nil {
		return Fail(err)
	}

	var targetPreset *shell.Preset
	for _, p := range cfg.Presets {
		if p.ID == presetID {
			targetPreset = &p
			break
		}
	}

	if targetPreset == nil {
		return Fail(errors.New("preset not found: " + presetID))
	}

	go func() {
		callback := func(current, total int, message string) {
			if a.ctx != nil {
				wailsRuntime.EventsEmit(a.ctx, "sync:progress", map[string]any{
					"current": current,
					"total":   total,
					"message": message,
				})
			}
		}

		result, syncErr := a.syncSvc.ExecutePreset(*targetPreset, callback)
		if a.ctx != nil {
			if syncErr != nil {
				wailsRuntime.EventsEmit(a.ctx, "sync:completed", map[string]any{
					"success": false,
					"error":   syncErr.Error(),
				})
			} else {
				wailsRuntime.EventsEmit(a.ctx, "sync:completed", map[string]any{
					"success": true,
					"result":  result,
				})
			}
		}
	}()

	return Ok("sync started")
}

// ToggleWatch enables or disables the background watch daemon for the active preset.
func (a *App) ToggleWatch(enable bool) Envelope {
	cfg, err := a.storage.Load()
	if err != nil {
		return Fail(err)
	}

	if !enable {
		a.watcher.Stop()
		return Ok(false)
	}

	var activePreset *shell.Preset
	for _, p := range cfg.Presets {
		if p.ID == cfg.ActivePresetID {
			activePreset = &p
			break
		}
	}

	if activePreset == nil {
		return Fail(errors.New("no active preset selected"))
	}

	onTrigger := func() {
		if a.ctx != nil {
			wailsRuntime.EventsEmit(a.ctx, "sync:progress", map[string]any{
				"message": "[WATCH] 변경 감지 - 자동 동기화 시작...",
			})
		}
		a.RunSync(activePreset.ID)
	}

	if err := a.watcher.Start(*activePreset, onTrigger); err != nil {
		return Fail(err)
	}

	return Ok(true)
}
