package main

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"

	"yueduqi-desktop/cache"
	"yueduqi-desktop/config"
	"yueduqi-desktop/model"
	"yueduqi-desktop/parser"
)

func init() {
	// Load user config first so we can route logs to the right place.
	// Default slog (stderr) handles any warnings emitted before reconfig.
	cfg, err := config.Load()
	if err != nil {
		slog.Warn("config load failed, using defaults", "err", err)
		cfg = config.Default()
	}

	setupSlog(cfg)
	slog.Info("app started")
}

// setupSlog creates a file-backed TextHandler so logs go to the path in
// config. On failure it falls back to stderr so the app is never silent.
func setupSlog(cfg *config.Config) {
	w := os.Stderr
	if cfg.LogPath != "" {
		// Ensure the parent directory exists before opening the log file.
		if err := os.MkdirAll(filepath.Dir(cfg.LogPath), 0755); err != nil {
			slog.Warn("cannot create log directory, falling back to stderr", "path", filepath.Dir(cfg.LogPath), "err", err)
			cfg.LogPath = "" // force stderr
		}
	}
	if cfg.LogPath != "" {
		f, err := os.OpenFile(cfg.LogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			slog.Warn("cannot open log file, falling back to stderr", "path", cfg.LogPath, "err", err)
		} else {
			w = f
		}
	}

	level := parseLogLevel(cfg.LogLevel)
	handler := slog.NewTextHandler(w, &slog.HandlerOptions{Level: level})
	slog.SetDefault(slog.New(handler))
}

func parseLogLevel(s string) slog.Level {
	switch s {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	slog.Info("startup called", "ctx_set", true)
}

func (a *App) SearchBooks(keyword string, source string) ([]model.Book, error) {
	slog.Info("SearchBooks called", "keyword", keyword, "source", source)
	books, err := parser.ForSource(source).SearchBooks(a.ctx, keyword)
	slog.Info("SearchBooks result", "count", len(books), "err", err)
	return books, err
}

func (a *App) GetHotBooks() ([]model.Book, error) {
	slog.Info("GetHotBooks called")
	books, err := parser.GetHotBooks(a.ctx)
	slog.Info("GetHotBooks result", "count", len(books), "err", err)
	return books, err
}

func (a *App) GetChapters(bookID string, source string, innerSource string, innerTab string) ([]model.Chapter, error) {
	slog.Info("GetChapters called", "bookID", bookID, "source", source, "innerSource", innerSource, "innerTab", innerTab)
	if innerSource == "" { innerSource = "番茄" }
	if innerTab == "" { innerTab = "小说" }
	chapters, err := parser.ForSource(source).GetChapters(a.ctx, bookID, innerSource, innerTab)
	slog.Info("GetChapters result", "count", len(chapters), "err", err)
	return chapters, err
}

func (a *App) GetChapterContent(bookID, itemID string, source string, innerSource string, innerTab string) (model.ChapterContent, error) {
	slog.Info("GetChapterContent called", "bookID", bookID, "itemID", itemID, "source", source, "innerSource", innerSource, "innerTab", innerTab)
	if innerSource == "" { innerSource = "番茄" }
	if innerTab == "" { innerTab = "小说" }
	content, err := parser.ForSource(source).GetChapterContent(a.ctx, bookID, itemID, innerSource, innerTab)
	slog.Info("GetChapterContent result", "title", content.Title, "len", len(content.Content), "err", err)
	return content, err
}

// CacheStats holds hit/miss counters for a single cache.
type CacheStats struct {
	Hits   int64 `json:"hits"`
	Misses int64 `json:"misses"`
}

// GetCacheStats returns per-cache hit/miss counters for frontend debug panels.
func (a *App) GetCacheStats() map[string]CacheStats {
	h, m := cache.HotBooks.Stats()
	s, sm := cache.Search.Stats()
	ch, cm := cache.Chapters.Stats()
	return map[string]CacheStats{
		"hotBooks": {Hits: h, Misses: m},
		"search":   {Hits: s, Misses: sm},
		"chapters": {Hits: ch, Misses: cm},
	}
}
