// Package plugins reads and changes SkyrimNet's installed plugins: their
// order and which are switched on, kept in content-registry.json, and what
// each one holds. The registry lists plugins highest priority first; a
// higher plugin's prompt replaces a lower one's at the same path, and the
// player's own edits (overlay/) are above them all, SkyrimNet's defaults
// (skyrimnet.base) below.
package plugins

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"hermaeus/internal/apperr"
	"hermaeus/internal/layers"
	"hermaeus/internal/store"
	"hermaeus/internal/textfile"
)

type Plugins struct {
	store   *store.Store
	library *layers.Library
}

func New(s *store.Store, l *layers.Library) *Plugins {
	return &Plugins{store: s, library: l}
}

// Plugin is one installed plugin, or SkyrimNet's defaults, or the player's edits.
type Plugin struct {
	ID          string   `json:"id"`
	Kind        string   `json:"kind"` // base, plugin, new (found, not yet registered), or overlay
	Title       string   `json:"title"`
	Tagline     string   `json:"tagline,omitempty"`
	Description string   `json:"description,omitempty"`
	Author      string   `json:"author,omitempty"`
	Version     string   `json:"version,omitempty"`
	Type        string   `json:"type,omitempty"` // as its manifest says: bundle, prompts, and so on
	Tags        []string `json:"tags"`
	NSFW        bool     `json:"nsfw,omitempty"`
	Source      string   `json:"source,omitempty"` // hub, external, or local
	Enabled     bool     `json:"enabled"`
	Found       bool     `json:"found"` // its folder is in the SkyrimNet folder or SkyrimNet's mod folder
	Folder      string   `json:"folder,omitempty"`
	External    bool     `json:"external"`         // in an external/ folder, as a mod's own plugin is
	ModOff      string   `json:"modOff,omitempty"` // the MO2 mod holding it, switched off in the profile, when it isn't found for that reason
	Contents    []Count  `json:"contents"`         // files by top folder: prompts, filters, knowledge…
	// Prompts it replaces in a layer below, and its prompts a layer above
	// replaces in turn; only for plugins that are switched on.
	Replaces []string `json:"replaces"`
	Replaced []string `json:"replaced"`
}

type Count struct {
	Folder string `json:"folder"`
	Files  int    `json:"files"`
}

// List is every plugin, lowest priority first as SkyrimNet's dashboard
// shows them: its defaults first, then the installed plugins, then the
// player's edits, which override everything.
type List struct {
	Modified int64    `json:"modified"` // the registry's time
	Plugins  []Plugin `json:"plugins"`
}

// Order is the installed plugins as they should be, highest priority first.
type Order struct {
	Modified int64    `json:"modified"`
	Priority []string `json:"priority"`
}

type Switch struct {
	Modified int64 `json:"modified"`
	Enabled  bool  `json:"enabled"`
}

type manifest struct {
	Title       string   `json:"title"`
	Tagline     string   `json:"tagline"`
	Description string   `json:"description"`
	Author      string   `json:"author"`
	Version     string   `json:"version"`
	Type        string   `json:"type"`
	Tags        []string `json:"tags"`
	NSFW        bool     `json:"nsfw"`
}

// registry is content-registry.json, every key kept so it can be written back.
type registry struct {
	raw      map[string]any
	plugins  map[string]map[string]any
	priority []string
	modified int64
}

func (p *Plugins) readRegistry() (registry, error) {
	path := p.library.RegistryPath()
	if path == "" {
		return registry{}, apperr.New(apperr.Unavailable, "No SkyrimNet folder chosen yet")
	}
	reg := registry{raw: map[string]any{}, plugins: map[string]map[string]any{}}
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return reg, nil
	}
	if err != nil {
		return reg, err
	}
	if info, err := os.Stat(path); err == nil {
		reg.modified = info.ModTime().UnixMilli()
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	if err := dec.Decode(&reg.raw); err != nil {
		return reg, apperr.New(apperr.Invalid, "Could not read content-registry.json: %v", err)
	}
	if ps, ok := reg.raw["plugins"].(map[string]any); ok {
		for id, v := range ps {
			if m, ok := v.(map[string]any); ok {
				reg.plugins[id] = m
			}
		}
	}
	if pr, ok := reg.raw["priority"].([]any); ok {
		for _, v := range pr {
			if s, ok := v.(string); ok {
				reg.priority = append(reg.priority, s)
			}
		}
	}
	return reg, nil
}

func (r registry) enabled(id string) bool {
	e, ok := r.plugins[id]["enabled"].(bool)
	return !ok || e
}

// List reads every plugin.
func (p *Plugins) List() (List, error) {
	reg, err := p.readRegistry()
	if err != nil {
		return List{}, err
	}
	out := List{Modified: reg.modified, Plugins: []Plugin{}}
	base := p.describe(layers.BasePlugin, "base", p.library.PluginDir(layers.BasePlugin))
	base.Title, base.Enabled = "SkyrimNet’s defaults", true
	out.Plugins = append(out.Plugins, base)
	// A plugin folder the registry doesn't have yet: SkyrimNet adds it, at
	// the bottom of its order, the next time it starts.
	listed := map[string]bool{strings.ToLower(layers.BasePlugin): true}
	for _, id := range reg.priority {
		listed[strings.ToLower(id)] = true
	}
	for id := range reg.plugins {
		listed[strings.ToLower(id)] = true
	}
	for _, id := range p.library.PluginIDs() {
		if !listed[strings.ToLower(id)] {
			pl := p.describe(id, "new", p.library.PluginDir(id))
			out.Plugins = append(out.Plugins, pl)
		}
	}
	for i := len(reg.priority) - 1; i >= 0; i-- {
		id := reg.priority[i]
		if id == layers.BasePlugin {
			continue
		}
		pl := p.describe(id, "plugin", p.library.PluginDir(id))
		pl.Enabled = reg.enabled(id)
		pl.Source, _ = reg.plugins[id]["source"].(string)
		out.Plugins = append(out.Plugins, pl)
	}
	if lay := p.library.Layers(); len(lay) > 0 && lay[0].Kind == "overlay" {
		ov := p.describe("overlay", "overlay", lay[0].Root)
		ov.Title, ov.Enabled, ov.Description = "Your edits", true, "Files you’ve edited or made in SkyrimNet’s dashboard or here, which aren’t in a plugin. They override every plugin."
		out.Plugins = append(out.Plugins, ov)
	}
	p.overrides(out.Plugins)
	return out, nil
}

// describe reads a plugin's manifest and counts its files.
func (p *Plugins) describe(id, kind, dir string) Plugin {
	pl := Plugin{ID: id, Kind: kind, Title: id, Tags: []string{}, Contents: []Count{}, Replaces: []string{}, Replaced: []string{}}
	if dir == "" {
		pl.ModOff = p.library.SwitchedOffMod(id)
		return pl
	}
	pl.Found = true
	pl.Folder = dir
	// Shown from the SkyrimNet folder, or under MO2 from the instance: mods/…
	for _, base := range []string{p.store.Dir(), p.library.MO2Root()} {
		if rel, err := filepath.Rel(base, dir); base != "" && err == nil && !strings.HasPrefix(rel, "..") {
			pl.Folder = filepath.ToSlash(rel)
			break
		}
	}
	pl.External = strings.Contains(filepath.ToSlash(dir), "/external/")
	if b, err := os.ReadFile(filepath.Join(dir, "manifest.json")); err == nil {
		var mf manifest
		if json.Unmarshal(b, &mf) == nil {
			if mf.Title != "" {
				pl.Title = mf.Title
			}
			pl.Tagline, pl.Description, pl.Author, pl.Version, pl.Type, pl.NSFW = mf.Tagline, mf.Description, mf.Author, mf.Version, mf.Type, mf.NSFW
			if mf.Tags != nil {
				pl.Tags = mf.Tags
			}
		}
	}
	entries, _ := os.ReadDir(dir)
	for _, e := range entries {
		if !e.IsDir() || strings.HasPrefix(e.Name(), "_") || strings.HasPrefix(e.Name(), ".") {
			continue
		}
		// Character bios sit in prompts/characters but are a kind of their
		// own, shown on the bio pages, so they're counted apart.
		n, bios := 0, 0
		filepath.WalkDir(filepath.Join(dir, e.Name()), func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			rel, _ := filepath.Rel(filepath.Join(dir, e.Name()), path)
			if strings.EqualFold(e.Name(), "prompts") && strings.HasPrefix(strings.ToLower(filepath.ToSlash(rel)), "characters/") {
				bios++
			} else {
				n++
			}
			return nil
		})
		if n > 0 {
			pl.Contents = append(pl.Contents, Count{Folder: e.Name(), Files: n})
		}
		if bios > 0 {
			pl.Contents = append(pl.Contents, Count{Folder: "bios", Files: bios})
		}
	}
	return pl
}

// overrides works out, for each plugin switched on, which of its prompts
// replace a lower layer's and which a higher layer's replace.
func (p *Plugins) overrides(list []Plugin) {
	sets := make([]map[string]string, len(list)) // folded path → as spelled
	dirs := make([]string, len(list))
	for i, pl := range list {
		switch pl.Kind {
		case "overlay":
			if lay := p.library.Layers(); len(lay) > 0 {
				dirs[i] = lay[0].Root
			}
		default:
			dirs[i] = p.library.PluginDir(pl.ID)
		}
	}
	for i, pl := range list {
		sets[i] = map[string]string{}
		if !pl.Enabled || !pl.Found {
			continue
		}
		prompts := layers.Lookup(dirs[i], "prompts")
		if prompts == "" {
			continue
		}
		filepath.WalkDir(prompts, func(path string, d fs.DirEntry, err error) error {
			if err == nil && !d.IsDir() {
				rel, _ := filepath.Rel(prompts, path)
				rel = filepath.ToSlash(rel)
				// Bios layer the same way, but have their own pages.
				if !strings.HasPrefix(strings.ToLower(rel), "characters/") {
					sets[i][strings.ToLower(rel)] = rel
				}
			}
			return nil
		})
	}
	for i := range list {
		for key, rel := range sets[i] {
			for j := 0; j < i; j++ {
				if _, ok := sets[j][key]; ok {
					list[i].Replaces = append(list[i].Replaces, rel)
					break
				}
			}
			for j := i + 1; j < len(list); j++ {
				if _, ok := sets[j][key]; ok {
					list[i].Replaced = append(list[i].Replaced, rel)
					break
				}
			}
		}
		sort.Strings(list[i].Replaces)
		sort.Strings(list[i].Replaced)
	}
}

// SetOrder puts the installed plugins in a new order, highest priority first.
func (p *Plugins) SetOrder(o Order) (List, error) {
	err := p.change("Reordered the plugins", o.Modified, func(reg *registry) error {
		have := map[string]bool{}
		for _, id := range reg.priority {
			have[id] = true
		}
		if len(o.Priority) != len(reg.priority) {
			return apperr.New(apperr.Conflict, "The plugins changed after you opened them. Reload to see the latest, then try again.")
		}
		seen := map[string]bool{}
		for _, id := range o.Priority {
			if !have[id] || seen[id] {
				return apperr.New(apperr.Conflict, "The plugins changed after you opened them. Reload to see the latest, then try again.")
			}
			seen[id] = true
		}
		order := make([]any, len(o.Priority))
		for i, id := range o.Priority {
			order[i] = id
		}
		reg.raw["priority"] = order
		return nil
	})
	if err != nil {
		return List{}, err
	}
	return p.List()
}

// SetEnabled switches a plugin on or off.
func (p *Plugins) SetEnabled(id string, s Switch) (List, error) {
	what := "Switched on the plugin " + id
	if !s.Enabled {
		what = "Switched off the plugin " + id
	}
	err := p.change(what, s.Modified, func(reg *registry) error {
		rec, ok := reg.plugins[id]
		if !ok {
			if id == layers.BasePlugin {
				return apperr.New(apperr.Invalid, "SkyrimNet’s defaults can’t be switched off")
			}
			return apperr.New(apperr.NotFound, "There’s no installed plugin %q", id)
		}
		rec["enabled"] = s.Enabled
		rec["enabled_explicit"] = true
		return nil
	})
	if err != nil {
		return List{}, err
	}
	return p.List()
}

// change edits the registry, refusing if it changed since modified, and
// writes it back as SkyrimNet does: sorted keys, two-space indents.
func (p *Plugins) change(what string, modified int64, edit func(*registry) error) error {
	return p.store.ChangeFiles(what, func() error {
		reg, err := p.readRegistry()
		if err != nil {
			return err
		}
		if reg.modified != modified {
			return apperr.New(apperr.Conflict, "The plugins changed after you opened them. Reload to see the latest, then try again.")
		}
		path := p.library.RegistryPath()
		old, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := edit(&reg); err != nil {
			return err
		}
		var buf bytes.Buffer
		enc := json.NewEncoder(&buf)
		enc.SetEscapeHTML(false)
		enc.SetIndent("", "  ")
		if err := enc.Encode(reg.raw); err != nil {
			return err
		}
		out := buf.Bytes()
		if !bytes.HasSuffix(old, []byte("\n")) {
			out = bytes.TrimRight(out, "\n")
		}
		if _, err := textfile.Backup(filepath.Join(p.store.BackupDir(), "content-registry.json"), old, 0o644); err != nil {
			return apperr.New(apperr.Conflict, "Could not back up content-registry.json, so nothing was changed: %v", err)
		}
		return textfile.WriteAtomic(path, out, 0o644)
	})
}
