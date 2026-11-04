package install

import (
	"os"
	"path/filepath"
	"testing"

	"hermaeus/internal/apperr"
)

func TestBrowseMarksSkyrimNetFolders(t *testing.T) {
	root := t.TempDir()
	mo2 := filepath.Join(root, "MO2")
	skyrimNetFolder(t, filepath.Join(mo2, "overwrite", "SKSE", "Plugins", "SkyrimNet"))
	skyrimNetFolder(t, filepath.Join(root, "SkyrimNet copy"))
	for _, d := range []string{"apps", ".hidden", "$RECYCLE.BIN"} {
		os.Mkdir(filepath.Join(root, d), 0o755)
	}
	os.WriteFile(filepath.Join(root, "notes.txt"), nil, 0o644)

	got, err := Browse(root, SkyrimNet)
	if err != nil {
		t.Fatal(err)
	}
	want := []Folder{
		{Name: "apps", Path: filepath.Join(root, "apps")},
		{Name: "MO2", Path: mo2, Mark: MarkInside, Inside: filepath.Join(mo2, "overwrite", "SKSE", "Plugins", "SkyrimNet")},
		{Name: "SkyrimNet copy", Path: filepath.Join(root, "SkyrimNet copy"), Mark: MarkHere},
	}
	if len(got.Folders) != len(want) {
		t.Fatalf("Browse listed %+v, want %+v", got.Folders, want)
	}
	for i := range want {
		if got.Folders[i] != want[i] {
			t.Errorf("folder %d = %+v, want %+v", i, got.Folders[i], want[i])
		}
	}
	if got.Path != root || got.Parent != filepath.Dir(root) || got.Mark != MarkNone || len(got.Roots) == 0 {
		t.Errorf("Browse = path %q, parent %q, mark %q, roots %v", got.Path, got.Parent, got.Mark, got.Roots)
	}
}

func TestBrowseMarksTheListedFolder(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "SkyrimNet")
	skyrimNetFolder(t, dir)
	got, err := Browse(dir, SkyrimNet)
	if err != nil || got.Mark != MarkHere {
		t.Errorf("Browse = mark %q, %v; want %q", got.Mark, err, MarkHere)
	}
}

func TestBrowseStartsWhereHermaeusRuns(t *testing.T) {
	got, err := Browse("", SkyrimNet)
	if err != nil || got.Path != programDir() {
		t.Errorf("Browse(\"\") = %q, %v; want %q", got.Path, err, programDir())
	}
}

func TestBrowseRefusesMissingFolder(t *testing.T) {
	_, err := Browse(filepath.Join(t.TempDir(), "gone"), SkyrimNet)
	if kind, ok := apperr.KindOf(err); !ok || kind != apperr.NotFound {
		t.Errorf("Browse of a missing folder = %v, want NotFound", err)
	}
}
