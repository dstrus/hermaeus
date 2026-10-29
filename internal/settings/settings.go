// Package settings reads and changes SkyrimNet's settings files, the YAML
// files in its config folder, and compares them with the defaults SkyrimNet
// records in defaults_manifest.json. API keys and other secrets are never
// sent to the page, only whether they're filled in.
package settings

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"unicode"

	"hermaeus/internal/apperr"
	"hermaeus/internal/install"
	"hermaeus/internal/prefs"
	"hermaeus/internal/store"

	"go.yaml.in/yaml/v3"
)

type Settings struct {
	store *store.Store
	prefs *prefs.Prefs // where pinned files are remembered
	lists modelLists   // providers' model lists, fetched this session
}

func New(s *store.Store, p *prefs.Prefs) *Settings {
	return &Settings{store: s, prefs: p}
}

// File is one settings file, named as SkyrimNet's defaults name it: the file
// name without .yaml, "game" for SkyrimNet.yaml, or Plugin_<name> for a
// plugin's settings.yaml.
type File struct {
	Name        string `json:"name"`
	Path        string `json:"path"` // under the SkyrimNet folder
	Exists      bool   `json:"exists"`
	Modified    int64  `json:"modified"` // the file's time, to spot changes made since
	HasDefaults bool   `json:"hasDefaults"`
	Settings    int    `json:"settings"`
	Changed     int    `json:"changed"` // settings that differ from their defaults
	Pinned      bool   `json:"pinned"`  // kept at the top of the Settings page
}

// Setting is one value in a file, named by its dotted key.
type Setting struct {
	Key        string `json:"key"`
	Value      any    `json:"value,omitempty"`
	Default    any    `json:"default,omitempty"`
	Set        bool   `json:"set"` // the file has it, rather than only a default
	HasDefault bool   `json:"hasDefault"`
	Changed    bool   `json:"changed"`
	Secret     bool   `json:"secret"`     // its value is only ever shown as SecretValue
	HasSecrets bool   `json:"hasSecrets"` // a list or map with a secret inside, so it can't be edited here
}

// SecretValue stands in for a secret wherever one would be sent.
type SecretValue struct {
	Secret bool `json:"secret"`
	Filled bool `json:"filled"`
}

// FileValues is a file with its settings: those in the file, in order, then
// any with a default the file doesn't set.
type FileValues struct {
	File   File      `json:"file"`
	Values []Setting `json:"values"`
}

type FileList struct {
	Files     []FileValues `json:"files"`
	Elsewhere []OwnPage    `json:"elsewhere"` // files the list leaves out, each for a page of its own
}

// OwnPage is a settings file that Hermaeus shows on a page of its own, in
// the feature's own terms, so the Settings page doesn't list it. The file
// can still be read and saved by name.
type OwnPage struct {
	Name  string `json:"name"`  // the file's name, as in File
	Label string `json:"label"` // the page's name
	Page  string `json:"page"`  // the page's path under the playthrough
}

var ownPages = []OwnPage{
	{Name: "UniversalTranslator", Label: "Universal Translator", Page: "translator"},
	{Name: "OpenRouter", Label: "Models", Page: "models"},
}

func hasOwnPage(name string) bool {
	for _, p := range ownPages {
		if strings.EqualFold(p.Name, name) {
			return true
		}
	}
	return false
}

func (s *Settings) configDir() (string, error) {
	dir := s.store.Dir()
	if dir == "" {
		return "", apperr.New(apperr.Unavailable, "No SkyrimNet folder chosen yet")
	}
	return filepath.Join(dir, "config"), nil
}

// A plugin's settings are named Plugin_<name>, as the defaults name them.
const pluginPrefix = "Plugin_"

// The defaults call some files by another name.
var fileFor = map[string]string{"game": "SkyrimNet"}

func nameFor(file string) string {
	for name, f := range fileFor {
		if strings.EqualFold(f, file) {
			return name
		}
	}
	return file
}

func filePath(configDir, name string) string {
	if p, ok := strings.CutPrefix(name, pluginPrefix); ok {
		return filepath.Join(configDir, "plugins", p, "settings.yaml")
	}
	if f, ok := fileFor[strings.ToLower(name)]; ok {
		name = f
	}
	return filepath.Join(configDir, name+".yaml")
}

var nameRe = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)

func validName(name string) bool {
	return nameRe.MatchString(name) && !strings.Contains(name, "..")
}

// defaults are SkyrimNet's recorded defaults, by file name then dotted key.
type defaults map[string]map[string]any

func loadDefaults(configDir string) defaults {
	var m struct {
		Fields defaults `json:"fields"`
	}
	if b, err := os.ReadFile(filepath.Join(configDir, "defaults_manifest.json")); err == nil {
		json.Unmarshal(b, &m)
	}
	if m.Fields == nil {
		m.Fields = defaults{}
	}
	return m.Fields
}

// of finds a file's defaults regardless of the name's case.
func (d defaults) of(name string) (map[string]any, bool) {
	if v, ok := d[name]; ok {
		return v, true
	}
	for n, v := range d {
		if strings.EqualFold(n, name) {
			return v, true
		}
	}
	return nil, false
}

// names lists every settings file there is or has defaults. Names match
// regardless of case, as on Windows (the defaults call MCP.yaml "mcp"), and
// a file's own spelling wins.
func names(configDir string, defs defaults) []string {
	found := map[string]string{}
	for n := range defs {
		found[strings.ToLower(n)] = n
	}
	for _, f := range install.FilesWithSuffix(configDir, ".yaml") {
		n := nameFor(strings.TrimSuffix(f, filepath.Ext(f)))
		found[strings.ToLower(n)] = n
	}
	plugins, _ := os.ReadDir(filepath.Join(configDir, "plugins"))
	for _, p := range plugins {
		if p.IsDir() && install.IsFile(filepath.Join(configDir, "plugins", p.Name(), "settings.yaml")) {
			found[strings.ToLower(pluginPrefix+p.Name())] = pluginPrefix + p.Name()
		}
	}
	out := make([]string, 0, len(found))
	for _, n := range found {
		if validName(n) {
			out = append(out, n)
		}
	}
	sort.Slice(out, func(i, j int) bool { return strings.ToLower(out[i]) < strings.ToLower(out[j]) })
	return out
}

// List reads every settings file, except those with a page of their own.
func (s *Settings) List() (FileList, error) {
	dir, err := s.configDir()
	if err != nil {
		return FileList{}, err
	}
	defs := loadDefaults(dir)
	out := FileList{Files: []FileValues{}, Elsewhere: ownPages}
	for _, name := range names(dir, defs) {
		if hasOwnPage(name) {
			continue
		}
		f, err := read(dir, name, defs)
		if err != nil {
			return FileList{}, err
		}
		f.File.Pinned = s.pinned(name)
		out.Files = append(out.Files, f)
	}
	// Pinned files first, in the order they were pinned; the rest by name.
	order := s.pinOrder()
	sort.SliceStable(out.Files, func(i, j int) bool {
		return order(out.Files[i].File.Name) < order(out.Files[j].File.Name)
	})
	return out, nil
}

// Read reads one settings file.
func (s *Settings) Read(name string) (FileValues, error) {
	dir, err := s.configDir()
	if err != nil {
		return FileValues{}, err
	}
	if !validName(name) {
		return FileValues{}, apperr.New(apperr.Invalid, "Invalid settings name %q", name)
	}
	f, err := read(dir, name, loadDefaults(dir))
	f.File.Pinned = s.pinned(name)
	return f, err
}

func read(configDir, name string, defs defaults) (FileValues, error) {
	path := filePath(configDir, name)
	rel, _ := filepath.Rel(filepath.Dir(configDir), path)
	f := FileValues{File: File{Name: name, Path: filepath.ToSlash(rel)}, Values: []Setting{}}
	fileDefs, hasDefs := defs.of(name)
	f.File.HasDefaults = hasDefs

	var doc *yaml.Node
	if b, err := os.ReadFile(path); err == nil {
		f.File.Exists = true
		if info, err := os.Stat(path); err == nil {
			f.File.Modified = info.ModTime().UnixMilli()
		}
		d, err := parse(b)
		if err != nil {
			return f, apperr.New(apperr.Invalid, "Could not read %s: %v", f.File.Path, err)
		}
		doc = d.doc
	} else if !os.IsNotExist(err) {
		return f, err
	}

	// Each default names either a leaf or a whole subtree (a default list or
	// map); the file's own leaves fill in the rest.
	covered := map[string]bool{}
	add := func(key string, node *yaml.Node) {
		st := Setting{Key: key, Secret: isSecret(name, key)}
		var value any
		if node != nil {
			st.Set = true
			node.Decode(&value)
		}
		def, hasDef := fileDefs[key]
		st.HasDefault = hasDef
		st.Changed = hasDef && st.Set && !sameValue(value, def)
		// SkyrimNet records defaults at single precision (0.699999988 for
		// 0.7); the page gets the number the player would write.
		def = cleanDefault(def)
		mask := func(v any) any {
			if st.Secret {
				return SecretValue{Secret: true, Filled: v != nil && fmt.Sprint(v) != ""}
			}
			return maskSecrets(key, v)
		}
		if st.Set {
			st.Value = mask(value)
		}
		st.HasSecrets = !st.Secret && (containsSecret(value) || containsSecret(def))
		if hasDef {
			st.Default = mask(def)
		}
		f.Values = append(f.Values, st)
		covered[key] = true
	}
	var walk func(n *yaml.Node, prefix string)
	walk = func(n *yaml.Node, prefix string) {
		for i := 0; i+1 < len(n.Content); i += 2 {
			key := n.Content[i].Value
			if prefix != "" {
				key = prefix + "." + key
			}
			v := n.Content[i+1]
			if _, isDefault := fileDefs[key]; !isDefault && v.Kind == yaml.MappingNode && len(v.Content) > 0 {
				walk(v, key)
				continue
			}
			add(key, v)
		}
	}
	if doc != nil && doc.Kind == yaml.MappingNode {
		walk(doc, "")
	}
	var unset []string
	for key := range fileDefs {
		if !covered[key] {
			unset = append(unset, key)
		}
	}
	sort.Strings(unset)
	for _, key := range unset {
		add(key, findNode(doc, key))
	}
	f.File.Settings = len(f.Values)
	for _, st := range f.Values {
		if st.Changed {
			f.File.Changed++
		}
	}
	return f, nil
}

// sameValue compares a setting with its default. SkyrimNet records float
// defaults at single precision (0.699999988 for 0.7), so numbers that agree
// as float32 are the same.
func sameValue(a, b any) bool {
	if x, ok := toNumber(a); ok {
		y, ok := toNumber(b)
		return ok && (x == y || float32(x) == float32(y))
	}
	switch x := a.(type) {
	case []any:
		y, ok := b.([]any)
		if !ok || len(x) != len(y) {
			return false
		}
		for i := range x {
			if !sameValue(x[i], y[i]) {
				return false
			}
		}
		return true
	case map[string]any:
		y, ok := b.(map[string]any)
		if !ok || len(x) != len(y) {
			return false
		}
		for k, v := range x {
			w, ok := y[k]
			if !ok || !sameValue(v, w) {
				return false
			}
		}
		return true
	}
	return a == b
}

func toNumber(v any) (float64, bool) {
	switch x := v.(type) {
	case int:
		return float64(x), true
	case int64:
		return float64(x), true
	case uint64:
		return float64(x), true
	case float64:
		return x, !math.IsNaN(x)
	}
	return 0, false
}

// ---------- secrets ----------

// isSecretPath reports whether a setting's last name part is a key, token,
// secret, or password, such as api_key, api_key_base64, access_token, or
// apiKey, but not max_tokens, keyword, or toggleMuteHotkey, or an identifier
// tied to the player's account with a service, such as Inworld's workspace_id.
func isSecretPath(path string) bool {
	part := path[strings.LastIndex(path, ".")+1:]
	var b strings.Builder
	for i, r := range part { // camelCase to snake_case
		if unicode.IsUpper(r) && i > 0 {
			b.WriteByte('_')
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return secretNameRe.MatchString(b.String())
}

var secretNameRe = regexp.MustCompile(`(^|_)(key|apikey|token|secret|password|passwd)(_base64)?$|^(workspace|account|organization|org)_id$`)

// Settings SkyrimNet's dashboard hides although their names don't say
// they're secret: a private Inworld endpoint can identify the account.
var dashboardSecrets = map[string]bool{"Inworld.endpoint": true}

func isSecret(file, key string) bool {
	return isSecretPath(key) || dashboardSecrets[file+"."+key]
}

// maskSecrets replaces secrets anywhere in a value, including inside lists
// and maps (Providers.yaml keeps an api_key in each provider).
func maskSecrets(key string, v any) any {
	if isSecretPath(key) {
		return SecretValue{Secret: true, Filled: v != nil && fmt.Sprint(v) != ""}
	}
	switch x := v.(type) {
	case []any:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = maskSecrets("", e)
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, e := range x {
			out[k] = maskSecrets(k, e)
		}
		return out
	}
	return v
}

// containsSecret reports whether a list or map holds a secret somewhere inside.
func containsSecret(v any) bool {
	switch x := v.(type) {
	case []any:
		for _, e := range x {
			if containsSecret(e) {
				return true
			}
		}
	case map[string]any:
		for k, e := range x {
			if isSecretPath(k) || containsSecret(e) {
				return true
			}
		}
	}
	return false
}
