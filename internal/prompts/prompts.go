// Package prompts is about SkyrimNet's prompts: the templates it fills in to
// instruct the model, for dialogue, diaries, memories, narration, and the
// parts those are built from. Each lives under prompts/ in a content layer,
// and SkyrimNet uses the copy from the highest layer that has it (see the
// layers package). Editing one writes the player's own copy into overlay/,
// where SkyrimNet's dashboard saves too; removing that copy brings back the
// pack or shipped version.
//
// Character bios sit under prompts/characters, but belong to each character,
// so the bios package handles them and they're left out here.
package prompts

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"hermaeus/internal/apperr"
	"hermaeus/internal/layers"
	"hermaeus/internal/prefs"
	"hermaeus/internal/store"
	"hermaeus/internal/textfile"
)

const (
	dir     = "prompts"
	suffix  = ".prompt"
	biosDir = "characters/"
	maxText = 2 << 20 // a prompt bigger than this isn't a prompt
)

type Prompts struct {
	store   *store.Store
	library *layers.Library
	prefs   *prefs.Prefs // where pinned prompts are remembered
}

func New(st *store.Store, lib *layers.Library, p *prefs.Prefs) *Prompts {
	return &Prompts{store: st, library: lib, prefs: p}
}

// Prompt is one prompt, named by its path under prompts/ with forward
// slashes, as the layer in use spells it.
type Prompt struct {
	Path     string `json:"path"`
	Layer    int    `json:"layer"`    // index into PromptList.Layers: the version SkyrimNet uses
	Layers   []int  `json:"layers"`   // every layer that has it, highest priority first
	Modified int64  `json:"modified"` // the version in use, Unix ms
	Pinned   bool   `json:"pinned"`   // kept at the top of the Prompts page
}

type PromptList struct {
	Layers  []layers.Layer `json:"layers"`
	Prompts []Prompt       `json:"prompts"`
}

// Version is one layer's copy of a prompt.
type Version struct {
	Layer    int    `json:"layer"`
	Path     string `json:"path"`
	Modified int64  `json:"modified"` // Unix ms; sent back on save so an outside change isn't overwritten
	Text     string `json:"text"`     // with \n line endings
}

// Document is a prompt in every layer that has it, highest priority first:
// the first version is the one SkyrimNet uses.
type Document struct {
	Layers   []layers.Layer `json:"layers"`
	Versions []Version      `json:"versions"`
}

type Matches struct {
	Paths []string `json:"paths"`
}

func (p *Prompts) layers() ([]layers.Layer, error) {
	lay := p.library.Layers()
	if len(lay) == 0 {
		return nil, apperr.New(apperr.Unavailable, "No SkyrimNet folder chosen yet")
	}
	return lay, nil
}

// List finds every prompt in every layer.
func (p *Prompts) List() (PromptList, error) {
	lay, err := p.layers()
	if err != nil {
		return PromptList{}, err
	}
	out := PromptList{Layers: lay, Prompts: []Prompt{}}
	byKey := map[string]*Prompt{}
	for i, l := range lay {
		root := filepath.Join(l.Root, dir)
		filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil || path == root {
				return nil
			}
			// SkyrimNet's dashboard keeps replaced files in _history.
			if strings.HasPrefix(d.Name(), ".") || (d.IsDir() && d.Name() == "_history") {
				return fs.SkipDir
			}
			if d.IsDir() || !strings.HasSuffix(strings.ToLower(d.Name()), suffix) {
				return nil
			}
			rel, _ := filepath.Rel(root, path)
			rel = filepath.ToSlash(rel)
			key := strings.ToLower(rel)
			if strings.HasPrefix(key, biosDir) {
				return nil
			}
			if e, ok := byKey[key]; ok {
				e.Layers = append(e.Layers, i)
				return nil
			}
			info, err := d.Info()
			if err != nil {
				return nil
			}
			byKey[key] = &Prompt{Path: rel, Layer: i, Layers: []int{i}, Modified: info.ModTime().UnixMilli()}
			return nil
		})
	}
	order := p.pinOrder()
	for _, e := range byKey {
		e.Pinned = order(e.Path) < len(p.prefs.Get().PinnedPrompts)
		out.Prompts = append(out.Prompts, *e)
	}
	// Pinned prompts first, in the order they were pinned; the rest by path.
	sort.Slice(out.Prompts, func(i, j int) bool {
		a, b := out.Prompts[i], out.Prompts[j]
		if x, y := order(a.Path), order(b.Path); x != y {
			return x < y
		}
		return strings.ToLower(a.Path) < strings.ToLower(b.Path)
	})
	return out, nil
}

// Search finds the prompts whose version in use contains q, ignoring case.
func (p *Prompts) Search(q string) (Matches, error) {
	out := Matches{Paths: []string{}}
	q = strings.ToLower(strings.TrimSpace(q))
	if q == "" {
		return out, nil
	}
	list, err := p.List()
	if err != nil {
		return out, err
	}
	for _, pr := range list.Prompts {
		path := layers.Lookup(filepath.Join(list.Layers[pr.Layer].Root, dir), pr.Path)
		if b, err := os.ReadFile(path); err == nil && strings.Contains(strings.ToLower(string(b)), q) {
			out.Paths = append(out.Paths, pr.Path)
		}
	}
	return out, nil
}

// Read returns a prompt in every layer that has it.
func (p *Prompts) Read(path string) (Document, error) {
	path, err := check(path)
	if err != nil {
		return Document{}, err
	}
	lay, err := p.layers()
	if err != nil {
		return Document{}, err
	}
	out := Document{Layers: lay, Versions: []Version{}}
	for i, l := range lay {
		found := layers.Lookup(filepath.Join(l.Root, dir), path)
		if found == "" {
			continue
		}
		info, err := os.Stat(found)
		if err != nil || !info.Mode().IsRegular() {
			continue
		}
		if info.Size() > maxText {
			return Document{}, apperr.New(apperr.Invalid, "%s is too large to be a prompt", path)
		}
		b, err := os.ReadFile(found)
		if err != nil {
			return Document{}, err
		}
		if bytes.IndexByte(b, 0) >= 0 {
			return Document{}, apperr.New(apperr.Invalid, "%s isn't text", path)
		}
		rel, _ := filepath.Rel(filepath.Join(l.Root, dir), found)
		out.Versions = append(out.Versions, Version{Layer: i, Path: filepath.ToSlash(rel), Modified: info.ModTime().UnixMilli(), Text: textfile.Normalize(string(b))})
	}
	if len(out.Versions) == 0 {
		return Document{}, apperr.New(apperr.NotFound, "No layer has a prompt %s", path)
	}
	return out, nil
}

// check validates a prompt path from the page.
func check(path string) (string, error) {
	path = filepath.ToSlash(strings.TrimSpace(path))
	ok := strings.HasSuffix(strings.ToLower(path), suffix) && !filepath.IsAbs(path)
	for _, part := range strings.Split(path, "/") {
		if part == "" || part == "." || part == ".." || strings.ContainsAny(part, `:\`) {
			ok = false
		}
	}
	if !ok {
		return "", apperr.New(apperr.Invalid, "Invalid prompt path %q", path)
	}
	if strings.HasPrefix(strings.ToLower(path), biosDir) {
		return "", apperr.New(apperr.Invalid, "Character bios are edited from the character's Bio tab")
	}
	return path, nil
}

// exists reports whether any layer has prompts/<path>, as a file or folder.
func (p *Prompts) exists(path string) bool {
	for _, l := range p.library.Layers() {
		if layers.Lookup(filepath.Join(l.Root, dir), path) != "" {
			return true
		}
	}
	return false
}
