package store

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"hermaeus/internal/apperr"
)

// Playthrough is one of SkyrimNet's databases, data/SkyrimNet-<save ID>.db.
// Its ID is the save ID, which addresses it in URLs.
type Playthrough struct {
	ID       string `json:"id"`
	File     string `json:"file"`
	Player   string `json:"player"` // "" until the save has a player name
	Size     int64  `json:"size"`
	Modified int64  `json:"modified"` // Unix ms
	path     string
	modTime  time.Time
}

// playerName is cached per database file until the file changes.
type playerName struct {
	modTime time.Time
	size    int64
	name    string
}

// playthroughID is the save ID in a database's file name, or the whole name
// without .db for a file that doesn't follow SkyrimNet's pattern.
func playthroughID(file string) string {
	id := strings.TrimSuffix(file, filepath.Ext(file))
	if rest, ok := strings.CutPrefix(id, "SkyrimNet-"); ok && rest != "" {
		return rest
	}
	return id
}

// Playthroughs lists the databases in the SkyrimNet folder, most recently
// played first.
func (s *Store) Playthroughs() []Playthrough {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.playthroughs()
}

// playthroughs is Playthroughs for callers that hold s.mu. If two files
// share an ID, the more recently played one wins.
func (s *Store) playthroughs() []Playthrough {
	out := []Playthrough{}
	if s.dir == "" {
		return out
	}
	dataDir := filepath.Join(s.dir, "data")
	entries, _ := os.ReadDir(dataDir)
	for _, e := range entries {
		if e.IsDir() || !strings.EqualFold(filepath.Ext(e.Name()), ".db") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		path := filepath.Join(dataDir, e.Name())
		out = append(out, Playthrough{
			ID:       playthroughID(e.Name()),
			File:     e.Name(),
			Player:   s.playerName(path, info),
			Size:     info.Size(),
			Modified: info.ModTime().UnixMilli(),
			path:     path,
			modTime:  info.ModTime(),
		})
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].modTime.After(out[j].modTime) })
	return out
}

// path finds the database for a playthrough ID. Every request names its
// playthrough, so tabs open on different playthroughs can't act on each
// other's.
func (s *Store) path(id string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.dir == "" {
		return "", apperr.New(apperr.Unavailable, "No SkyrimNet folder chosen yet")
	}
	for _, p := range s.playthroughs() {
		if p.ID == id {
			return p.path, nil
		}
	}
	return "", apperr.New(apperr.NotFound, "There's no playthrough %q in %s. Reload the page.", id, filepath.Join(s.dir, "data"))
}

// playerName is the name of the player in a save's database, so Hermaeus
// can show "Gram" rather than a save ID. Callers hold s.mu.
func (s *Store) playerName(path string, info os.FileInfo) string {
	if c, ok := s.players[path]; ok && c.modTime.Equal(info.ModTime()) && c.size == info.Size() {
		return c.name
	}
	// A short timeout, since the game may be writing to the database.
	db, err := openTimeout(path, true, 200)
	if err != nil {
		return ""
	}
	defer db.Close()
	var name string
	// The player is always form 0x14 in Skyrim.
	err = db.QueryRow(`select actor_name from uuid_mappings where form_id = 20 and coalesce(actor_name, '') <> ''
		order by updated_at desc limit 1`).Scan(&name)
	if err != nil && !IsNoRows(err) {
		return "" // try again next time
	}
	s.players[path] = playerName{info.ModTime(), info.Size(), name}
	return name
}
