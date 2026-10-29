package plugins

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

func write(t *testing.T, path, content string) {
	t.Helper()
	os.MkdirAll(filepath.Dir(path), 0o755)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// A registry as SkyrimNet writes it: sorted keys, two-space indents, no
// final newline. "high" has the highest priority; "off" is switched off.
const registryJSON = `{
  "items": {},
  "pins": {},
  "plugins": {
    "author.high": {
      "enabled": true,
      "enabled_explicit": true,
      "source": "hub"
    },
    "author.low": {
      "enabled": true,
      "source": "external"
    },
    "author.off": {
      "enabled": false,
      "enabled_explicit": true,
      "source": "hub"
    }
  },
  "priority": [
    "author.high",
    "author.off",
    "author.low"
  ],
  "schema_version": 1
}`

type harness struct {
	*testenv.Folder
	running bool
	plugins *Plugins
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	h := &harness{Folder: testenv.NewFolder(t)}
	s := store.New("", func() bool { return h.running })
	if _, err := s.SetDir(h.Dir); err != nil {
		t.Fatal(err)
	}
	mod := filepath.Join(t.TempDir(), "SkyrimNet")
	base := filepath.Join(mod, "SKSE", "Plugins", "SkyrimNet", "library", layers.BasePlugin)
	write(t, filepath.Join(base, "manifest.json"), `{"version": "0.28"}`)
	write(t, filepath.Join(base, "prompts", "dialogue.prompt"), "base")
	write(t, filepath.Join(base, "prompts", "characters", "lydia_a2c.prompt"), "base")
	write(t, filepath.Join(base, "filters", "a.yaml"), "kind: memory")
	lib := filepath.Join(h.Dir, "library")
	write(t, filepath.Join(lib, "author.low", "manifest.json"), `{"title": "Low", "version": "1.0", "author": "author", "tags": ["x"], "type": "bundle"}`)
	write(t, filepath.Join(lib, "author.low", "prompts", "Dialogue.prompt"), "low") // replaces the base's, in another case
	write(t, filepath.Join(lib, "author.high", "manifest.json"), `{"title": "High", "version": "2.0"}`)
	write(t, filepath.Join(lib, "author.high", "prompts", "dialogue.prompt"), "high")
	write(t, filepath.Join(lib, "author.high", "knowledge", "k.yaml"), "k")
	write(t, filepath.Join(lib, "author.high", "prompts", "characters", "lydia_a2c.prompt"), "a bio")
	write(t, filepath.Join(lib, "author.off", "manifest.json"), `{"title": "Off"}`)
	write(t, filepath.Join(lib, "author.off", "prompts", "dialogue.prompt"), "off")
	write(t, filepath.Join(h.Dir, "overlay", "prompts", "dialogue.prompt"), "mine")
	write(t, filepath.Join(h.Dir, "external", "author.new", "manifest.json"), `{"title": "New", "version": "0.1"}`) // not registered yet
	write(t, filepath.Join(h.Dir, "content-registry.json"), registryJSON)
	l := layers.New(s.Dir)
	if _, err := l.SetModDir(mod); err != nil {
		t.Fatal(err)
	}
	os.MkdirAll(filepath.Join(base, "prompts"), 0o755)
	h.plugins = New(s, l)
	return h
}

func kindOf(t *testing.T, err error, want apperr.Kind) {
	t.Helper()
	if kind, ok := apperr.KindOf(err); !ok || kind != want {
		t.Errorf("error = %v, want kind %v", err, want)
	}
}

func ids(l List) string {
	var out []string
	for _, p := range l.Plugins {
		out = append(out, p.ID)
	}
	return strings.Join(out, " ")
}

func TestListLowestFirst(t *testing.T) {
	h := newHarness(t)
	l, err := h.plugins.List()
	if err != nil {
		t.Fatal(err)
	}
	if got := ids(l); got != "skyrimnet.base author.new author.low author.off author.high overlay" {
		t.Fatalf("order = %s", got)
	}
	if n := l.Plugins[1]; n.Kind != "new" || n.Title != "New" || n.Enabled {
		t.Errorf("unregistered = %+v", n)
	}
	base, low, off, high, mine := l.Plugins[0], l.Plugins[2], l.Plugins[3], l.Plugins[4], l.Plugins[5]
	if base.Kind != "base" || !base.Enabled || base.Title != "SkyrimNet’s defaults" {
		t.Errorf("base = %+v", base)
	}
	if low.Title != "Low" || low.Version != "1.0" || low.Source != "external" || low.Author != "author" || low.Type != "bundle" || !low.Found || low.Folder != "library/author.low" {
		t.Errorf("low = %+v", low)
	}
	if off.Enabled || len(off.Replaces) != 0 || len(off.Replaced) != 0 {
		t.Errorf("a plugin switched off overrides nothing: %+v", off)
	}
	// Each prompt replaces the one below it and is replaced by the one above.
	if strings.Join(low.Replaces, ",") != "Dialogue.prompt" || strings.Join(low.Replaced, ",") != "Dialogue.prompt" {
		t.Errorf("low replaces %v, replaced %v", low.Replaces, low.Replaced)
	}
	if strings.Join(high.Replaced, ",") != "dialogue.prompt" || len(mine.Replaced) != 0 || len(mine.Replaces) != 1 {
		t.Errorf("high replaced %v; mine %v / %v", high.Replaced, mine.Replaces, mine.Replaced)
	}
	if len(base.Replaces) != 0 || strings.Join(base.Replaced, ",") != "dialogue.prompt" {
		t.Errorf("base replaces %v, replaced %v", base.Replaces, base.Replaced)
	}
	counts := map[string]int{}
	for _, c := range high.Contents {
		counts[c.Folder] = c.Files
	}
	if counts["prompts"] != 1 || counts["bios"] != 1 || counts["knowledge"] != 1 {
		t.Errorf("high's contents = %+v", high.Contents)
	}
	if mine.Kind != "overlay" || mine.Title != "Your edits" {
		t.Errorf("overlay = %+v", mine)
	}
}

func TestOrderAndSwitch(t *testing.T) {
	h := newHarness(t)
	path := filepath.Join(h.Dir, "content-registry.json")
	past := time.Now().Add(-time.Hour)
	os.Chtimes(path, past, past)
	l, _ := h.plugins.List()

	// The same order writes the file back byte for byte.
	l, err := h.plugins.SetOrder(Order{Modified: l.Modified, Priority: []string{"author.high", "author.off", "author.low"}})
	if err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(path); string(b) != registryJSON {
		t.Errorf("unchanged order rewrote the file:\n%s", b)
	}

	// Low to the top: it now wins, and only the priority list changes.
	l, err = h.plugins.SetOrder(Order{Modified: l.Modified, Priority: []string{"author.low", "author.high", "author.off"}})
	if err != nil {
		t.Fatal(err)
	}
	if got := ids(l); got != "skyrimnet.base author.new author.off author.high author.low overlay" {
		t.Errorf("order = %s", got)
	}
	want := strings.Replace(registryJSON, `"author.high",
    "author.off",
    "author.low"`, `"author.low",
    "author.high",
    "author.off"`, 1)
	if b, _ := os.ReadFile(path); string(b) != want {
		t.Errorf("after reordering:\n%s", b)
	}

	// Switching one on marks it explicit, as the dashboard does.
	l, err = h.plugins.SetEnabled("author.low", Switch{Modified: l.Modified, Enabled: false})
	if err != nil {
		t.Fatal(err)
	}
	if b, _ := os.ReadFile(path); !strings.Contains(string(b), `"author.low": {
      "enabled": false,
      "enabled_explicit": true,
      "source": "external"
    }`) {
		t.Errorf("after switching off:\n%s", b)
	}
	if l.Plugins[4].ID != "author.low" || l.Plugins[4].Enabled {
		t.Errorf("low = %+v", l.Plugins[4])
	}

	// What's refused.
	_, err = h.plugins.SetOrder(Order{Modified: l.Modified, Priority: []string{"author.low", "author.high"}})
	kindOf(t, err, apperr.Conflict)
	_, err = h.plugins.SetOrder(Order{Modified: l.Modified, Priority: []string{"author.low", "author.low", "author.off"}})
	kindOf(t, err, apperr.Conflict)
	_, err = h.plugins.SetEnabled("author.low", Switch{Modified: 1})
	kindOf(t, err, apperr.Conflict)
	_, err = h.plugins.SetEnabled(layers.BasePlugin, Switch{Modified: l.Modified})
	kindOf(t, err, apperr.Invalid)
	_, err = h.plugins.SetEnabled("nobody.nothing", Switch{Modified: l.Modified})
	kindOf(t, err, apperr.NotFound)
	h.running = true
	_, err = h.plugins.SetEnabled("author.low", Switch{Modified: l.Modified, Enabled: true})
	kindOf(t, err, apperr.Conflict)
}
