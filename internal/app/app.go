package app

import (
	"context"
	"file_sync/internal/engine/shell"
	"sync"

	"github.com/getlantern/systray"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App manages Wails desktop lifecycle and wires services.
type App struct {
	ctx     context.Context
	storage *shell.PresetStorage
	syncSvc *shell.SyncService
	watcher *shell.WatchDaemon
	mu      sync.Mutex
}

// NewApp creates a new App application struct.
func NewApp(storagePath string) *App {
	storage := shell.NewPresetStorage(storagePath)
	syncSvc := shell.NewSyncService(nil)
	watcher := shell.NewWatchDaemon()

	return &App{
		storage: storage,
		syncSvc: syncSvc,
		watcher: watcher,
	}
}

// Startup is called when the app starts. The context is saved so we can call the runtime methods.
func (a *App) Startup(ctx context.Context) {
	a.ctx = ctx
	SetupTray(a)
}

// DomReady is called after front-end resources are loaded.
func (a *App) DomReady(ctx context.Context) {
}

// Shutdown is called at application termination.
func (a *App) Shutdown(ctx context.Context) {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.watcher != nil {
		a.watcher.Stop()
	}
	systray.Quit()
}

// ToggleWatchFromTray toggles the watch daemon from tray menu.
func (a *App) ToggleWatchFromTray() {
	if a.watcher == nil {
		return
	}
	running := a.watcher.IsRunning()
	a.ToggleWatch(!running)
}

// ShowWindow brings the window to the foreground.
func (a *App) ShowWindow() {
	if a.ctx != nil {
		wailsRuntime.WindowShow(a.ctx)
	}
}

// HideWindow hides the window into background/tray.
func (a *App) HideWindow() {
	if a.ctx != nil {
		wailsRuntime.WindowHide(a.ctx)
	}
}

// Quit terminates the desktop application.
func (a *App) Quit() {
	if a.ctx != nil {
		wailsRuntime.Quit(a.ctx)
	}
}
