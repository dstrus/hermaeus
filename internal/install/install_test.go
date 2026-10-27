package install

import (
	"os"
	"path/filepath"
	"testing"
)

// skyrimNetFolder makes a SkyrimNet folder with one database at dir.
func skyrimNetFolder(t *testing.T, dir string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Join(dir, "data"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "data", "SkyrimNet-1.db"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestFindsFolderInMO2Overwrite(t *testing.T) {
	// MO2 folder names often include brackets, which file patterns would misread.
	mo2 := filepath.Join(t.TempDir(), "[NoDelete] MO2")
	want := filepath.Join(mo2, "overwrite", "SKSE", "Plugins", "SkyrimNet")
	skyrimNetFolder(t, want)
	if got := findIn([]string{t.TempDir(), mo2}); got != want {
		t.Errorf("findIn = %q, want %q", got, want)
	}
}

func TestFindsNothingWithoutDatabases(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "data"), 0o755)
	if got := findIn([]string{dir}); got != "" {
		t.Errorf("findIn = %q, want none", got)
	}
}

func TestCheckAcceptsDataFolder(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "SkyrimNet")
	skyrimNetFolder(t, dir)
	for _, path := range []string{dir, filepath.Join(dir, "data"), "  " + dir + "  "} {
		if got, err := Check(path); err != nil || got != dir {
			t.Errorf("Check(%q) = %q, %v; want %q", path, got, err, dir)
		}
	}
	for _, path := range []string{"", t.TempDir(), filepath.Join(dir, "missing")} {
		if _, err := Check(path); err == nil {
			t.Errorf("Check(%q) accepted a folder that isn't SkyrimNet's", path)
		}
	}
}
