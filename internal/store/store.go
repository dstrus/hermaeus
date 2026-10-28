// Package store is the only code that opens SkyrimNet's databases. It knows
// which SkyrimNet folder Hermaeus is working on, and makes every write safe:
// refused while Skyrim is running, in one transaction, and after a backup of
// the database made the first time it changes in a session.
package store

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"hermaeus/internal/apperr"
	"hermaeus/internal/install"
)

type Store struct {
	mu          sync.Mutex
	writeMu     sync.Mutex
	dir         string            // SkyrimNet folder; "" until one is found or chosen
	backupRoot  string            // "" means <dir>/hermaeus-backups
	backups     map[string]string // database path -> its backup this session
	players     map[string]playerName
	gameRunning func() bool
}

// New makes a store with no SkyrimNet folder yet. backupRoot is where
// backups go, or "" for hermaeus-backups in the SkyrimNet folder.
func New(backupRoot string, gameRunning func() bool) *Store {
	return &Store{backupRoot: backupRoot, backups: map[string]string{}, players: map[string]playerName{}, gameRunning: gameRunning}
}

func (s *Store) Dir() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.dir
}

// SetDir points Hermaeus at a SkyrimNet folder and returns its absolute
// path. See install.Check for what it accepts.
func (s *Store) SetDir(dir string) (string, error) {
	abs, err := install.Check(dir)
	if err != nil {
		return "", err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.dir = abs
	return abs, nil
}

func (s *Store) BackupDir() string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.backupDir()
}

func (s *Store) backupDir() string {
	if s.backupRoot != "" {
		return s.backupRoot
	}
	if s.dir == "" {
		return ""
	}
	return filepath.Join(s.dir, "hermaeus-backups")
}

func (s *Store) GameRunning() bool { return s.gameRunning() }

// SaveDir is a playthrough's folder under saves, where SkyrimNet keeps its
// save-specific files such as bios. It needn't exist yet.
func (s *Store) SaveDir(playthrough string) (string, error) {
	if _, err := s.path(playthrough); err != nil {
		return "", err
	}
	return filepath.Join(s.Dir(), "saves", playthrough), nil
}

// ChangeFiles runs fn, which changes files SkyrimNet reads, under the same
// rules as a database write: refused while Skyrim is running, and one change
// at a time. what describes the change for the log.
func (s *Store) ChangeFiles(what string, fn func() error) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	if s.gameRunning() {
		return apperr.New(apperr.Conflict, "Skyrim is running. Close the game before saving; SkyrimNet changes its files while you play.")
	}
	if err := fn(); err != nil {
		return err
	}
	log.Print(what)
	return nil
}

// Read runs fn on a playthrough's database, opened read-only.
func Read[T any](s *Store, playthrough string, fn func(DB) (T, error)) (T, error) {
	var zero T
	path, err := s.path(playthrough)
	if err != nil {
		return zero, err
	}
	db, err := open(path, true)
	if err != nil {
		return zero, err
	}
	defer db.Close()
	return fn(db)
}

// Write runs fn in one transaction on a playthrough's database, after
// checking the game is closed and that the database has been backed up this
// session. what describes the change for the log.
func Write[T any](s *Store, playthrough, what string, fn func(DB) (T, error)) (T, error) {
	var zero T
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	path, err := s.path(playthrough)
	if err != nil {
		return zero, err
	}
	if s.gameRunning() {
		return zero, apperr.New(apperr.Conflict, "Skyrim is running. Close the game before saving; SkyrimNet keeps this database open.")
	}
	if err := s.backup(path); err != nil {
		return zero, fmt.Errorf("could not back up the database, so nothing was changed: %w", err)
	}
	db, err := open(path, false)
	if err != nil {
		return zero, err
	}
	defer db.Close()
	tx, err := db.Begin()
	if err != nil {
		return zero, err
	}
	result, err := fn(tx)
	if err != nil {
		tx.Rollback()
		return zero, err
	}
	if err := tx.Commit(); err != nil {
		return zero, err
	}
	log.Print(what)
	return result, nil
}

func (s *Store) backup(path string) error {
	s.mu.Lock()
	_, done := s.backups[path]
	dir := s.backupDir()
	s.mu.Unlock()
	if done {
		return nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	dest := filepath.Join(dir, fmt.Sprintf("%s.%s.bak", filepath.Base(path), time.Now().Format("20060102-150405")))
	if err := copyFile(path, dest); err != nil {
		return err
	}
	s.mu.Lock()
	s.backups[path] = dest
	s.mu.Unlock()
	log.Printf("Backup saved: %s", dest)
	return nil
}

// copyFile copies src to dest, refusing to overwrite an existing file.
func copyFile(src, dest string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dest, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		os.Remove(dest)
		return err
	}
	return out.Close()
}
