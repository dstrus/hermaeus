// Package layers is SkyrimNet's content library: the layers a prompt, bio or
// other content file can come from, highest priority first. Per SkyrimNet's
// own docs (MIGRATING_TO_BETA25.md): the player's dashboard edits (overlay/),
// installed plugins in the player's order, then SkyrimNet's shipped defaults
// (the skyrimnet.base plugin). Per-save files sit above all of these, but only
// cover bios, and the bios package handles them.
//
// With Mod Organizer 2, SkyrimNet's own files, skyrimnet.base among them,
// stay in its mod folder while everything SkyrimNet writes lands in overwrite,
// so both folders are searched.
package layers

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"hermaeus/internal/apperr"
	"hermaeus/internal/install"
)

const (
	BasePlugin   = "skyrimnet.base"
	registryFile = "content-registry.json"
)

// Library finds SkyrimNet's content, in the SkyrimNet folder and, under
// MO2, in SkyrimNet's mod folder.
type Library struct {
	skyrimNetDir func() string
	mu           sync.Mutex
	modDir       string // chosen; "" means find it
}

func New(skyrimNetDir func() string) *Library {
	return &Library{skyrimNetDir: skyrimNetDir}
}

// Layer is one place content comes from.
type Layer struct {
	ID     string `json:"id"`
	Kind   string `json:"kind"`   // overlay, pack, or base
	Source string `json:"source"` // for people: "your dashboard edits", a plugin's title…
	Root   string `json:"-"`      // the folder holding prompts/, triggers/ and so on
}

// HasBase reports whether dir holds SkyrimNet's shipped defaults.
func HasBase(dir string) bool {
	return install.IsDir(filepath.Join(dir, "library", BasePlugin, "prompts"))
}

// ModDir is SkyrimNet's mod folder: the one chosen, or one found next to an
// MO2 overwrite folder, or "" when the SkyrimNet folder has the defaults
// itself. detected says it was found rather than chosen.
func (l *Library) ModDir() (dir string, detected bool) {
	l.mu.Lock()
	dir = l.modDir
	l.mu.Unlock()
	sky := l.skyrimNetDir()
	if dir != "" || sky == "" || HasBase(sky) {
		return dir, false
	}
	return detectModDir(sky), true
}

// SetModDir points Hermaeus at SkyrimNet's mod folder, and returns the
// folder it will use. "" goes back to finding it. The mod's top folder or
// its SKSE/Plugins/SkyrimNet both work.
func (l *Library) SetModDir(dir string) (string, error) {
	dir = strings.TrimSpace(dir)
	if dir != "" {
		abs, err := filepath.Abs(dir)
		if err != nil {
			return "", apperr.New(apperr.Invalid, "Invalid folder: %v", err)
		}
		found := ""
		for _, c := range []string{abs, filepath.Join(abs, "SKSE", "Plugins", "SkyrimNet")} {
			if HasBase(c) {
				found = c
				break
			}
		}
		if found == "" {
			return "", apperr.New(apperr.Invalid, "No SkyrimNet defaults in %s. Choose SkyrimNet's mod folder (it contains SKSE\\Plugins\\SkyrimNet\\library\\%s).", abs, BasePlugin)
		}
		dir = found
	}
	l.mu.Lock()
	l.modDir = dir
	l.mu.Unlock()
	return dir, nil
}

// HasDefaults reports whether SkyrimNet's shipped defaults can be found.
func (l *Library) HasDefaults() bool {
	sky := l.skyrimNetDir()
	if sky != "" && HasBase(sky) {
		return true
	}
	mod, _ := l.ModDir()
	return mod != "" && HasBase(mod)
}

// detectModDir looks for SkyrimNet in the mods folder of the MO2 instance
// whose overwrite folder is skyrimNetDir. If several mods contain it (say,
// two versions), the newest by its manifest wins.
func detectModDir(skyrimNetDir string) string {
	parts := strings.Split(filepath.ToSlash(skyrimNetDir), "/")
	n := len(parts)
	if n < 5 || !strings.EqualFold(parts[n-4], "overwrite") {
		return ""
	}
	root := filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(skyrimNetDir))))
	mods, _ := os.ReadDir(filepath.Join(root, "mods"))
	best, bestVersion := "", ""
	for _, mod := range mods {
		dir := filepath.Join(root, "mods", mod.Name(), "SKSE", "Plugins", "SkyrimNet")
		mf, err := readManifest(filepath.Join(dir, "library", BasePlugin))
		if err != nil {
			continue
		}
		if best == "" || compareVersions(mf.Version, bestVersion) > 0 {
			best, bestVersion = dir, mf.Version
		}
	}
	return best
}

// compareVersions compares dotted version numbers such as 0.25.3.
func compareVersions(a, b string) int {
	as, bs := strings.Split(a, "."), strings.Split(b, ".")
	for i := 0; i < len(as) || i < len(bs); i++ {
		var x, y int
		if i < len(as) {
			fmt.Sscan(as[i], &x)
		}
		if i < len(bs) {
			fmt.Sscan(bs[i], &y)
		}
		if x != y {
			return x - y
		}
	}
	return 0
}

// Roots are the folders that can hold library/ and external/ plugin
// folders: the SkyrimNet folder, and SkyrimNet's mod folder under MO2.
func (l *Library) Roots() []string {
	sky := l.skyrimNetDir()
	if sky == "" {
		return nil
	}
	roots := []string{sky}
	if mod, _ := l.ModDir(); mod != "" {
		roots = append(roots, mod)
	}
	return roots
}

// Layers lists the content layers, highest priority first: the player's
// dashboard edits, enabled plugins in the player's order, then SkyrimNet's
// shipped defaults.
func (l *Library) Layers() []Layer {
	sky := l.skyrimNetDir()
	if sky == "" {
		return nil
	}
	roots := l.Roots()
	layers := []Layer{{ID: "overlay", Kind: "overlay", Source: "your dashboard edits", Root: filepath.Join(sky, "overlay")}}
	reg, err := readRegistry(sky)
	if err != nil {
		log.Printf("Could not read %s, so installed plugins are skipped: %v", registryFile, err)
	}
	for _, id := range reg.Priority {
		if !reg.enabled(id) || id == BasePlugin {
			continue
		}
		if dir := pluginDir(roots, id); dir != "" {
			layers = append(layers, Layer{ID: id, Kind: "pack", Source: pluginTitle(dir, id), Root: dir})
		}
	}
	if dir := pluginDir(roots, BasePlugin); dir != "" {
		layers = append(layers, Layer{ID: BasePlugin, Kind: "base", Source: "SkyrimNet’s defaults", Root: dir})
	}
	return layers
}

// pluginDir is the folder SkyrimNet uses for a plugin: of every copy in the
// roots' library/ and external/ folders, the highest version, and on a tie a
// library copy over an external one.
func pluginDir(roots []string, id string) string {
	type copy struct {
		dir, version string
	}
	var copies []copy
	for _, kind := range []string{"library", "external"} {
		for _, root := range roots {
			if d := filepath.Join(root, kind, id); install.IsDir(d) {
				mf, _ := readManifest(d)
				copies = append(copies, copy{d, mf.Version})
			}
		}
	}
	if len(copies) == 0 {
		return ""
	}
	sort.SliceStable(copies, func(i, j int) bool { return compareVersions(copies[i].version, copies[j].version) > 0 })
	return copies[0].dir
}

type manifest struct {
	Title   string `json:"title"`
	Version string `json:"version"`
}

func readManifest(dir string) (manifest, error) {
	var mf manifest
	b, err := os.ReadFile(filepath.Join(dir, "manifest.json"))
	if err != nil {
		return mf, err
	}
	return mf, json.Unmarshal(b, &mf)
}

func pluginTitle(dir, id string) string {
	if mf, err := readManifest(dir); err == nil && mf.Title != "" {
		return mf.Title
	}
	return id
}

// registry is the part of content-registry.json that decides plugin order.
type registry struct {
	Plugins map[string]struct {
		Enabled *bool `json:"enabled"`
	} `json:"plugins"`
	Priority []string `json:"priority"`
}

// A plugin without an entry, or without "enabled", is enabled.
func (r registry) enabled(id string) bool {
	p, ok := r.Plugins[id]
	return !ok || p.Enabled == nil || *p.Enabled
}

func readRegistry(sky string) (registry, error) {
	var reg registry
	b, err := os.ReadFile(filepath.Join(sky, registryFile))
	if os.IsNotExist(err) {
		return reg, nil
	}
	if err != nil {
		return reg, err
	}
	return reg, json.Unmarshal(b, &reg)
}
