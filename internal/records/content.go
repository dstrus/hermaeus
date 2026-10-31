package records

import (
	"bytes"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"hermaeus/internal/apperr"
	"hermaeus/internal/layers"

	"go.yaml.in/yaml/v3"
)

// Besides prompts, SkyrimNet's content is YAML records, one per file, in a
// folder per kind. Like a prompt, a record can come from any layer, and the
// player's edits are copies in overlay/, which replace the same file below.

// Kind is a folder of records.
type Kind struct {
	Root  string `json:"root"`
	Label string `json:"label"`
}

var Kinds = []Kind{
	{"triggers", "Triggers"}, {"actions", "Actions"}, {"filters", "Filters"}, {"voice_effects", "Voice effects"},
	{"entities", "Entities"}, {"dialogue_actions", "Dialogue actions"}, {"translator", "Speech rules"},
	{"items", "Items"}, {"spells", "Spells"}, {"furniture", "Furniture names"}, {"identity", "Identity links"}, {"places", "Named places"},
}

func isKind(root string) bool {
	for _, k := range Kinds {
		if k.Root == root {
			return true
		}
	}
	return false
}

// Provider is one layer's copy of a file.
type Provider struct {
	Layer    int    `json:"layer"` // index into the layers listed with it
	Path     string `json:"path"`  // as spelled in that layer, forward slashes
	Size     int64  `json:"size"`
	Modified int64  `json:"modified"` // Unix ms
}

// Record is one record, the highest layer's copy in use.
type Record struct {
	Path      string     `json:"path"` // e.g. triggers/PlayerDiaryExample.yaml
	Kind      string     `json:"kind"`
	Name      string     `json:"name"`
	Enabled   bool       `json:"enabled"`
	Switch    bool       `json:"switch"`    // it says whether it's on, so it can be switched
	Providers []Provider `json:"providers"` // highest priority first; the first is used
	Error     string     `json:"error,omitempty"`
}

type ContentList struct {
	Layers  []layers.Layer `json:"layers"`
	Kinds   []Kind         `json:"kinds"`
	Records []Record       `json:"records"` // by kind, then path
}

// Override is a file a layer above SkyrimNet's defaults replaces or, in the
// player's own edits, adds.
type Override struct {
	Path      string     `json:"path"`
	Providers []Provider `json:"providers"` // highest priority first; the first wins
	// The player's copy is older than the file it replaces, so it may
	// predate changes made since. (Shipped files carry the release's date,
	// so only the player's own edits are flagged.)
	Stale bool `json:"stale"`
	// Two or more plugins provide it, so all but one are ignored.
	Conflict bool `json:"conflict"`
}

type OverrideList struct {
	Layers    []layers.Layer `json:"layers"`
	Overrides []Override     `json:"overrides"`
}

// Version is one layer's copy of a file, as text when it is text.
type Version struct {
	Provider `tstype:",extends"`
	Kind     string  `json:"kind"` // the layer's: overlay, pack, or base
	Source   string  `json:"source"`
	Text     *string `json:"text"` // nil for a binary or very large file; \n line endings
	CRLF     bool    `json:"crlf"`
}

const maxText = 2 << 20

// skip is bookkeeping that isn't content.
func skip(rel string, d fs.DirEntry) bool {
	name := d.Name()
	if strings.HasPrefix(name, ".") || (d.IsDir() && name == "_history") {
		return true
	}
	return !d.IsDir() && !strings.Contains(rel, "/") && (name == "manifest.json" || name == "install.json")
}

// layerFiles lists a folder's files by folded path, as Windows compares them.
func layerFiles(root string, layer int) map[string]Provider {
	files := map[string]Provider{}
	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil || path == root {
			return nil
		}
		rel, _ := filepath.Rel(root, path)
		rel = filepath.ToSlash(rel)
		if skip(rel, d) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		files[strings.ToLower(rel)] = Provider{Layer: layer, Path: rel, Size: info.Size(), Modified: info.ModTime().UnixMilli()}
		return nil
	})
	return files
}

func (r *Records) layerList() []layers.Layer {
	lay := r.library.Layers()
	if lay == nil {
		lay = []layers.Layer{}
	}
	return lay
}

// Content lists every record in any layer.
func (r *Records) Content() ContentList {
	lay := r.layerList()
	byPath := map[string][]Provider{}
	for i, l := range lay {
		for _, k := range Kinds {
			dir := layers.Lookup(l.Root, k.Root)
			if dir == "" {
				continue
			}
			for key, f := range layerFiles(dir, i) {
				if !strings.HasSuffix(f.Path, recordSuffix(k.Root)) {
					continue
				}
				f.Path = k.Root + "/" + f.Path
				byPath[k.Root+"/"+key] = append(byPath[k.Root+"/"+key], f)
			}
		}
	}
	out := ContentList{Layers: lay, Kinds: Kinds, Records: make([]Record, 0, len(byPath))}
	for _, providers := range byPath {
		top := providers[0]
		kind, _, _ := strings.Cut(top.Path, "/")
		rec := Record{Path: top.Path, Kind: kind, Providers: providers, Enabled: true}
		stem := strings.TrimSuffix(filepath.Base(top.Path), filepath.Ext(top.Path))
		rec.Name = stem
		if b, err := os.ReadFile(layers.Lookup(lay[top.Layer].Root, top.Path)); err == nil {
			var m map[string]any
			if err := yaml.Unmarshal(b, &m); err != nil {
				rec.Error = yamlError(err)
			} else {
				for _, k := range []string{"name", "entityName", "key", "id"} {
					if s, ok := m[k].(string); ok && strings.TrimSpace(s) != "" {
						rec.Name = s
						break
					}
				}
				if e, ok := m["enabled"].(bool); ok {
					rec.Enabled, rec.Switch = e, true
				} else if kind == "triggers" || kind == "actions" || kind == "voice_effects" {
					rec.Switch = true // on unless it says otherwise; switching adds the key
				}
			}
		}
		out.Records = append(out.Records, rec)
	}
	order := map[string]int{}
	for i, k := range Kinds {
		order[k.Root] = i
	}
	sort.Slice(out.Records, func(i, j int) bool {
		a, b := out.Records[i], out.Records[j]
		if a.Kind != b.Kind {
			return order[a.Kind] < order[b.Kind]
		}
		return strings.ToLower(a.Path) < strings.ToLower(b.Path)
	})
	return out
}

// Overrides lists the files that a layer above SkyrimNet's defaults
// replaces, or that the player's own edits add. Files a plugin adds without
// replacing anything, such as most bio packs' bios, are left out.
func (r *Records) Overrides() OverrideList {
	lay := r.layerList()
	byPath := map[string][]Provider{}
	for i, l := range lay {
		for key, f := range layerFiles(l.Root, i) {
			byPath[key] = append(byPath[key], f)
		}
	}
	out := OverrideList{Layers: lay, Overrides: []Override{}}
	for _, providers := range byPath {
		winner := lay[providers[0].Layer]
		if winner.Kind == "base" || (len(providers) == 1 && winner.Kind != "overlay") {
			continue
		}
		o := Override{Path: providers[0].Path, Providers: providers}
		if len(providers) > 1 && winner.Kind == "overlay" {
			o.Stale = providers[1].Modified > providers[0].Modified
		}
		packs := 0
		for _, p := range providers {
			if lay[p.Layer].Kind == "pack" {
				packs++
			}
		}
		o.Conflict = packs > 1
		out.Overrides = append(out.Overrides, o)
	}
	sort.Slice(out.Overrides, func(i, j int) bool {
		return strings.ToLower(out.Overrides[i].Path) < strings.ToLower(out.Overrides[j].Path)
	})
	return out
}

// checkPath refuses a path that could leave a layer's folder.
func checkPath(rel string) (string, error) {
	rel = filepath.ToSlash(strings.TrimSpace(rel))
	bad := rel == "" || strings.HasPrefix(rel, "/") || filepath.IsAbs(rel) || strings.Contains(rel, ":") || strings.Contains(rel, `\`)
	for _, part := range strings.Split(rel, "/") {
		if part == "" || part == "." || part == ".." {
			bad = true
		}
	}
	if bad {
		return "", apperr.New(apperr.Invalid, "Invalid path %q", rel)
	}
	return rel, nil
}

// Versions lists every layer's copy of the file at rel, highest priority first.
func (r *Records) Versions(rel string) ([]Version, error) {
	rel, err := checkPath(rel)
	if err != nil {
		return nil, err
	}
	out := []Version{}
	for i, l := range r.layerList() {
		path := layers.Lookup(l.Root, rel)
		if path == "" {
			continue
		}
		info, err := os.Stat(path)
		if err != nil || !info.Mode().IsRegular() {
			continue
		}
		here, _ := filepath.Rel(l.Root, path)
		v := Version{Provider: Provider{Layer: i, Path: filepath.ToSlash(here), Size: info.Size(), Modified: info.ModTime().UnixMilli()}, Kind: l.Kind, Source: l.Source}
		if info.Size() <= maxText {
			if b, err := os.ReadFile(path); err == nil && !bytes.Contains(b, []byte{0}) {
				v.CRLF = bytes.Contains(b, []byte("\r\n"))
				text := strings.ReplaceAll(string(b), "\r\n", "\n")
				v.Text = &text
			}
		}
		out = append(out, v)
	}
	if len(out) == 0 {
		return nil, apperr.New(apperr.NotFound, "No layer has %s", rel)
	}
	return out, nil
}
