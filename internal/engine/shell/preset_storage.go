package shell

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
)

const (
	DefaultConfigFileName = "presets.json"
	FilePermUserOnly      = 0644
	DirPermUserOnly       = 0755
)

// Preset represents a named sync configuration.
type Preset struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Sources   []string `json:"sources"`
	Targets   []string `json:"targets"`
	AutoWatch bool     `json:"auto_watch"`
}

// PresetConfig holds the active selection and the list of presets.
type PresetConfig struct {
	ActivePresetID string   `json:"active_preset_id"`
	Presets        []Preset `json:"presets"`
}

// PresetStorage handles thread-safe read/write of presets JSON.
type PresetStorage struct {
	filePath string
	mu       sync.RWMutex
}

// NewPresetStorage creates a storage instance targeting the given file path.
func NewPresetStorage(filePath string) *PresetStorage {
	if filePath == "" {
		filePath = DefaultConfigFileName
	}
	return &PresetStorage{
		filePath: filePath,
	}
}

// Load reads and parses presets from disk. Returns default empty config if file does not exist.
func (s *PresetStorage) Load() (*PresetConfig, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, err := os.ReadFile(s.filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &PresetConfig{
				ActivePresetID: "",
				Presets:        []Preset{},
			}, nil
		}
		return nil, err
	}

	var cfg PresetConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	if cfg.Presets == nil {
		cfg.Presets = []Preset{}
	}

	return &cfg, nil
}

// Save atomically writes presets to disk.
func (s *PresetStorage) Save(cfg *PresetConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if cfg == nil {
		return errors.New("cannot save nil config")
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	dir := filepath.Dir(s.filePath)
	if dir != "" && dir != "." {
		if err := os.MkdirAll(dir, DirPermUserOnly); err != nil {
			return err
		}
	}

	tmpPath := s.filePath + ".tmp"
	if err := os.WriteFile(tmpPath, data, FilePermUserOnly); err != nil {
		return err
	}

	return os.Rename(tmpPath, s.filePath)
}
