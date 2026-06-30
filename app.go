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
	"yueduqi-desktop/storage"
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
	ctx   context.Context
	store storage.Store
}

// NewApp creates the application with an optional persistence layer.
// Pass nil to use an in-memory store (state is lost on restart).
func NewApp(store storage.Store) *App {
	if store == nil {
		store, _ = storage.New("")
	}
	return &App{store: store}
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

// ---- Bookshelf methods -----------------------------------------------------

// AddToBookshelf saves a book to the user's bookshelf.
// Returns the shelf row id (re-used when the book is already present).
func (a *App) AddToBookshelf(entry storage.BookshelfEntry) (int64, error) {
	slog.Info("AddToBookshelf called", "bookID", entry.BookID)
	id, err := a.store.AddToBookshelf(entry)
	slog.Info("AddToBookshelf result", "id", id, "err", err)
	return id, err
}

// GetBookshelf returns every book on the shelf with its latest reading progress.
func (a *App) GetBookshelf() ([]storage.BookshelfEntry, error) {
	slog.Info("GetBookshelf called")
	entries, err := a.store.GetBookshelf()
	slog.Info("GetBookshelf result", "count", len(entries), "err", err)
	return entries, err
}

// RemoveFromBookshelf deletes a shelf entry and its reading progress.
func (a *App) RemoveFromBookshelf(id int64) error {
	slog.Info("RemoveFromBookshelf called", "id", id)
	err := a.store.RemoveFromBookshelf(id)
	slog.Info("RemoveFromBookshelf result", "err", err)
	return err
}

// ---- Reading-progress methods ----------------------------------------------

// UpdateProgress records where the user stopped reading for a given book.
func (a *App) UpdateProgress(bookID string, chapterIndex int, itemID string) error {
	slog.Info("UpdateProgress called", "bookID", bookID, "chapterIndex", chapterIndex)
	err := a.store.UpdateProgress(bookID, chapterIndex, itemID)
	slog.Info("UpdateProgress result", "err", err)
	return err
}

// GetProgress returns the last reading position for a book.
func (a *App) GetProgress(bookID string) (storage.ReadingProgress, error) {
	slog.Info("GetProgress called", "bookID", bookID)
	prog, err := a.store.GetProgress(bookID)
	slog.Info("GetProgress result", "chapterIndex", prog.ChapterIndex, "err", err)
	return prog, err
}

// ---- Settings methods ------------------------------------------------------

// GetSetting reads a user preference by key.
func (a *App) GetSetting(key string) (string, error) {
	return a.store.GetSetting(key)
}

// SetSetting writes a user preference.
func (a *App) SetSetting(key, value string) error {
	return a.store.SetSetting(key, value)
}
