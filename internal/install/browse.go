package install

import (
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"hermaeus/internal/apperr"
)

// The folder picker lists one folder's subfolders at a time, marking the
// ones that are what it's looking for (a Target), or lead to it.

// Mark says how a folder relates to what the picker is looking for.
type Mark string

const (
	MarkNone   Mark = ""
	MarkHere   Mark = "here"   // it's one
	MarkInside Mark = "inside" // one sits below it, at Folder.Inside
)

type Folder struct {
	Name   string `json:"name"`
	Path   string `json:"path"`
	Mark   Mark   `json:"mark"`
	Inside string `json:"inside,omitempty"` // with MarkInside, the folder it leads to
}

type Listing struct {
	Path    string   `json:"path"`
	Parent  string   `json:"parent"` // "" at the top of a drive
	Mark    Mark     `json:"mark"`   // the listed folder's own
	Folders []Folder `json:"folders"`
	Files   []string `json:"files"` // the names of its files, shown to tell where you are
	Roots   []string `json:"roots"` // where browsing can start over: the drives on Windows, / elsewhere
}

// Browse lists what's in dir, or in the folder Hermaeus runs from when dir
// is "": its subfolders, marked for t, and its files. Hidden ones are left
// out.
func Browse(dir string, t Target) (Listing, error) {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		dir = programDir()
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return Listing{}, apperr.New(apperr.Invalid, "Invalid folder: %v", err)
	}
	entries, err := os.ReadDir(abs)
	if err != nil {
		if !IsDir(abs) {
			return Listing{}, apperr.New(apperr.NotFound, "Folder not found: %s", abs)
		}
		return Listing{}, apperr.New(apperr.Invalid, "Can't open %s: %v", abs, err)
	}
	out := Listing{Path: abs, Mark: t.mark(abs).Mark, Folders: []Folder{}, Files: []string{}, Roots: roots()}
	if parent := filepath.Dir(abs); parent != abs {
		out.Parent = parent
	}
	for _, e := range entries {
		name := e.Name()
		if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "$") {
			continue
		}
		p := filepath.Join(abs, name)
		if !IsDir(p) { // a link to a folder counts as a folder
			out.Files = append(out.Files, name)
			continue
		}
		f := t.mark(p)
		f.Name = name
		out.Folders = append(out.Folders, f)
	}
	sort.Slice(out.Folders, func(i, j int) bool {
		return strings.ToLower(out.Folders[i].Name) < strings.ToLower(out.Folders[j].Name)
	})
	sort.Slice(out.Files, func(i, j int) bool {
		return strings.ToLower(out.Files[i]) < strings.ToLower(out.Files[j])
	})
	return out, nil
}

func (t Target) mark(dir string) Folder {
	f := Folder{Path: dir}
	if t.Is(dir) {
		f.Mark = MarkHere
		return f
	}
	for _, w := range t.Within {
		if inside := filepath.Join(dir, w); t.Is(inside) {
			f.Mark, f.Inside = MarkInside, inside
			return f
		}
	}
	return f
}

func hasDatabases(dir string) bool {
	return len(FilesWithSuffix(filepath.Join(dir, "data"), ".db")) > 0
}

// programDir is the folder Hermaeus runs from, or the current one if that
// can't be told.
func programDir() string {
	if exe, err := os.Executable(); err == nil {
		return filepath.Dir(exe)
	}
	if wd, err := os.Getwd(); err == nil {
		return wd
	}
	return string(filepath.Separator)
}

func roots() []string {
	if runtime.GOOS != "windows" {
		return []string{"/"}
	}
	var out []string
	for d := 'A'; d <= 'Z'; d++ {
		if root := string(d) + `:\`; IsDir(root) {
			out = append(out, root)
		}
	}
	return out
}
