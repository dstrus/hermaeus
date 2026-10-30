package records

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"hermaeus/internal/layers"
	"hermaeus/internal/store"
)

func write(t *testing.T, path, content string) {
	t.Helper()
	os.MkdirAll(filepath.Dir(path), 0o755)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestActionsAndTriggersAcrossLayers(t *testing.T) {
	sky := t.TempDir()
	mod := filepath.Join(t.TempDir(), "SkyrimNet")
	base := filepath.Join(mod, "SKSE", "Plugins", "SkyrimNet", "library", layers.BasePlugin)
	write(t, filepath.Join(base, "manifest.json"), `{"version": "0.28"}`)
	os.MkdirAll(filepath.Join(base, "prompts"), 0o755)
	write(t, filepath.Join(sky, "content-registry.json"), `{"plugins": {"a.pack": {"enabled": true}}, "priority": ["a.pack"]}`)
	pack := filepath.Join(sky, "library", "a.pack")
	write(t, filepath.Join(pack, "manifest.json"), `{"title": "Pack"}`)
	write(t, filepath.Join(pack, "actions", "sub", "wave.yaml"), `name: Wave
description: "Wave at {{ player.name }}."
customCategory: greetings
enabled: true
defaultPriority: 5
questEditorId: Q
scriptName: S
executionFunctionName: Wave
parameterMapping:
  - type: speaker
    name: akActor
    description: the one waving
  - type: dynamic
    name: how
    description: how, in a word
    optional: true
eligibilityRules:
  - conditions: []
  - conditions: []
`)
	write(t, filepath.Join(pack, "actions", "bow.yaml"), "name: Bow\ndescription: Bow.\nenabled: false\neventString: bow_event\n")
	write(t, filepath.Join(pack, "triggers", "sleep.yaml"), `name: Sleep
description: On sleeping.
eventCriteria:
  eventType: sleep_start
  schemaConditions:
    - fieldPath: who
      operator: equals
      value: player
audience: player
response:
  type: diary_entry
  content: "Dear diary"
  targetScope: player
enabled: true
cooldownSeconds: 120
probability: 0.5
interrupt: true
`)
	// Your own copy of the trigger replaces the pack's.
	write(t, filepath.Join(sky, "overlay", "triggers", "Sleep.yaml"), "name: Sleep\ndescription: Mine.\neventCriteria:\n  eventType: sleep_start\nenabled: false\n")
	l := layers.New(func() string { return sky })
	if _, err := l.SetModDir(mod); err != nil {
		t.Fatal(err)
	}
	r := New(store.New("", func() bool { return false }), l)

	acts := r.Actions().Actions
	if len(acts) != 2 || acts[0].Name != "Bow" || acts[1].Name != "Wave" {
		t.Fatalf("actions = %+v", acts)
	}
	bow, wave := acts[0], acts[1]
	if bow.Enabled || bow.Event != "bow_event" || bow.Source != "Pack" {
		t.Errorf("bow = %+v", bow)
	}
	if !wave.Enabled || wave.Category != "greetings" || *wave.Priority != 5 || wave.Rules != 2 || wave.Function != "Wave" || len(wave.Params) != 2 || !wave.Params[1].Optional || wave.Path != "actions/sub/wave.yaml" {
		t.Errorf("wave = %+v", wave)
	}

	trs := r.Triggers().Triggers
	if len(trs) != 1 {
		t.Fatalf("triggers = %+v", trs)
	}
	if s := trs[0]; s.LayerKind != "overlay" || s.Replaces != "Pack" || s.Description != "Mine." || s.Enabled {
		t.Errorf("your copy = %+v", s)
	}
	os.Remove(filepath.Join(sky, "overlay", "triggers", "Sleep.yaml"))
	s := r.Triggers().Triggers[0]
	if s.EventType != "sleep_start" || len(s.Conditions) != 1 || s.Conditions[0].Value != "player" || s.ResponseType != "diary_entry" || s.Content != "Dear diary" || *s.Cooldown != 120 || *s.Probability != 0.5 || !s.Interrupt || s.Replaces != "" {
		t.Errorf("pack's = %+v", s)
	}
	if !strings.HasPrefix(s.Path, "triggers/") {
		t.Errorf("path = %s", s.Path)
	}
}
