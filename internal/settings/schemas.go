package settings

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"go.yaml.in/yaml/v3"
)

// A SkyrimNet plugin can describe its own settings in a schema:
// config/plugins/<Name>/manifest.yaml, in its mod or plugin folder, lists
// each setting with a label, a description, its type, a default, and a
// category. The values live in config/plugins/<Name>/settings.yaml, which
// is the settings file Plugin_<Name>. Hermaeus shows each such file as a
// page of its own, built from the schema.

// PluginSchema is one plugin's settings, as its schema describes them.
type PluginSchema struct {
	Name        string        `json:"name"`   // the settings file: Plugin_<Name>
	Folder      string        `json:"folder"` // <Name>
	Title       string        `json:"title"`
	Description string        `json:"description,omitempty"`
	Version     string        `json:"version,omitempty"`
	Plugin      string        `json:"plugin,omitempty"` // the installed plugin shipped alongside it, when known
	Fields      []SchemaField `json:"fields"`
	Variants    []string      `json:"variants"` // model tasks it adds, as named on the Models page
}

// SchemaField is one setting: type is bool, int, float, string, select, or
// hotkey (a virtual-key code).
type SchemaField struct {
	Name        string   `json:"name"`
	Path        string   `json:"path"`
	Type        string   `json:"type"`
	Description string   `json:"description,omitempty"`
	Default     any      `json:"default,omitempty"`
	Category    string   `json:"category,omitempty"`
	Min         *float64 `json:"min,omitempty"`
	Max         *float64 `json:"max,omitempty"`
	Options     []string `json:"options,omitempty"`
	Secret      bool     `json:"secret,omitempty"` // written but never read back, like an API key
}

type schemaFile struct {
	Variants []struct {
		Name string `yaml:"name"`
	} `yaml:"variants"`
	Plugin struct {
		Name        string `yaml:"name"`
		Version     string `yaml:"version"`
		Description string `yaml:"description"`
	} `yaml:"plugin"`
	Schema struct {
		Fields []struct {
			Name         string   `yaml:"name"`
			Path         string   `yaml:"path"`
			Type         string   `yaml:"type"`
			Description  string   `yaml:"description"`
			DefaultValue any      `yaml:"defaultValue"`
			Category     string   `yaml:"category"`
			Min          *float64 `yaml:"min"`
			Max          *float64 `yaml:"max"`
			Options      []string `yaml:"options"`
		} `yaml:"fields"`
	} `yaml:"schema"`
}

// WithRoots tells Settings where SkyrimNet's plugins are: the SkyrimNet
// folder and, under MO2, each mod's SkyrimNet folder.
func (s *Settings) WithRoots(roots func() []string) *Settings {
	s.roots = roots
	return s
}

// PluginSchemas finds every plugin's settings schema, by title.
func (s *Settings) PluginSchemas() []PluginSchema {
	out := []PluginSchema{}
	var roots []string
	if s.roots != nil {
		roots = s.roots()
	} else if dir := s.store.Dir(); dir != "" {
		roots = []string{dir}
	}
	seen := map[string]bool{}
	for _, root := range roots {
		dirs, _ := os.ReadDir(filepath.Join(root, "config", "plugins"))
		for _, d := range dirs {
			key := strings.ToLower(d.Name())
			if !d.IsDir() || seen[key] {
				continue
			}
			b, err := os.ReadFile(filepath.Join(root, "config", "plugins", d.Name(), "manifest.yaml"))
			if err != nil {
				continue
			}
			var f schemaFile
			if yaml.Unmarshal(b, &f) != nil || len(f.Schema.Fields) == 0 {
				continue
			}
			seen[key] = true
			ps := PluginSchema{Name: pluginPrefix + d.Name(), Folder: d.Name(), Title: f.Plugin.Name, Description: f.Plugin.Description, Version: f.Plugin.Version, Fields: []SchemaField{}, Variants: []string{}}
			for _, v := range f.Variants {
				if v.Name != "" {
					ps.Variants = append(ps.Variants, v.Name)
				}
			}
			if ps.Title == "" {
				ps.Title = d.Name()
			}
			// A mod's own plugin sits beside its schema.
			if ext, _ := os.ReadDir(filepath.Join(root, "external")); len(ext) == 1 && ext[0].IsDir() {
				ps.Plugin = ext[0].Name()
			}
			for _, fd := range f.Schema.Fields {
				if fd.Path == "" {
					continue
				}
				ps.Fields = append(ps.Fields, SchemaField{
					Name: fd.Name, Path: fd.Path, Type: strings.ToLower(fd.Type), Description: fd.Description,
					Default: fd.DefaultValue, Category: fd.Category, Min: fd.Min, Max: fd.Max, Options: fd.Options,
					Secret: isSecretPath(fd.Path),
				})
			}
			out = append(out, ps)
		}
	}
	sort.Slice(out, func(i, j int) bool { return strings.ToLower(out[i].Title) < strings.ToLower(out[j].Title) })
	return out
}

// schemaFor is a plugin settings file's schema, if it has one.
func (s *Settings) schemaFor(name string, schemas []PluginSchema) *PluginSchema {
	for i := range schemas {
		if strings.EqualFold(schemas[i].Name, name) {
			return &schemas[i]
		}
	}
	return nil
}

// loadDefaults is SkyrimNet's recorded defaults, with a plugin schema's
// defaults for any setting SkyrimNet hasn't recorded.
func (s *Settings) loadDefaults(configDir string) defaults {
	defs := loadDefaults(configDir)
	for _, ps := range s.PluginSchemas() {
		var have map[string]any
		for n, v := range defs {
			if strings.EqualFold(n, ps.Name) {
				have = v
			}
		}
		if have == nil {
			have = map[string]any{}
			defs[ps.Name] = have
		}
		for _, f := range ps.Fields {
			if _, ok := have[f.Path]; !ok && f.Default != nil {
				have[f.Path] = f.Default
			}
		}
	}
	return defs
}

// markPluginPage notes a plugin settings file's page, from its schema.
func (s *Settings) markPluginPage(f *File, schemas []PluginSchema) {
	if ps := s.schemaFor(f.Name, schemas); ps != nil {
		f.Page, f.PageLabel = "settings/plugin/"+ps.Folder, ps.Title
	}
}
