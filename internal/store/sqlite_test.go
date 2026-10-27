package store

import (
	"net/url"
	"path/filepath"
	"testing"

	"hermaeus/internal/install"
)

func TestSqliteURI(t *testing.T) {
	ro := url.Values{"mode": {"ro"}}
	if got := sqliteURI("/a b/x.db", ro); got != "file:///a%20b/x.db?mode=ro" {
		t.Errorf("unix path: %s", got)
	}
	if got := sqliteURI("C:/My Games/x#1.db", ro); got != "file:///C:/My%20Games/x%231.db?mode=ro" {
		t.Errorf("windows path: %s", got)
	}
}

func TestMissingDatabaseIsNotCreated(t *testing.T) {
	p := filepath.Join(t.TempDir(), "gone.db")
	if db, err := open(p, false); err == nil {
		db.Close()
		t.Error("opened a missing database")
	}
	if install.IsFile(p) {
		t.Error("created an empty database")
	}
}

func TestPlaythroughID(t *testing.T) {
	for file, want := range map[string]string{
		"SkyrimNet-1791685096394-371331.db": "1791685096394-371331",
		"SkyrimNet-.db":                     "SkyrimNet-",
		"backup.db":                         "backup",
	} {
		if got := playthroughID(file); got != want {
			t.Errorf("playthroughID(%q) = %q, want %q", file, got, want)
		}
	}
}
