package store_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"
	"time"

	"hermaeus/internal/apperr"
	"hermaeus/internal/install"
	"hermaeus/internal/store"
	"hermaeus/internal/testenv"
)

type harness struct {
	*testenv.Folder
	running bool
	store   *store.Store
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	h := &harness{Folder: testenv.NewFolder(t)}
	h.store = store.New("", func() bool { return h.running })
	if _, err := h.store.SetDir(h.Dir); err != nil {
		t.Fatal(err)
	}
	return h
}

func (h *harness) backups() []string {
	dir := filepath.Join(h.Dir, "hermaeus-backups")
	var files []string
	for _, f := range install.FilesWithSuffix(dir, ".bak") {
		files = append(files, filepath.Join(dir, f))
	}
	return files
}

func renameSerana(name string) func(store.DB) (bool, error) {
	return func(db store.DB) (bool, error) {
		_, err := db.Exec(`update uuid_mappings set actor_name = ? where uuid = `+testenv.SeranaUUID, name)
		return true, err
	}
}

func TestBackupBeforeFirstChangeOnly(t *testing.T) {
	h := newHarness(t)
	for _, name := range []string{"Serana the Bold", "Serana the Wise"} {
		if _, err := store.Write(h.store, testenv.SaveID, "rename", renameSerana(name)); err != nil {
			t.Fatal(err)
		}
	}
	b := h.backups()
	if len(b) != 1 {
		t.Fatalf("backups = %v, want 1", b)
	}
	db, _ := sql.Open("sqlite", b[0])
	defer db.Close()
	var name string
	db.QueryRow(`select actor_name from uuid_mappings where uuid = ` + testenv.SeranaUUID).Scan(&name)
	if name != "Serana" {
		t.Errorf("backup has %q, want the original", name)
	}
}

func TestSavingRefusedWhileSkyrimRuns(t *testing.T) {
	h := newHarness(t)
	h.running = true
	_, err := store.Write(h.store, testenv.SaveID, "rename", renameSerana("x"))
	if k, _ := apperr.KindOf(err); k != apperr.Conflict {
		t.Errorf("write while running: %v", err)
	}
	if len(h.backups()) != 0 {
		t.Errorf("a refused save shouldn't make a backup")
	}
	// Browsing still works.
	if _, err := store.Read(h.store, testenv.SaveID, func(db store.DB) (bool, error) { return store.HasTable(db, "memories") }); err != nil {
		t.Errorf("read while running: %v", err)
	}
}

func TestFailedWriteChangesNothing(t *testing.T) {
	h := newHarness(t)
	_, err := store.Write(h.store, testenv.SaveID, "half a change", func(db store.DB) (bool, error) {
		if _, err := db.Exec(`update uuid_mappings set actor_name = 'Changed'`); err != nil {
			return false, err
		}
		_, err := db.Exec(`update no_such_table set x = 1`)
		return false, err
	})
	if err == nil {
		t.Fatal("write succeeded")
	}
	if n := h.Query(`select actor_name from uuid_mappings where uuid = ` + testenv.SeranaUUID); n != "Serana" {
		t.Errorf("name = %v after a failed write", n)
	}
}

func TestEachRequestNamesItsPlaythrough(t *testing.T) {
	h := newHarness(t)
	other := filepath.Join(h.Dir, "data", "SkyrimNet-1791685096394-371331.db")
	b, _ := os.ReadFile(h.DBPath)
	os.WriteFile(other, b, 0o644)

	if _, err := store.Write(h.store, "1791685096394-371331", "rename", renameSerana("Elsewhere")); err != nil {
		t.Fatal(err)
	}
	if n := h.Query(`select actor_name from uuid_mappings where uuid = ` + testenv.SeranaUUID); n != "Serana" {
		t.Errorf("the other playthrough changed: %v", n)
	}
	for _, id := range []string{"gone", "../SkyrimNet-test", "SkyrimNet-test.db", ""} {
		_, err := store.Read(h.store, id, func(store.DB) (bool, error) { return true, nil })
		if k, _ := apperr.KindOf(err); k != apperr.NotFound {
			t.Errorf("playthrough %q: %v", id, err)
		}
	}
}

func TestNoFolderYet(t *testing.T) {
	s := store.New("", func() bool { return false })
	_, err := store.Read(s, testenv.SaveID, func(store.DB) (bool, error) { return true, nil })
	if k, _ := apperr.KindOf(err); k != apperr.Unavailable {
		t.Errorf("read without a folder: %v", err)
	}
	if p := s.Playthroughs(); len(p) != 0 {
		t.Errorf("playthroughs = %v", p)
	}
}

func TestPlaythroughsShowThePlayersName(t *testing.T) {
	h := newHarness(t)
	player := func() string {
		p := h.store.Playthroughs()
		if len(p) != 1 || p[0].ID != testenv.SaveID {
			t.Fatalf("playthroughs = %+v", p)
		}
		return p[0].Player
	}
	if p := player(); p != "" {
		t.Errorf("player before the save has one = %q", p)
	}
	h.Exec(`insert into uuid_mappings (uuid, form_id, actor_name, updated_at) values (7, 20, 'Adventurer', 1), (8, 20, 'Gram', 2)`)
	if p := player(); p != "Gram" {
		t.Errorf("player = %q, want the most recently updated name", p)
	}
	time.Sleep(10 * time.Millisecond) // so the file's time changes
	h.Exec(`update uuid_mappings set actor_name = 'Gram the Bold', updated_at = 3 where uuid = 8`)
	if p := player(); p != "Gram the Bold" {
		t.Errorf("player after renaming = %q; the cached name wasn't refreshed", p)
	}
}
