package main

import (
	"file_sync/internal/app"
	"file_sync/internal/templates"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

const (
	WindowWidth         = 960
	WindowHeight        = 700
	WindowTitle         = "FILE SYNC - 초경량 파일 동기화"
	ConfigPath          = "presets.json"
	SingleInstanceAppID = "com.filesync.desktop.app.lock"
)

func main() {
	application := app.NewApp(ConfigPath)

	err := wails.Run(&options.App{
		Title:  WindowTitle,
		Width:  WindowWidth,
		Height: WindowHeight,
		AssetServer: &assetserver.Options{
			Assets: templates.Assets,
		},
		BackgroundColour: &options.RGBA{R: 9, G: 9, B: 11, A: 255},
		OnStartup:        application.Startup,
		OnDomReady:       application.DomReady,
		OnShutdown:       application.Shutdown,
		HideWindowOnClose: true,
		SingleInstanceLock: &options.SingleInstanceLock{
			UniqueId: SingleInstanceAppID,
			OnSecondInstanceLaunch: func(secondInstanceData options.SecondInstanceData) {
				application.ShowWindow()
			},
		},
		Bind: []interface{}{
			application,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
		},
	})

	if err != nil {
		log.Fatal("Error starting file_sync:", err.Error())
	}
}

