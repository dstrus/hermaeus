package settings

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPluginSchemas(t *testing.T) {
	h := newHarness(t)
	mod := filepath.Join(t.TempDir(), "SeverActions 4.5.0", "SKSE", "Plugins", "SkyrimNet")
	os.MkdirAll(filepath.Join(mod, "external", "severause.severactions"), 0o755)
	os.MkdirAll(filepath.Join(mod, "config", "plugins", "SeverActions"), 0o755)
	os.WriteFile(filepath.Join(mod, "config", "plugins", "SeverActions", "manifest.yaml"), []byte(`variants:
  - name: sever_background
    description: "Background prompts"
plugin:
  name: SeverActions
  version: "3.0"
  description: "An action pack"
schema:
  fields:
    - name: "Model Name"
      path: "llm.model_name"
      type: "string"
      defaultValue: ""
      category: "LLM Overrides"
    - name: "API Key"
      path: "llm.api_key"
      type: "string"
      category: "LLM Overrides"
    - name: "Temperature"
      path: "llm.temperature"
      type: "float"
      min: 0.0
      max: 2.0
      defaultValue: 0.7
      category: "LLM Overrides"
    - name: "Mood"
      path: "mood"
      type: "select"
      options: [calm, wild]
      defaultValue: calm
`), 0o644)
	h.config(filepath.Join("plugins", "SeverActions", "settings.yaml"), "llm:\n  model_name: x\n  api_key: sk-SECRET\nkeys:\n  mute: \"\"\n")
	h.config("defaults_manifest.json", `{"fields": {"Plugin_SeverActions": {"keys.mute": null}}}`)
	h.settings.WithRoots(func() []string { return []string{h.Dir, mod} })

	schemas := h.settings.PluginSchemas()
	if len(schemas) != 1 {
		t.Fatalf("schemas = %+v", schemas)
	}
	ps := schemas[0]
	if ps.Name != "Plugin_SeverActions" || ps.Title != "SeverActions" || ps.Plugin != "severause.severactions" || len(ps.Fields) != 4 || strings.Join(ps.Variants, ",") != "sever_background" {
		t.Errorf("schema = %+v", ps)
	}
	if f := ps.Fields[1]; !f.Secret || f.Category != "LLM Overrides" {
		t.Errorf("api key = %+v", f)
	}
	if f := ps.Fields[2]; f.Type != "float" || *f.Min != 0 || *f.Max != 2 || f.Default != 0.7 {
		t.Errorf("temperature = %+v", f)
	}
	if f := ps.Fields[3]; strings.Join(f.Options, ",") != "calm,wild" {
		t.Errorf("mood = %+v", f)
	}

	// The schema's defaults fill in, the file gets its page, the key stays masked.
	f, err := h.settings.Read("Plugin_SeverActions")
	if err != nil {
		t.Fatal(err)
	}
	if f.File.Page != "plugin-settings/SeverActions" || f.File.PageLabel != "SeverActions" {
		t.Errorf("file = %+v", f.File)
	}
	got := map[string]Setting{}
	for _, v := range f.Values {
		got[v.Key] = v
	}
	if s := got["llm.temperature"]; !s.HasDefault || s.Default != 0.7 || s.Set {
		t.Errorf("temperature = %+v", s)
	}
	if s := got["mood"]; s.Default != "calm" {
		t.Errorf("mood = %+v", s)
	}
	if s := got["keys.mute"]; s.Changed {
		t.Errorf("an empty value against a null default counts as changed: %+v", s)
	}
	if sent, _ := json.Marshal(f); strings.Contains(string(sent), "SECRET") {
		t.Error("the key was sent to the page")
	}
	// Saving a setting that's only a default adds it to the file.
	if _, err := h.settings.Save("Plugin_SeverActions", Edit{Modified: f.File.Modified, Changes: []Change{{Key: "llm.temperature", Value: 0.9}}}); err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(filepath.Join(h.Dir, "config", "plugins", "SeverActions", "settings.yaml")); !strings.Contains(string(b), "temperature: 0.9") {
		t.Errorf("after saving:\n%s", b)
	}
}
