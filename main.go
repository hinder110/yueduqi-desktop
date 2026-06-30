package main

import (
	"embed"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	"yueduqi-desktop/storage"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	// Resolve the database path under the user's config directory.
	// Falls back to in-memory when the data directory cannot be created
	// (e.g. in tightly-sandboxed environments).
	dataDir, err := ensureDataDir()
	if err != nil {
		slog.Warn("cannot create data directory, using in-memory storage", "err", err)
		dataDir = ""
	}
	dbPath := ""
	if dataDir != "" {
		dbPath = filepath.Join(dataDir, "yueduqi.db")
	}
	store, err := storage.New(dbPath)
	if err != nil {
		slog.Warn("storage init failed, using in-memory", "err", err)
		store, _ = storage.New("")
	}

	// Create an instance of the app structure
	app := NewApp(store)

	// Create application with options
	err = wails.Run(&options.App{
		Title:  "yueduqi-desktop",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}

// ensureDataDir returns the path to the app's writable data directory,
// creating it if necessary. It follows the XDG Base Directory convention.
func ensureDataDir() (string, error) {
	base := os.Getenv("XDG_DATA_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".local", "share")
	}
	dir := filepath.Join(base, "yueduqi-desktop")
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return dir, nil
}
