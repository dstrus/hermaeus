package settings

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"hermaeus/internal/apperr"
	"hermaeus/internal/prefs"
	"hermaeus/internal/store"
	"hermaeus/internal/testenv"

	"go.yaml.in/yaml/v3"
)

type harness struct {
	*testenv.Folder
	t        *testing.T
	running  bool
	store    *store.Store
	settings *Settings
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	h := &harness{Folder: testenv.NewFolder(t), t: t}
	h.store = store.New("", func() bool { return h.running })
	if _, err := h.store.SetDir(h.Dir); err != nil {
		t.Fatal(err)
	}
	h.settings = New(h.store, prefs.Load(""))
	return h
}

// config writes a file under the SkyrimNet folder's config folder.
func (h *harness) config(name, content string) string {
	h.t.Helper()
	path := filepath.Join(h.Dir, "config", name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		h.t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		h.t.Fatal(err)
	}
	return path
}

func (h *harness) read(name string) string {
	h.t.Helper()
	b, err := os.ReadFile(filepath.Join(h.Dir, "config", name))
	if err != nil {
		h.t.Fatal(err)
	}
	return string(b)
}

func kindOf(t *testing.T, err error, want apperr.Kind) {
	t.Helper()
	if kind, ok := apperr.KindOf(err); !ok || kind != want {
		t.Errorf("error = %v, want kind %v", err, want)
	}
}

func TestComparedWithDefaults(t *testing.T) {
	h := newHarness(t)
	h.config("defaults_manifest.json", `{"fields": {
		"Memory": {"scoring.semantic": 0.44999998807907104, "scoring.actor": 0.2, "retrieval.max": 5, "filters": [], "unset.thing": true},
		"OpenRouter": {"openrouter.api_key": "", "openrouter.max_tokens": 4096},
		"Plugin_Sever": {"llm.model": "a"}}}`)
	h.config("Memory.yaml", "scoring:\n  semantic: 0.45\n  actor: 0.3\nretrieval:\n  max: 5\nfilters:\n  - animals\nextra: 1\n")
	h.config("OpenRouter.yaml", "openrouter:\n  api_key: sk-or-SECRET-ONE\n  max_tokens: 4096\n  apiKey: SECRET-TWO\n")
	h.config("Providers.yaml", "providers:\n  - id: openrouter\n    api_key: SECRET-THREE\n    api_key_from: openrouter\n")
	h.config("Auth.yaml", "access_token: SECRET-FOUR\n")
	h.config(filepath.Join("plugins", "Sever", "settings.yaml"), "llm:\n  model: b\n  toggleMuteHotkey: F5\n")
	h.config("UniversalTranslator.yaml", "enabled: true\n") // shown on its own page, so not listed

	list, err := h.settings.List()
	if err != nil {
		t.Fatal(err)
	}
	// A file with a page of its own says so, as listed and as read by name.
	for _, f := range list.Files {
		if f.File.Name == "UniversalTranslator" && (f.File.Page != "behavior/translator" || f.File.PageLabel != "Universal Translator") {
			t.Errorf("the translator's page = %+v", f.File)
		}
		if f.File.Name == "Memory" && f.File.Page != "memory/memories" {
			t.Errorf("Memory's page = %+v", f.File)
		}
	}
	if f, err := h.settings.Read("UniversalTranslator"); err != nil || !f.File.Exists || f.File.Page != "behavior/translator" {
		t.Errorf("read by name: %+v, %v", f.File, err)
	}
	// OpenRouter.yaml and Providers.yaml have pages of their own, so their
	// secrets are checked as read by name.
	openRouter, err := h.settings.Read("OpenRouter")
	if err != nil {
		t.Fatal(err)
	}
	providersFile, err := h.settings.Read("Providers")
	if err != nil {
		t.Fatal(err)
	}
	sent, _ := json.Marshal([]any{list, openRouter, providersFile})
	for _, secret := range []string{"SECRET-ONE", "SECRET-TWO", "SECRET-THREE", "SECRET-FOUR"} {
		if strings.Contains(string(sent), secret) {
			t.Errorf("%s would be sent to the page", secret)
		}
	}
	files := map[string]map[string]Setting{}
	for _, f := range append(list.Files, openRouter, providersFile) {
		files[f.File.Name] = map[string]Setting{}
		for _, v := range f.Values {
			files[f.File.Name][v.Key] = v
		}
	}
	check := func(file, key string, changed, set bool) {
		t.Helper()
		v, ok := files[file][key]
		if !ok {
			t.Errorf("%s: %s missing", file, key)
			return
		}
		if v.Changed != changed || v.Set != set {
			t.Errorf("%s: %s changed=%v set=%v, want %v %v", file, key, v.Changed, v.Set, changed, set)
		}
	}
	check("Memory", "scoring.semantic", false, true) // 0.45 vs 0.449999988: rounding only
	check("Memory", "scoring.actor", true, true)
	check("Memory", "retrieval.max", false, true)
	check("Memory", "filters", true, true) // a list default, compared whole
	check("Memory", "unset.thing", false, false)
	check("Memory", "extra", false, true) // no default to compare with
	check("Plugin_Sever", "llm.model", true, true)
	if files["Memory"]["extra"].HasDefault {
		t.Error("extra has a default")
	}
	if d := files["Memory"]["scoring.semantic"].Default; d != 0.45 {
		t.Errorf("scoring.semantic default sent as %v, want 0.45 (single-precision rounding undone)", d)
	}

	secret := func(file, key string) bool {
		v, _ := files[file][key].Value.(SecretValue)
		return v.Secret
	}
	for _, k := range [][2]string{{"OpenRouter", "openrouter.api_key"}, {"OpenRouter", "openrouter.apiKey"}, {"Auth", "access_token"}} {
		if !secret(k[0], k[1]) {
			t.Errorf("%s %s isn't masked", k[0], k[1])
		}
	}
	if secret("OpenRouter", "openrouter.max_tokens") || secret("Plugin_Sever", "llm.toggleMuteHotkey") {
		t.Error("an ordinary setting was masked")
	}
	providers := files["Providers"]["providers"].Value.([]any)[0].(map[string]any)
	if providers["api_key"] != (SecretValue{Secret: true, Filled: true}) || providers["api_key_from"] != "openrouter" {
		t.Errorf("provider = %v, want its api_key masked and api_key_from shown", providers)
	}
	if !files["Providers"]["providers"].HasSecrets {
		t.Error("providers doesn't say it holds secrets")
	}
}

func TestSecretNames(t *testing.T) {
	for name, want := range map[string]bool{
		"api_key": true, "openrouter.api_key": true, "api_key_base64": true, "apiKey": true, "apikey": true,
		"access_token": true, "clone_key": true, "client_secret": true, "password": true,
		"workspace_id": true, "workspaceId": true, "account_id": true, "org_id": true, "model_id": false, "provider_id": false,
		"max_tokens": false, "keyword": false, "scoring_weights.keyword": false, "toggleMuteHotkey": false,
		"api_key_from": false, "unconditional_keys": false, "banned_keywords": false,
	} {
		if got := isSecretPath(name); got != want {
			t.Errorf("isSecretPath(%q) = %v, want %v", name, got, want)
		}
	}
}

func TestSavingChangesOnlyTheValue(t *testing.T) {
	h := newHarness(t)
	h.config("defaults_manifest.json", `{"fields": {"game": {"narration.enabled": true, "narration.chance": 0.699999988079071, "general.maxTokens": 4096, "general.newThing": "x"}}}`)
	// As SkyrimNet writes it: Windows line endings, no final newline, plus a
	// comment and a quoted value, which a player's own edit might add.
	orig := "narration:\r\n  enabled: true\r\n  chance: 0.5  # keep low\r\ngeneral:\r\n  maxTokens: 4096\r\n  name: 'it''s'\r\nempty:\r\n  []"
	path := h.config("SkyrimNet.yaml", orig)
	// Written a while ago, so saving visibly changes the file's time.
	past := time.Now().Add(-time.Hour)
	os.Chtimes(path, past, past)
	modified := past.UnixMilli()

	f, err := h.settings.Save("game", Edit{Modified: modified, Changes: []Change{
		{Key: "narration.enabled", Value: false},
		{Key: "narration.chance", Reset: true},
		{Key: "general.maxTokens", Value: 2048.0}, // as JSON numbers arrive
		{Key: "general.name", Value: "she's"},
	}})
	if err != nil {
		t.Fatal(err)
	}
	want := "narration:\r\n  enabled: false\r\n  chance: 0.7  # keep low\r\ngeneral:\r\n  maxTokens: 2048\r\n  name: 'she''s'\r\nempty:\r\n  []"
	if got := h.read("SkyrimNet.yaml"); got != want {
		t.Errorf("SkyrimNet.yaml =\n%q\nwant\n%q", got, want)
	}
	if f.File.Changed != 2 || f.File.Modified == modified { // enabled and maxTokens differ from their defaults now
		t.Errorf("after saving: changed=%d modified=%d (was %d)", f.File.Changed, f.File.Modified, modified)
	}
	backups, _ := os.ReadDir(filepath.Join(h.store.BackupDir(), "settings", "config"))
	if len(backups) != 1 || !strings.HasPrefix(backups[0].Name(), "SkyrimNet.yaml.") {
		t.Fatalf("backups = %v, want one of SkyrimNet.yaml", backups)
	}
	if b, _ := os.ReadFile(filepath.Join(h.store.BackupDir(), "settings", "config", backups[0].Name())); string(b) != orig {
		t.Error("the backup isn't the file as it was")
	}

	// Adding a setting the file lacks keeps the rest of the file as it was.
	f, err = h.settings.Save("game", Edit{Modified: f.File.Modified, Changes: []Change{{Key: "general.newThing", Value: "later"}}})
	if err != nil {
		t.Fatal(err)
	}
	want = "narration:\r\n  enabled: false\r\n  chance: 0.7  # keep low\r\ngeneral:\r\n  maxTokens: 2048\r\n  name: 'she''s'\r\n  newThing: later\r\nempty:\r\n  []"
	if got := h.read("SkyrimNet.yaml"); got != want {
		t.Errorf("after adding a setting =\n%q\nwant\n%q", got, want)
	}

	// Refusals, each leaving the file alone.
	for _, c := range []struct {
		name string
		edit Edit
		kind apperr.Kind
	}{
		{"wrong kind", Edit{Modified: f.File.Modified, Changes: []Change{{Key: "narration.enabled", Value: "yes"}}}, apperr.Invalid},
		{"fraction for a whole number", Edit{Modified: f.File.Modified, Changes: []Change{{Key: "general.maxTokens", Value: 1.5}}}, apperr.Invalid},
		{"unknown setting", Edit{Modified: f.File.Modified, Changes: []Change{{Key: "nope", Value: 1}}}, apperr.Invalid},
		{"reset without a default", Edit{Modified: f.File.Modified, Changes: []Change{{Key: "general.name", Reset: true}}}, apperr.Invalid},
		{"stale", Edit{Modified: 1, Changes: []Change{{Key: "narration.enabled", Value: true}}}, apperr.Conflict},
		{"nothing", Edit{Modified: f.File.Modified}, apperr.Invalid},
	} {
		_, err := h.settings.Save("game", c.edit)
		kindOf(t, err, c.kind)
		if h.read("SkyrimNet.yaml") != want {
			t.Errorf("%s: the file changed", c.name)
		}
	}
	_, err = h.settings.Save("Missing", Edit{Modified: 1, Changes: []Change{{Key: "a", Value: 1}}})
	kindOf(t, err, apperr.NotFound)
	h.running = true
	_, err = h.settings.Save("game", Edit{Modified: f.File.Modified, Changes: []Change{{Key: "narration.enabled", Value: true}}})
	kindOf(t, err, apperr.Conflict)
}

func TestSecretsCannotBeEditedInBulk(t *testing.T) {
	h := newHarness(t)
	path := h.config("Providers.yaml", "providers:\n  - id: openrouter\n    api_key: SECRET\n")
	info, _ := os.Stat(path)
	_, err := h.settings.Save("Providers", Edit{Modified: info.ModTime().UnixMilli(), Changes: []Change{{Key: "providers", Value: []any{}}}})
	kindOf(t, err, apperr.Invalid)
	if err != nil && strings.Contains(err.Error(), "SECRET") {
		t.Error("the error repeats the secret")
	}
}

func TestPatchValuesKeepsStyle(t *testing.T) {
	orig := "name: wave\n" +
		"description: >-\n" +
		"  Occasional nearby narration when the player conducts a spell into a weapon.\n" +
		"  Kept rare so combat casting does not spam the scene.\n" +
		"\n" +
		"response:\n" +
		"  type: direct_narration\n" +
		"  content: |\n" +
		"    Line one.\n" +
		"    Line two.\n" +
		"  # who hears it\n" +
		"  targetScope: player   # the player only\n" +
		"quoted: 'it''s'\n" +
		"enabled: true\n"
	d0, _ := parse([]byte(orig))
	d1, _ := parse([]byte(orig))
	set := func(key string) *yaml.Node { return findNode(d1.doc, key) }
	set("description").Value = "Occasional narration.\nA second paragraph that is long enough that it has to be wrapped onto another line here."
	set("response.content").Value = "Line one.\nLine 2.\n"
	set("response.targetScope").Value = "everyone"
	set("quoted").Value = "she's"
	set("enabled").Value, set("enabled").Tag = "false", "!!bool"
	got, ok := patchValues(orig, d0.doc, d1.doc)
	want := "name: wave\n" +
		"description: >-\n" +
		"  Occasional narration.\n" +
		"\n" +
		"  A second paragraph that is long enough that it has to be wrapped onto\n" + // the original's width
		"  another line here.\n" +
		"\n" +
		"response:\n" +
		"  type: direct_narration\n" +
		"  content: |\n" +
		"    Line one.\n" +
		"    Line 2.\n" +
		"  # who hears it\n" +
		"  targetScope: everyone   # the player only\n" +
		"quoted: 'she''s'\n" +
		"enabled: false\n"
	if !ok || got != want {
		t.Errorf("patched (%v) =\n%s\nwant\n%s", ok, got, want)
	}
	var a, b any
	yaml.Unmarshal([]byte(got), &a)
	out, _ := d1.encode()
	yaml.Unmarshal(out, &b)
	if !reflect.DeepEqual(a, b) {
		t.Errorf("patched text means %v, want %v", a, b)
	}
}

func TestKeepFormattingMergesAroundComments(t *testing.T) {
	orig := "# top\na: 1\n\n# two\nb: 2\nc:\n  - x\n"
	before := "a: 1\nb: 2\nc:\n  - x\n"
	after := "a: 1\nb: 3\nc:\n  - x\n  - y\n"
	want := "# top\na: 1\n\n# two\nb: 3\nc:\n  - x\n  - y\n"
	if got := keepFormatting(orig, before, after); got != want {
		t.Errorf("merged =\n%s\nwant\n%s", got, want)
	}
	if got := keepFormatting(orig, before, before); got != orig {
		t.Errorf("unchanged =\n%s\nwant the original", got)
	}
}

func TestQuickToggles(t *testing.T) {
	h := newHarness(t)
	h.config("defaults_manifest.json", `{"fields": {"game": {"narration.enabled": true, "general.globalAIEnabled": true}, "BardSinging": {"bard_singing.enabled": true}}}`)
	h.config("SkyrimNet.yaml", "general:\r\n  globalAIEnabled: true\r\nnarration:\r\n  enabled: true")
	state := func(l ToggleList) map[string][2]bool {
		out := map[string][2]bool{}
		for _, t := range l.Toggles {
			out[t.ID] = [2]bool{t.On, t.Available}
		}
		return out
	}
	list, err := h.settings.Toggles()
	if err != nil {
		t.Fatal(err)
	}
	s := state(list)
	// Bard singing has no file yet, so it shows its default but can't be switched.
	if s["narration"] != [2]bool{true, true} || s["skyrimnet"] != [2]bool{true, true} || s["bard_singing"] != [2]bool{true, false} {
		t.Errorf("toggles = %v", s)
	}
	if len(list.Toggles) != len(toggles) || !list.Toggles[0].Master {
		t.Errorf("toggles = %+v", list.Toggles)
	}
	list, err = h.settings.SetToggle("narration", false)
	if err != nil {
		t.Fatal(err)
	}
	if s = state(list); s["narration"] != [2]bool{false, true} {
		t.Errorf("after switching narration off: %v", s)
	}
	if got := h.read("SkyrimNet.yaml"); got != "general:\r\n  globalAIEnabled: true\r\nnarration:\r\n  enabled: false" {
		t.Errorf("SkyrimNet.yaml = %q", got)
	}
	_, err = h.settings.SetToggle("bard_singing", false)
	kindOf(t, err, apperr.NotFound)
	_, err = h.settings.SetToggle("nope", false)
	kindOf(t, err, apperr.NotFound)
	h.running = true
	_, err = h.settings.SetToggle("narration", true)
	kindOf(t, err, apperr.Conflict)
}

// Every settings file in a real SkyrimNet folder should come back byte for
// byte, and switching a setting should change one line. Reports file names
// and line numbers only, since these files hold API keys.
// Run with HERMAEUS_TEST_SKYRIMNET=/path/to/SkyrimNet go test -run TestRealSettings
func TestRealSettingsRoundTrip(t *testing.T) {
	dir := os.Getenv("HERMAEUS_TEST_SKYRIMNET")
	if dir == "" {
		t.Skip("HERMAEUS_TEST_SKYRIMNET not set")
	}
	configDir := filepath.Join(dir, "config")
	defs := loadDefaults(configDir)
	for _, name := range names(configDir, defs) {
		path := filePath(configDir, name)
		old, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		d, err := parse(old)
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		out, err := d.encode()
		if err != nil || string(out) != string(old) {
			t.Errorf("%s: written back differently (%v)", name, err)
			continue
		}
		// Flip the first true/false setting, and expect one changed line.
		f, err := read(configDir, name, defs)
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		for _, st := range f.Values {
			b, ok := st.Value.(bool)
			if !ok {
				continue
			}
			fileDefs, _ := defs.of(name)
			updated, err := apply(old, name, fileDefs, []Change{{Key: st.Key, Value: !b}})
			if err != nil {
				t.Errorf("%s: %s: %v", name, st.Key, err)
				break
			}
			was, now := lines(string(old)), lines(string(updated))
			changed := 0
			for i := range was {
				if i >= len(now) || was[i] != now[i] {
					changed++
				}
			}
			if len(was) != len(now) || changed != 1 {
				t.Errorf("%s: switching %s changed %d lines", name, st.Key, changed)
			}
			break
		}
	}
}

func TestPinnedFiles(t *testing.T) {
	h := newHarness(t)
	h.config("defaults_manifest.json", `{"fields": {"game": {"narration.enabled": true}, "Memory": {"retrieval.max": 5}}}`)
	h.config("SkyrimNet.yaml", "narration:\n  enabled: true\n")
	for _, name := range []string{"Memory", "game"} {
		if _, err := h.settings.SetPinned(name, true); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := h.settings.SetPinned("../x", true); err == nil {
		t.Error("pinned an invalid name")
	}
	pins, _ := h.settings.SetPinned("memory", false) // regardless of case
	if strings.Join(pins.Pins, ",") != "game" {
		t.Errorf("pins = %v", pins.Pins)
	}
	list, err := h.settings.List()
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range list.Files {
		if f.File.Pinned != (f.File.Name == "game") {
			t.Errorf("%s pinned = %v", f.File.Name, f.File.Pinned)
		}
	}
	if list.Files[0].File.Name != "game" {
		t.Errorf("first file = %s, want the pinned one", list.Files[0].File.Name)
	}
	h.settings.SetPinned("Memory", true)
	list, _ = h.settings.List()
	if list.Files[0].File.Name != "game" || list.Files[1].File.Name != "Memory" {
		t.Errorf("files = %v, want pins first in the order pinned", []string{list.Files[0].File.Name, list.Files[1].File.Name})
	}
	if f, _ := h.settings.Read("game"); !f.File.Pinned {
		t.Error("Read doesn't say game is pinned")
	}
}

func TestSpeechPatterns(t *testing.T) {
	h := newHarness(t)
	if r, err := h.settings.Speech("aya_81A"); err != nil || r.Found {
		t.Fatalf("without the file: %+v, %v", r, err)
	}
	_, err := h.settings.SetSpeech("aya_81A", SpeechEdit{Pattern: "Clipped."})
	kindOf(t, err, apperr.NotFound)

	orig := "enabled: true\r\nrules:\r\n  - id: rule_1\r\n    entityType: npc\r\n    entityId: ria_6d8\r\n    entityDisplayName: ria_6D8\r\n    speechPattern: Earnest.\r\n  - id: global_rule\r\n    entityType: global\r\n    entityId: global\r\n    entityDisplayName: Global (All NPCs)\r\n    speechPattern: \"\""
	path := h.config("UniversalTranslator.yaml", orig)
	past := time.Now().Add(-time.Hour)
	os.Chtimes(path, past, past)

	r, err := h.settings.Speech("ria_6D8") // regardless of case
	if err != nil || !r.Found || !r.Enabled || r.Pattern != "Earnest." || r.Modified != past.UnixMilli() {
		t.Fatalf("Ria's rule = %+v, %v", r, err)
	}
	if r, _ := h.settings.Speech("aya_81A"); r.Pattern != "" {
		t.Errorf("Aya has a pattern already: %+v", r)
	}

	// A new rule goes before the global one; an edit changes one line; an
	// empty pattern removes the rule. Windows line endings stay throughout.
	r, err = h.settings.SetSpeech("aya_81A", SpeechEdit{Modified: r.Modified, Pattern: "Clipped and punchy."})
	if err != nil || r.Pattern != "Clipped and punchy." {
		t.Fatalf("after adding: %+v, %v", r, err)
	}
	got := h.read("UniversalTranslator.yaml")
	if !strings.Contains(got, "    entityId: aya_81a\r\n    entityDisplayName: aya_81A\r\n    speechPattern: Clipped and punchy.\r\n  - id: global_rule") || strings.Contains(got, "\n\n") || !strings.HasPrefix(got, "enabled: true\r\nrules:\r\n  - id: rule_1\r\n") {
		t.Errorf("after adding =\n%q", got)
	}
	r, err = h.settings.SetSpeech("ria_6D8", SpeechEdit{Modified: r.Modified, Pattern: "Earnest and quick."})
	if err != nil {
		t.Fatal(err)
	}
	if got := h.read("UniversalTranslator.yaml"); !strings.Contains(got, "    speechPattern: Earnest and quick.\r\n") {
		t.Errorf("after editing =\n%q", got)
	}
	r, err = h.settings.SetSpeech("aya_81A", SpeechEdit{Modified: r.Modified})
	if err != nil || r.Pattern != "" {
		t.Fatalf("after removing: %+v, %v", r, err)
	}
	if got := h.read("UniversalTranslator.yaml"); strings.Contains(got, "aya") || !strings.Contains(got, "global_rule") {
		t.Errorf("after removing =\n%q", got)
	}
	_, err = h.settings.SetSpeech("ria_6D8", SpeechEdit{Modified: 1, Pattern: "x"})
	kindOf(t, err, apperr.Conflict)
	_, err = h.settings.SetSpeech("../x", SpeechEdit{Modified: r.Modified, Pattern: "x"})
	kindOf(t, err, apperr.Invalid)
	h.running = true
	_, err = h.settings.SetSpeech("ria_6D8", SpeechEdit{Modified: r.Modified, Pattern: "x"})
	kindOf(t, err, apperr.Conflict)
}

// A voice engine's file: its pronunciation list replaced whole, several of
// its fallback voices in one save, and its key and clone cache left alone.
func TestVoiceEngineLists(t *testing.T) {
	h := newHarness(t)
	path := h.config("Inworld.yaml", `endpoint: https://example.invalid
pronunciation_dictionary:
  - original: Jarl
    replacement: Yarl
  - original: " - "
    replacement: ", "
allowed_audio_tags:
  []
api_key_base64: SECRET
voice_fallback:
  race_gender_fallbacks:
    nord_female: femalecommoner
    nord_male: malecommoner
clone_cache:
  entries:
    - clone_key: CLONE
      last_used_epoch: 1
`)
	info, _ := os.Stat(path)
	f, err := h.settings.Save("Inworld", Edit{Modified: info.ModTime().UnixMilli(), Changes: []Change{
		{Key: "pronunciation_dictionary", Value: []any{map[string]any{"original": "Jarl", "replacement": "Yarl"}, map[string]any{"original": "draugr", "replacement": "drawgur"}}},
		{Key: "allowed_audio_tags", Value: []any{"sigh", "laugh"}},
		{Key: "voice_fallback.race_gender_fallbacks.nord_female", Value: "femalenord"},
		{Key: "voice_fallback.race_gender_fallbacks.nord_male", Value: "malenord"},
	}})
	if err != nil {
		t.Fatal(err)
	}
	got := h.read("Inworld.yaml")
	for _, want := range []string{"original: draugr", "replacement: drawgur", "nord_female: femalenord", "nord_male: malenord", "- sigh", "api_key_base64: SECRET", "clone_key: CLONE"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in\n%s", want, got)
		}
	}
	if strings.Contains(got, `" - "`) {
		t.Errorf("a removed pronunciation is still there:\n%s", got)
	}
	// A list holding keys still can't be replaced, and secrets aren't sent.
	_, err = h.settings.Save("Inworld", Edit{Modified: f.File.Modified, Changes: []Change{{Key: "clone_cache.entries", Value: []any{}}}})
	kindOf(t, err, apperr.Invalid)
	for _, v := range f.Values {
		if s := fmt.Sprint(v.Value); strings.Contains(s, "SECRET") || strings.Contains(s, "CLONE") || strings.Contains(s, "example.invalid") {
			t.Errorf("%s is sent: %v", v.Key, s)
		}
	}
}
