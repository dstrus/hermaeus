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

// ModFolder is SkyrimNet's mod folder, for the folder picker: the mod's top
// folder or its SKSE/Plugins/SkyrimNet, as SetModDir takes either.
var ModFolder = install.Target{Is: func(dir string) bool {
	return HasBase(dir) || HasBase(filepath.Join(dir, "SKSE", "Plugins", "SkyrimNet"))
}}

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
	root := mo2Root(skyrimNetDir)
	if root == "" {
		return ""
	}
	best, bestVersion := "", ""
	for _, mod := range mo2Mods(root) {
		dir := filepath.Join(root, "mods", mod, "SKSE", "Plugins", "SkyrimNet")
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
// folders: the SkyrimNet folder, SkyrimNet's mod folder under MO2, and,
// under MO2, every other mod's SkyrimNet folder, as the game sees them all
// merged. That's where a mod's own SkyrimNet plugin (external/) lives.
func (l *Library) Roots() []string {
	sky := l.skyrimNetDir()
	if sky == "" {
		return nil
	}
	roots := []string{sky}
	seen := map[string]bool{filepath.Clean(sky): true}
	add := func(dir string) {
		if dir != "" && !seen[filepath.Clean(dir)] {
			seen[filepath.Clean(dir)] = true
			roots = append(roots, dir)
		}
	}
	mod, _ := l.ModDir()
	add(mod)
	if root := mo2Root(sky); root != "" {
		for _, m := range mo2Mods(root) {
			if d := filepath.Join(root, "mods", m, "SKSE", "Plugins", "SkyrimNet"); install.IsDir(d) {
				add(d)
			}
		}
	}
	return roots
}

// MO2Root is the MO2 instance the SkyrimNet folder is in, or "".
func (l *Library) MO2Root() string {
	if sky := l.skyrimNetDir(); sky != "" {
		return mo2Root(sky)
	}
	return ""
}

// mo2Root is the MO2 instance whose overwrite folder holds skyrimNetDir
// (overwrite/SKSE/Plugins/SkyrimNet), or "".
func mo2Root(skyrimNetDir string) string {
	parts := strings.Split(filepath.ToSlash(filepath.Clean(skyrimNetDir)), "/")
	n := len(parts)
	if n < 5 || !strings.EqualFold(parts[n-4], "overwrite") {
		return ""
	}
	return filepath.Dir(filepath.Dir(filepath.Dir(filepath.Dir(skyrimNetDir))))
}

// mo2Mods lists an MO2 instance's mods, by folder name: those the selected
// profile switches on, when the profile can be read, or else every mod. A
// profile that names none of the mods on disk belongs to another layout, so
// it's ignored rather than hiding everything.
func mo2Mods(root string) []string {
	all := mo2AllMods(root)
	on, ok := mo2Enabled(root)
	if !ok {
		return all
	}
	var out []string
	named := false
	for _, m := range all {
		_, listed := on[strings.ToLower(m)]
		named = named || listed
		if on[strings.ToLower(m)] {
			out = append(out, m)
		}
	}
	if !named {
		return all
	}
	return out
}

func mo2AllMods(root string) []string {
	entries, _ := os.ReadDir(filepath.Join(root, "mods"))
	var all []string
	for _, e := range entries {
		if e.IsDir() {
			all = append(all, e.Name())
		}
	}
	return all
}

// SwitchedOffMod names the MO2 mod holding a plugin when the profile has
// that mod switched off, so SkyrimNet can't see the plugin; "" otherwise.
func (l *Library) SwitchedOffMod(id string) string {
	root := l.MO2Root()
	if root == "" {
		return ""
	}
	on := map[string]bool{}
	for _, m := range mo2Mods(root) {
		on[m] = true
	}
	for _, m := range mo2AllMods(root) {
		if on[m] {
			continue
		}
		for _, kind := range []string{"library", "external"} {
			if install.IsDir(filepath.Join(root, "mods", m, "SKSE", "Plugins", "SkyrimNet", kind, id)) {
				return m
			}
		}
	}
	return ""
}

// mo2Enabled reads the selected profile's modlist.txt. ok is false when
// there's no profile to read.
func mo2Enabled(root string) (on map[string]bool, ok bool) {
	profile := ""
	if b, err := os.ReadFile(filepath.Join(root, "ModOrganizer.ini")); err == nil {
		for _, line := range strings.Split(strings.ReplaceAll(string(b), "\r\n", "\n"), "\n") {
			if v, found := strings.CutPrefix(strings.TrimSpace(line), "selected_profile="); found {
				v = strings.TrimSuffix(strings.TrimPrefix(v, "@ByteArray("), ")")
				profile = strings.Trim(v, `"`)
			}
		}
	}
	if profile == "" {
		return nil, false
	}
	b, err := os.ReadFile(filepath.Join(root, "profiles", profile, "modlist.txt"))
	if err != nil {
		return nil, false
	}
	// Listed mods by folder name, folded as Windows compares them: true for
	// "+Name", switched on, and false for "-Name", switched off.
	on = map[string]bool{}
	for _, line := range strings.Split(strings.ReplaceAll(string(b), "\r\n", "\n"), "\n") {
		line = strings.TrimRight(line, "\r")
		if name, found := strings.CutPrefix(line, "+"); found {
			on[strings.ToLower(name)] = true
		} else if name, found := strings.CutPrefix(line, "-"); found {
			on[strings.ToLower(name)] = false
		}
	}
	return on, true
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

// PluginIDs lists every plugin folder SkyrimNet reads in the roots' library/
// and external/ folders, whether or not the registry has it yet.
func (l *Library) PluginIDs() []string {
	seen := map[string]bool{}
	var out []string
	for _, root := range l.Roots() {
		for _, kind := range []string{"library", "external"} {
			entries, _ := os.ReadDir(filepath.Join(root, kind))
			for _, e := range entries {
				if e.IsDir() && installSource(filepath.Join(root, kind, e.Name()), kind) != "" && !seen[strings.ToLower(e.Name())] {
					seen[strings.ToLower(e.Name())] = true
					out = append(out, e.Name())
				}
			}
		}
	}
	sort.Strings(out)
	return out
}

// PluginDir is the folder SkyrimNet uses for a plugin, or "" when it isn't
// in any of the roots (an external plugin in another mod's folder, say).
func (l *Library) PluginDir(id string) string {
	return pluginDir(l.Roots(), id)
}

// RegistryPath is content-registry.json in the SkyrimNet folder, which keeps
// the installed plugins, their order, and which are switched on.
func (l *Library) RegistryPath() string {
	sky := l.skyrimNetDir()
	if sky == "" {
		return ""
	}
	return filepath.Join(sky, registryFile)
}

// pluginDir is the folder SkyrimNet uses for a plugin: of every copy in the
// roots' library/ and external/ folders, the highest version, and on a tie a
// copy the player packaged, then an external one, then one from the Hub.
func pluginDir(roots []string, id string) string {
	type copy struct {
		dir, version string
		rank         int
	}
	rank := map[string]int{"local": 0, "external": 1}
	var copies []copy
	for _, kind := range []string{"library", "external"} {
		for _, root := range roots {
			d := filepath.Join(root, kind, id)
			if !install.IsDir(d) {
				continue
			}
			src := installSource(d, kind)
			if src == "" {
				continue
			}
			r, ok := rank[src]
			if !ok {
				r = 2
			}
			mf, _ := readManifest(d)
			copies = append(copies, copy{d, mf.Version, r})
		}
	}
	if len(copies) == 0 {
		return ""
	}
	sort.SliceStable(copies, func(i, j int) bool {
		if c := compareVersions(copies[i].version, copies[j].version); c != 0 {
			return c > 0
		}
		return copies[i].rank < copies[j].rank
	})
	return copies[0].dir
}

// installSource says where a plugin folder came from: "external" for one in
// external/, and for one in library/ the source SkyrimNet recorded in its
// install.json ("hub" or "local"). It's "" for a library/ folder SkyrimNet
// didn't install, which SkyrimNet ignores. SkyrimNet's own defaults come
// with it, without an install.json.
func installSource(dir, kind string) string {
	if kind == "external" {
		return "external"
	}
	if strings.EqualFold(filepath.Base(dir), BasePlugin) {
		return "base"
	}
	b, err := os.ReadFile(filepath.Join(dir, "install.json"))
	if err != nil {
		return ""
	}
	var in struct {
		Source string `json:"source"`
	}
	json.Unmarshal(b, &in)
	if in.Source == "" {
		return "hub"
	}
	return in.Source
}

// Lookup finds rel (forward slashes) under root the way SkyrimNet does on
// Windows, ignoring case, and returns its path as spelled on disk, or "" if
// there's no such file or folder.
func Lookup(root, rel string) string {
	dir := root
	for _, part := range strings.Split(rel, "/") {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return ""
		}
		found := ""
		for _, e := range entries {
			if strings.EqualFold(e.Name(), part) {
				found = e.Name()
				break
			}
		}
		if found == "" {
			return ""
		}
		dir = filepath.Join(dir, found)
	}
	if rel, err := filepath.Rel(root, dir); err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return ""
	}
	return dir
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
