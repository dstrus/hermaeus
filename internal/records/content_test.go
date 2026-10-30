package records

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"hermaeus/internal/apperr"
	"hermaeus/internal/layers"
	"hermaeus/internal/store"
	"hermaeus/internal/testenv"
)

type harness struct {
	*testenv.Folder
	running bool
	base    string
	r       *Records
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	h := &harness{Folder: testenv.NewFolder(t)}
	s := store.New("", func() bool { return h.running })
	if _, err := s.SetDir(h.Dir); err != nil {
		t.Fatal(err)
	}
	mod := filepath.Join(t.TempDir(), "SkyrimNet")
	h.base = filepath.Join(mod, "SKSE", "Plugins", "SkyrimNet", "library", layers.BasePlugin)
	write(t, filepath.Join(h.base, "manifest.json"), `{"version": "0.28"}`)
	os.MkdirAll(filepath.Join(h.base, "prompts"), 0o755)
	l := layers.New(s.Dir)
	if _, err := l.SetModDir(mod); err != nil {
		t.Fatal(err)
	}
	h.r = New(s, l)
	return h
}

func (h *harness) lib(id string, parts ...string) string {
	return filepath.Join(append([]string{h.Dir, "library", id}, parts...)...)
}

func (h *harness) mine(rel string) string {
	b, _ := os.ReadFile(filepath.Join(h.Dir, "overlay", filepath.FromSlash(rel)))
	return string(b)
}

func kindOf(t *testing.T, err error, want apperr.Kind) {
	t.Helper()
	if kind, ok := apperr.KindOf(err); !ok || kind != want {
		t.Errorf("error = %v, want kind %v", err, want)
	}
}

func TestOverridesListWhatReplacesWhat(t *testing.T) {
	h := newHarness(t)
	old, newer := time.Now().Add(-48*time.Hour), time.Now()
	write(t, filepath.Join(h.base, "prompts", "dialogue_response.prompt"), "shipped\r\nline\r\n")
	os.Chtimes(filepath.Join(h.base, "prompts", "dialogue_response.prompt"), newer, newer)
	write(t, filepath.Join(h.Dir, "overlay", "Prompts", "Dialogue_Response.prompt"), "mine\nline\n") // another case, as on Windows
	os.Chtimes(filepath.Join(h.Dir, "overlay", "Prompts", "Dialogue_Response.prompt"), old, old)
	write(t, filepath.Join(h.Dir, "overlay", "triggers", "mine.yaml"), "added: true\n")
	write(t, filepath.Join(h.Dir, "overlay", "_history", "prompts", "x.prompt.backup.1"), "history")
	write(t, filepath.Join(h.base, "prompts", "tags.prompt"), "shipped tags")
	write(t, h.lib("bios.pack", "prompts", "tags.prompt"), "pack tags")
	write(t, h.lib("bios.pack", "manifest.json"), `{"title": "Lydia Pack"}`)
	write(t, h.lib("bios.two", "prompts", "tags.prompt"), "other pack tags")
	write(t, h.lib("bios.two", "manifest.json"), `{"title": "Second Pack"}`)
	write(t, h.lib("bios.two", "install.json"), `{}`)
	write(t, h.lib("bios.two", "knowledge", "new.sknpack"), "\x00binary")
	write(t, filepath.Join(h.Dir, "content-registry.json"), `{"plugins": {"bios.pack": {}, "bios.two": {}, "bios.off": {"enabled": false}}, "priority": ["bios.off", "bios.pack", "bios.two"]}`)

	list := h.r.Overrides()
	got := map[string]Override{}
	for _, o := range list.Overrides {
		got[strings.ToLower(o.Path)] = o
	}
	sources := func(o Override) string {
		var s []string
		for _, p := range o.Providers {
			s = append(s, list.Layers[p.Layer].Source)
		}
		return strings.Join(s, " > ")
	}
	if o, ok := got["prompts/dialogue_response.prompt"]; !ok || sources(o) != "your dashboard edits > SkyrimNet’s defaults" || !o.Stale {
		t.Errorf("dialogue_response = %+v", o)
	}
	if o, ok := got["prompts/tags.prompt"]; !ok || sources(o) != "Lydia Pack > Second Pack > SkyrimNet’s defaults" || !o.Conflict || o.Stale {
		t.Errorf("tags = %+v (%s)", o, sources(o))
	}
	if _, ok := got["triggers/mine.yaml"]; !ok {
		t.Error("a file added in your edits isn't listed")
	}
	for _, p := range []string{"knowledge/new.sknpack", "manifest.json", "install.json", "_history/prompts/x.prompt.backup.1"} {
		if _, ok := got[p]; ok {
			t.Errorf("%s shouldn't be listed", p)
		}
	}

	v, err := h.r.Versions("prompts/dialogue_response.prompt")
	if err != nil || len(v) != 2 {
		t.Fatalf("versions = %+v, %v", v, err)
	}
	if *v[0].Text != "mine\nline\n" || v[0].CRLF || !strings.HasSuffix(v[0].Path, "/Dialogue_Response.prompt") {
		t.Errorf("your version = %+v", v[0])
	}
	if *v[1].Text != "shipped\nline\n" || !v[1].CRLF {
		t.Errorf("shipped version = %+v", v[1])
	}
	if v, _ := h.r.Versions("knowledge/new.sknpack"); len(v) != 1 || v[0].Text != nil {
		t.Errorf("a binary file has text: %+v", v)
	}
	for _, bad := range []string{"../x", "/etc/passwd", "prompts/../../x", "C:/x", ""} {
		_, err := h.r.Versions(bad)
		kindOf(t, err, apperr.Invalid)
	}
	_, err = h.r.Versions("prompts/nothing.prompt")
	kindOf(t, err, apperr.NotFound)
}

func TestRecordsSaveSwitchAndRevert(t *testing.T) {
	h := newHarness(t)
	write(t, filepath.Join(h.base, "voice_effects", "draugr.yaml"), "id: draugr\r\nname: Draugr\r\nstages:\r\n  - type: pitch_shift\r\n    params:\r\n      semitones: -6.0\r\n")
	write(t, filepath.Join(h.base, "filters", "animals.yaml"), "kind: memory\nname: Animals\nRaceBlacklist:\n  - DogRace\n")
	write(t, h.lib("a.pack", "manifest.json"), `{"title": "Pack"}`)
	src := `customCategory: brawl
name: AcceptBrawl
description: "Use when dared."
parameterMapping:
  - type: dynamic
    name: challenger
    # optional is load-bearing: without it the action aborts.
    optional: true
eligibilityRules:
  - conditions:
      # Custom decorator: STRING compare, "" = no cooldown.
      - decoratorName: "sever_combat_cooldown"
`
	write(t, h.lib("a.pack", "actions", "acceptbrawl.yaml"), src)
	write(t, h.lib("a.pack", "triggers", "wave.yaml"), "name: wave\neventCriteria:\n  eventType: sleep_start\nresponse:\n  type: player_thought\n  content: \"Time to rest.\"\nenabled: true\n")
	write(t, filepath.Join(h.Dir, "content-registry.json"), `{"plugins": {"a.pack": {}}, "priority": ["a.pack"]}`)

	c := h.r.Content()
	var paths []string
	for _, rec := range c.Records {
		paths = append(paths, rec.Path+"="+rec.Name)
	}
	if strings.Join(paths, " ") != "triggers/wave.yaml=wave actions/acceptbrawl.yaml=AcceptBrawl filters/animals.yaml=Animals voice_effects/draugr.yaml=Draugr" {
		t.Fatalf("records = %v", paths)
	}
	if a := c.Records[1]; !a.Switch || !a.Enabled {
		t.Errorf("an action without enabled = %+v", a)
	}
	if f := c.Records[2]; f.Switch {
		t.Errorf("a filter list has no switch: %+v", f)
	}

	// Switching an action off copies it into your edits, changing one line
	// and keeping its comments.
	if _, err := h.r.SetEnabled("actions/acceptbrawl.yaml", Switch{Enabled: false}); err != nil {
		t.Fatal(err)
	}
	if got := h.mine("actions/acceptbrawl.yaml"); got != src+"enabled: false\n" {
		t.Errorf("switched off =\n%s", got)
	}
	// The trigger's own enabled is changed in place, its quoting kept.
	if _, err := h.r.SetEnabled("triggers/wave.yaml", Switch{Enabled: false}); err != nil {
		t.Fatal(err)
	}
	if got := h.mine("triggers/wave.yaml"); !strings.Contains(got, "enabled: false\n") || !strings.Contains(got, `content: "Time to rest."`) {
		t.Errorf("trigger = %q", got)
	}

	// Saving your copy keeps the line endings of what it replaces.
	v, err := h.r.SaveRecord("voice_effects/draugr.yaml", RecordEdit{Text: "id: draugr\nname: Draugr lord\nstages:\n  - type: pitch_shift\n    params:\n      semitones: -6.0\n"})
	if err != nil {
		t.Fatal(err)
	}
	if got := h.mine("voice_effects/draugr.yaml"); got != "id: draugr\r\nname: Draugr lord\r\nstages:\r\n  - type: pitch_shift\r\n    params:\r\n      semitones: -6.0\r\n" || len(v) != 2 {
		t.Errorf("voice effect = %q (%d versions)", got, len(v))
	}
	// A second save needs the copy's time.
	_, err = h.r.SaveRecord("voice_effects/draugr.yaml", RecordEdit{Text: "id: draugr\n"})
	kindOf(t, err, apperr.Conflict)

	// Checks: the identity field matches the file, and kinds need their fields.
	for rel, text := range map[string]string{
		"voice_effects/new.yaml": "id: other\nstages: []\n",
		"triggers/new.yaml":      "name: new\neventCriteria:\n  eventType: death\n",
		"filters/new.yaml":       "kind: tts_rule\nid: new\n",
		"filters/x.yaml":         "kind: actor\nRaceBlacklist: DogRace\n",
		"actions/x.yaml":         "name: x\nparameterMapping:\n  - type: nope\n",
		"triggers/y.yaml":        "name: [unclosed\n",
		"../x.yaml":              "a: 1\n",
		"prompts/x.yaml":         "a: 1\n",
	} {
		_, err := h.r.SaveRecord(rel, RecordEdit{Text: text})
		kindOf(t, err, apperr.Invalid)
	}
	// A new record goes into your edits.
	if _, err := h.r.SaveRecord("triggers/new.yaml", RecordEdit{Text: "name: new\neventCriteria:\n  eventType: death\nresponse:\n  type: player_thought\n"}); err != nil {
		t.Fatal(err)
	}

	// Reverting removes your copy, bringing back the plugin's.
	v, _ = h.r.Versions("triggers/wave.yaml")
	v, err = h.r.RevertRecord("triggers/wave.yaml", Revert{Modified: v[0].Modified})
	if err != nil || len(v) != 1 || h.mine("triggers/wave.yaml") != "" {
		t.Errorf("after reverting: %+v, %v", v, err)
	}
	// Reverting a record only you have leaves nothing.
	v, _ = h.r.Versions("triggers/new.yaml")
	if v, err := h.r.RevertRecord("triggers/new.yaml", Revert{Modified: v[0].Modified}); err != nil || len(v) != 0 {
		t.Errorf("after reverting your own: %+v, %v", v, err)
	}
	h.running = true
	_, err = h.r.SetEnabled("actions/acceptbrawl.yaml", Switch{Enabled: true})
	kindOf(t, err, apperr.Conflict)
}

func TestVoiceEffects(t *testing.T) {
	h := newHarness(t)
	src := "id: ghost\nname: Ghost\n# Kept for the dashboard.\nsample: null\nstages:\n  - type: pitch_shift\n    params:\n      semitones: -1.0\n  - type: reverb\n    params:\n      room_size: 0.9\n      decay: 0.85\n      wet: 0.55\n"
	write(t, filepath.Join(h.base, "voice_effects", "ghost.yaml"), src)

	l := h.r.VoiceEffects()
	if len(l.Effects) != 1 || l.Effects[0].Name != "Ghost" || len(l.Effects[0].Stages) != 2 || !l.Effects[0].Enabled || l.Effects[0].Stages[1].Params["wet"] != 0.55 {
		t.Fatalf("effects = %+v", l.Effects)
	}
	if len(l.Kinds) != len(StageKinds) {
		t.Error("the kinds of stage aren't listed")
	}

	// Saving your copy rewrites the stages, keeping the rest of the file.
	l, err := h.r.SaveEffect("ghost", EffectEdit{Name: "Restless ghost", Stages: []Stage{
		{Type: "pitch_shift", Params: map[string]any{"semitones": -2.0}},
		{Type: "chorus", Params: map[string]any{"voices": 2.0, "rate_hz": 0.8, "depth_ms": 5.0}},
	}})
	if err != nil {
		t.Fatal(err)
	}
	got := h.mine("voice_effects/ghost.yaml")
	for _, want := range []string{"id: ghost", "name: Restless ghost", "# Kept for the dashboard.", "semitones: -2.0", "rate_hz: 0.8\n      depth_ms: 5.0\n      voices: 2\n"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in\n%s", want, got)
		}
	}
	if strings.Contains(got, "reverb") {
		t.Errorf("a removed stage is still there:\n%s", got)
	}

	// A new effect starts from its id.
	if _, err := h.r.SaveEffect("frost_atronach", EffectEdit{Name: "Frost atronach", Sample: "voices/frost.wav", Stages: []Stage{{Type: "gain", Params: map[string]any{"gain_db": 2.0}}}}); err != nil {
		t.Fatal(err)
	}
	if got := h.mine("voice_effects/frost_atronach.yaml"); !strings.HasPrefix(got, "id: frost_atronach\nname: Frost atronach\nsample: voices/frost.wav\nstages:\n") {
		t.Errorf("new effect =\n%s", got)
	}

	// The id, a name, and each stage's type are checked before anything is
	// written; the limits on a stage's settings aren't published, so they
	// aren't.
	for _, c := range []struct {
		id string
		e  EffectEdit
	}{
		{"Bad Id", EffectEdit{Name: "x"}},
		{"x", EffectEdit{Name: " "}},
		{"x", EffectEdit{Name: "x", Stages: []Stage{{Type: " "}}}},
	} {
		_, err := h.r.SaveEffect(c.id, c.e)
		kindOf(t, err, apperr.Invalid)
	}
	if _, err := os.Stat(filepath.Join(h.Dir, "overlay", "voice_effects", "x.yaml")); err == nil {
		t.Error("a refused effect was written")
	}
	// A stage of a type Hermaeus doesn't know is kept, for a newer SkyrimNet.
	if _, err := h.r.SaveEffect("future", EffectEdit{Name: "Future", Stages: []Stage{{Type: "flanger", Params: map[string]any{"rate_hz": 0.5}}}}); err != nil {
		t.Errorf("an unknown stage type: %v", err)
	}
	// The raw editor on Content applies the same check.
	_, err = h.r.SaveRecord("voice_effects/ghost.yaml", RecordEdit{Modified: l.Effects[0].Providers[0].Modified, Text: "id: ghost\nstages:\n  - params:\n      wet: 0.3\n"})
	kindOf(t, err, apperr.Invalid)
}
