package main

import (
	"context"
	"embed"
	"log"
	"log/slog"

	"agent"
	utils "utils"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:generate go run gen.go

//go:embed frontend/*
var assetDir embed.FS

func main() {

	// logging
	logger, logFile, err := utils.Logger("desktop.log")
	if err != nil {
		log.Fatalln(err) // exit
	}
	defer logFile.Close()

	slog.SetDefault(logger)

	desktopAgent, err := agent.DesktopAgent(logger)
	if err != nil {
		log.Fatalln("error creating desktopAgent:", err) // exit
	}

	// Create an instance of the app structure
	app := NewApp(desktopAgent)

	// Create application with options
	if err = wails.Run(&options.App{
		Title:  "LiScreMon",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets:  assetDir,
			Handler: desktopAgent.Routes(),
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		// OnShutdown:       app.shutdown,
		Bind: []interface{}{
			app,
		},
	}); err != nil {
		println("Error:", err.Error())
	}

	log.Println("we got a close")
	desktopAgent.CloseDaemonConnection()
}

type AppInterface interface {
	CheckDaemonService() (utils.Message, error)
	CloseDaemonConnection() error
}

// App struct
type App struct {
	ctx          context.Context
	desktopAgent AppInterface
}

// NewApp creates a new App application struct
func NewApp(d AppInterface) *App {
	return &App{desktopAgent: d}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	if _, err := a.desktopAgent.CheckDaemonService(); err != nil {
		log.Fatalln("error connecting to daemon service:", err)
	}
}

func (a *App) shutdown(ctx context.Context) {
	log.Println("we got a close")
	a.desktopAgent.CloseDaemonConnection()
}
