// Package storage provides persistence for the desktop reader app.
// It uses modernc.org/sqlite (pure Go, no CGO) as the primary backend,
// falling back to an in-memory store when SQLite is unavailable.
package storage

import (
	"database/sql"
	"fmt"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// BookshelfEntry represents a book added to the user's bookshelf,
// including the latest reading progress via a join.
type BookshelfEntry struct {
	ID            int64  `json:"id"`
	BookID        string `json:"bookId"`
	Title         string `json:"title"`
	Author        string `json:"author"`
	Cover         string `json:"cover"`
	Intro         string `json:"intro"`
	SourceKey     string `json:"sourceKey"`
	AddedAt       string `json:"addedAt"`
	ChapterIndex  int    `json:"chapterIndex"`
	ChapterItemID string `json:"chapterItemId"`
}

// ReadingProgress tracks where the user left off in a book.
type ReadingProgress struct {
	BookID       string `json:"bookId"`
	ChapterIndex int    `json:"chapterIndex"`
	ItemID       string `json:"itemId"`
	UpdatedAt    string `json:"updatedAt"`
}

// Store is the persistence contract consumed by the App.
type Store interface {
	AddToBookshelf(entry BookshelfEntry) (int64, error)
	GetBookshelf() ([]BookshelfEntry, error)
	RemoveFromBookshelf(id int64) error
	UpdateProgress(bookID string, chapterIndex int, itemID string) error
	GetProgress(bookID string) (ReadingProgress, error)
	GetSetting(key string) (string, error)
	SetSetting(key, value string) error
	Close() error
}

// ---- sqliteStore -----------------------------------------------------------

type sqliteStore struct {
	db *sql.DB
}

// New opens (or creates) a SQLite database at path.
// If the driver cannot be loaded, it silently returns an in-memory store.
func New(path string) (Store, error) {
	// An empty path forces the in-memory fallback.
	if path == "" {
		return newMemStore(), nil
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return newMemStore(), nil
	}

	// WAL mode improves concurrent read performance.
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return newMemStore(), nil
	}

	if err := migrate(db); err != nil {
		db.Close()
		return newMemStore(), nil
	}

	return &sqliteStore{db: db}, nil
}

// migrate creates tables if they do not exist (idempotent).
func migrate(db *sql.DB) error {
	schema := `
		CREATE TABLE IF NOT EXISTS bookshelf (
			id          INTEGER PRIMARY KEY AUTOINCREMENT,
			book_id     TEXT    NOT NULL UNIQUE,
			title       TEXT    NOT NULL,
			author      TEXT    NOT NULL DEFAULT '',
			cover       TEXT    NOT NULL DEFAULT '',
			intro       TEXT    NOT NULL DEFAULT '',
			source_key  TEXT    NOT NULL DEFAULT '',
			added_at    TEXT    NOT NULL DEFAULT (datetime('now'))
		);
		CREATE TABLE IF NOT EXISTS reading_progress (
			book_id       TEXT PRIMARY KEY,
			chapter_index INTEGER NOT NULL DEFAULT 0,
			item_id       TEXT    NOT NULL DEFAULT '',
			updated_at    TEXT    NOT NULL DEFAULT (datetime('now'))
		);
		CREATE TABLE IF NOT EXISTS settings (
			key   TEXT PRIMARY KEY,
			value TEXT NOT NULL DEFAULT ''
		);
		`
	_, err := db.Exec(schema)
	return err
}

func (s *sqliteStore) AddToBookshelf(entry BookshelfEntry) (int64, error) {
	now := time.Now().UTC().Format(time.RFC3339)
	res, err := s.db.Exec(
		`INSERT OR IGNORE INTO bookshelf (book_id, title, author, cover, intro, source_key, added_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		entry.BookID, entry.Title, entry.Author, entry.Cover, entry.Intro, entry.SourceKey, now,
	)
	if err != nil {
		return 0, fmt.Errorf("add bookshelf: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("last insert id: %w", err)
	}
	// INSERT OR IGNORE returns 0 for an existing row; fetch the real id.
	if id == 0 {
		err = s.db.QueryRow("SELECT id FROM bookshelf WHERE book_id = ?", entry.BookID).Scan(&id)
		if err != nil {
			return 0, fmt.Errorf("fetch existing bookshelf id: %w", err)
		}
	}
	return id, nil
}

func (s *sqliteStore) GetBookshelf() ([]BookshelfEntry, error) {
	rows, err := s.db.Query(`
			SELECT b.id, b.book_id, b.title, b.author, b.cover, b.intro,
			       b.source_key, b.added_at,
			       COALESCE(p.chapter_index, 0), COALESCE(p.item_id, '')
			FROM bookshelf b
			LEFT JOIN reading_progress p ON b.book_id = p.book_id
			ORDER BY b.added_at DESC
		`)
	if err != nil {
		return nil, fmt.Errorf("query bookshelf: %w", err)
	}
	defer rows.Close()

	var entries []BookshelfEntry
	for rows.Next() {
		var e BookshelfEntry
		if err := rows.Scan(&e.ID, &e.BookID, &e.Title, &e.Author,
			&e.Cover, &e.Intro, &e.SourceKey, &e.AddedAt,
			&e.ChapterIndex, &e.ChapterItemID); err != nil {
			return nil, fmt.Errorf("scan bookshelf row: %w", err)
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func (s *sqliteStore) RemoveFromBookshelf(id int64) error {
	// Cascade-delete reading progress via the known book_id.
	var bookID string
	if err := s.db.QueryRow("SELECT book_id FROM bookshelf WHERE id = ?", id).Scan(&bookID); err != nil {
		return fmt.Errorf("bookshelf entry %d not found: %w", id, err)
	}
	if _, err := s.db.Exec("DELETE FROM bookshelf WHERE id = ?", id); err != nil {
		return fmt.Errorf("delete from bookshelf: %w", err)
	}
	// Clean up orphaned progress row (best-effort).
	_, _ = s.db.Exec("DELETE FROM reading_progress WHERE book_id = ?", bookID)
	return nil
}

func (s *sqliteStore) UpdateProgress(bookID string, chapterIndex int, itemID string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.Exec(
		`INSERT INTO reading_progress (book_id, chapter_index, item_id, updated_at)
		 VALUES (?, ?, ?, ?)
		 ON CONFLICT(book_id) DO UPDATE SET chapter_index=excluded.chapter_index,
		                                     item_id=excluded.item_id,
		                                     updated_at=excluded.updated_at`,
		bookID, chapterIndex, itemID, now,
	)
	if err != nil {
		return fmt.Errorf("upsert progress: %w", err)
	}
	return nil
}

func (s *sqliteStore) GetProgress(bookID string) (ReadingProgress, error) {
	var p ReadingProgress
	err := s.db.QueryRow(
		"SELECT book_id, chapter_index, item_id, updated_at FROM reading_progress WHERE book_id = ?",
		bookID,
	).Scan(&p.BookID, &p.ChapterIndex, &p.ItemID, &p.UpdatedAt)
	if err != nil {
		return ReadingProgress{}, fmt.Errorf("get progress %s: %w", bookID, err)
	}
	return p, nil
}

func (s *sqliteStore) GetSetting(key string) (string, error) {
	var val string
	err := s.db.QueryRow("SELECT value FROM settings WHERE key = ?", key).Scan(&val)
	if err != nil {
		return "", fmt.Errorf("get setting %s: %w", key, err)
	}
	return val, nil
}

func (s *sqliteStore) SetSetting(key, value string) error {
	_, err := s.db.Exec(
		"INSERT INTO settings (key, value) VALUES (?, ?) ON CONFLICT(key) DO UPDATE SET value=excluded.value",
		key, value,
	)
	if err != nil {
		return fmt.Errorf("set setting %s: %w", key, err)
	}
	return nil
}

func (s *sqliteStore) Close() error {
	return s.db.Close()
}

// ---- memStore (in-memory fallback) -----------------------------------------

type memStore struct {
	mu       sync.RWMutex
	shelf    []BookshelfEntry
	nextID   int64
	progress map[string]ReadingProgress
	settings map[string]string
}

// newMemStore returns a Store that keeps all data in process memory.
// Used when SQLite is unavailable.
func newMemStore() Store {
	return &memStore{
		shelf:    make([]BookshelfEntry, 0),
		nextID:   1,
		progress: make(map[string]ReadingProgress),
		settings: make(map[string]string),
	}
}

func (m *memStore) AddToBookshelf(entry BookshelfEntry) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Deduplicate by book_id so the same book is not added twice.
	for _, e := range m.shelf {
		if e.BookID == entry.BookID {
			return e.ID, nil
		}
	}

	entry.ID = m.nextID
	m.nextID++
	entry.AddedAt = time.Now().UTC().Format(time.RFC3339)
	m.shelf = append(m.shelf, entry)
	return entry.ID, nil
}

func (m *memStore) GetBookshelf() ([]BookshelfEntry, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]BookshelfEntry, len(m.shelf))
	for i, e := range m.shelf {
		result[i] = e
		if p, ok := m.progress[e.BookID]; ok {
			result[i].ChapterIndex = p.ChapterIndex
			result[i].ChapterItemID = p.ItemID
		}
	}
	return result, nil
}

func (m *memStore) RemoveFromBookshelf(id int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, e := range m.shelf {
		if e.ID == id {
			delete(m.progress, e.BookID)
			m.shelf = append(m.shelf[:i], m.shelf[i+1:]...)
			return nil
		}
	}
	return fmt.Errorf("bookshelf entry %d not found", id)
}

func (m *memStore) UpdateProgress(bookID string, chapterIndex int, itemID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.progress[bookID] = ReadingProgress{
		BookID:       bookID,
		ChapterIndex: chapterIndex,
		ItemID:       itemID,
		UpdatedAt:    time.Now().UTC().Format(time.RFC3339),
	}
	return nil
}

func (m *memStore) GetProgress(bookID string) (ReadingProgress, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	p, ok := m.progress[bookID]
	if !ok {
		return ReadingProgress{}, fmt.Errorf("no progress for book %s", bookID)
	}
	return p, nil
}

func (m *memStore) GetSetting(key string) (string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	v, ok := m.settings[key]
	if !ok {
		return "", fmt.Errorf("setting %s not found", key)
	}
	return v, nil
}

func (m *memStore) SetSetting(key, value string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.settings[key] = value
	return nil
}

func (m *memStore) Close() error {
	return nil
}
