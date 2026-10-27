package shell

import (
	"errors"
	"file_sync/internal/engine/core"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

const (
	DefaultDebounceDurationMs = 300
)

// WatchDaemon monitors source directories and triggers synchronization upon changes.
type WatchDaemon struct {
	watcher    *fsnotify.Watcher
	mu         sync.Mutex
	running    bool
	stopChan   chan struct{}
	debounceMs time.Duration
}

// NewWatchDaemon initializes a new watch daemon.
func NewWatchDaemon() *WatchDaemon {
	return &WatchDaemon{
		debounceMs: DefaultDebounceDurationMs * time.Millisecond,
	}
}

// IsRunning reports whether the daemon is actively monitoring files.
func (d *WatchDaemon) IsRunning() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.running
}

// Start begins monitoring all sources in the preset and triggers the callback after debouncing.
func (d *WatchDaemon) Start(preset Preset, onTrigger func()) error {
	d.mu.Lock()
	if d.running {
		d.mu.Unlock()
		return errors.New("watch daemon is already running")
	}

	w, err := fsnotify.NewWatcher()
	if err != nil {
		d.mu.Unlock()
		return err
	}

	d.watcher = w
	d.running = true
	d.stopChan = make(chan struct{})
	d.mu.Unlock()

	// Register all sources recursively
	for _, src := range preset.Sources {
		normSrc := core.NormalizePath(src)
		info, statErr := os.Stat(normSrc)
		if statErr != nil {
			continue
		}

		if !info.IsDir() {
			_ = w.Add(normSrc)
			continue
		}

		_ = filepath.WalkDir(normSrc, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr == nil && entry.IsDir() {
				_ = w.Add(path)
			}
			return nil
		})
	}

	go d.eventLoop(onTrigger)

	return nil
}

// Stop stops the file watcher and cleans up goroutines.
func (d *WatchDaemon) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()

	if !d.running {
		return
	}

	d.running = false
	close(d.stopChan)
	if d.watcher != nil {
		_ = d.watcher.Close()
		d.watcher = nil
	}
}

func (d *WatchDaemon) eventLoop(onTrigger func()) {
	var timer *time.Timer
	var timerMu sync.Mutex

	for {
		select {
		case <-d.stopChan:
			timerMu.Lock()
			if timer != nil {
				timer.Stop()
			}
			timerMu.Unlock()
			return

		case event, ok := <-d.watcher.Events:
			if !ok {
				return
			}
			// Skip chmod only events
			if event.Has(fsnotify.Chmod) {
				continue
			}

			// Dynamically register newly created directory
			if event.Has(fsnotify.Create) {
				if fi, err := os.Stat(event.Name); err == nil && fi.IsDir() {
					_ = d.watcher.Add(event.Name)
				}
			}

			timerMu.Lock()
			if timer != nil {
				timer.Stop()
			}
			timer = time.AfterFunc(d.debounceMs, func() {
				if onTrigger != nil {
					onTrigger()
				}
			})
			timerMu.Unlock()

		case _, ok := <-d.watcher.Errors:
			if !ok {
				return
			}
		}
	}
}
