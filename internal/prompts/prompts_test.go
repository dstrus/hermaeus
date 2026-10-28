package prompts

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"hermaeus/internal/apperr"
	"hermaeus/internal/layers"
	"hermaeus/internal/store"
	"hermaeus/internal/testenv"
)

type harness struct {
	*testenv.Folder
	t       *testing.T
	running bool
	store   *store.Store
	prompts *Prompts
	base    string // the shipped prompts folder, in an MO2-style mod folder
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	h := &harness{Folder: testenv.NewFolder(t), t: t}
	h.store = store.New("", func() bool { return h.running })
	if _, err := h.store.SetDir(h.Dir); err != nil {
		t.Fatal(err)
	}
	mod := filepath.Join(t.TempDir(), "[NoDelete] SkyrimNet Beta27")
	h.base = filepath.Join(mod, "SKSE", "Plugins", "SkyrimNet", "library", layers.BasePlugin, "prompts")
	writeFile(t, filepath.Join(h.base, "..", "manifest.json"), `{"version": "0.25.3"}`)
	writeFile(t, filepath.Join(h.base, "characters", "lydia_A2C.prompt"), "a bio")
	lib := layers.New(h.store.Dir)
	if _, err := lib.SetModDir(mod); err != nil {
		t.Fatal(err)
	}
	h.prompts = New(h.store, lib)
	return h
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	os.MkdirAll(filepath.Dir(path), 0o755)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func (h *harness) mine(path string) string {
	return filepath.Join(h.Dir, "overlay", "prompts", filepath.FromSlash(path))
}

func (h *harness) backups() []string {
	var out []string
	filepath.WalkDir(filepath.Join(h.Dir, "hermaeus-backups", "overlay"), func(path string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(path, ".bak") {
			out = append(out, path)
		}
		return nil
	})
	return out
}

func kindOf(t *testing.T, err error, want apperr.Kind) {
	t.Helper()
	if kind, ok := apperr.KindOf(err); !ok || kind != want {
		t.Errorf("error = %v, want kind %v", err, want)
	}
}

func TestListEditAndRemove(t *testing.T) {
	h := newHarness(t)
	writeFile(t, filepath.Join(h.base, "dialogue_response.prompt"), "[ system ]\r\nYou are {{ npc.name }}.\r\n")
	writeFile(t, filepath.Join(h.base, "components", "event_history.prompt"), "history")
	writeFile(t, filepath.Join(h.base, "diary_entry.prompt"), "shipped diary")
	writeFile(t, filepath.Join(h.base, "_history", "old.prompt.backup.1"), "old")
	writeFile(t, filepath.Join(h.Dir, "library", "diary.pack", "prompts", "Diary_Entry.prompt"), "pack diary")
	writeFile(t, filepath.Join(h.Dir, "library", "diary.pack", "manifest.json"), `{"title": "Diary Pack"}`)
	writeFile(t, filepath.Join(h.Dir, layers.BasePlugin+"-unused"), "")
	writeFile(t, filepath.Join(h.Dir, "content-registry.json"), `{"priority": ["diary.pack"]}`)

	list, err := h.prompts.List()
	if err != nil {
		t.Fatal(err)
	}
	byPath := map[string]Prompt{}
	for _, p := range list.Prompts {
		byPath[p.Path] = p
	}
	source := func(i int) string { return list.Layers[i].Source }
	if d := byPath["Diary_Entry.prompt"]; source(d.Layer) != "Diary Pack" || len(d.Layers) != 2 {
		t.Errorf("diary_entry = %+v (layers %v)", d, list.Layers)
	}
	if _, ok := byPath["components/event_history.prompt"]; !ok {
		t.Errorf("component prompt missing from %v", list.Prompts)
	}
	for _, p := range list.Prompts {
		if strings.HasPrefix(p.Path, "characters/") || strings.HasPrefix(p.Path, "_history/") {
			t.Errorf("%s shouldn't be listed", p.Path)
		}
	}
	if m, _ := h.prompts.Search("PACK diary"); strings.Join(m.Paths, ",") != "Diary_Entry.prompt" {
		t.Errorf("search = %v", m.Paths)
	}

	// The first save makes the player's own copy, with the shipped file's line endings.
	const path = "dialogue_response.prompt"
	saved, err := h.prompts.Save(path, Edit{Text: "[ system ]\nYou are {{ npc.name }}. Be brief.\n{{ render_template(\"components\\\\event_history\") }}{{ render_template(\"components\\\\gone\") }}\n"})
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(h.mine(path)); !strings.HasPrefix(string(got), "[ system ]\r\nYou are {{ npc.name }}. Be brief.\r\n") {
		t.Errorf("overlay copy = %q", got)
	}
	if w := strings.Join(saved.Warnings, "\n"); !strings.Contains(w, "components/gone.prompt") || strings.Contains(w, "event_history") {
		t.Errorf("warnings = %v", saved.Warnings)
	}
	if len(saved.Document.Versions) != 2 || source(saved.Document.Versions[0].Layer) != "your dashboard edits" || strings.Contains(saved.Document.Versions[0].Text, "\r") {
		t.Errorf("versions after saving = %+v", saved.Document.Versions)
	}
	modified := saved.Document.Versions[0].Modified

	_, err = h.prompts.Save(path, Edit{Modified: modified, Text: "{% if x %}never closed"})
	kindOf(t, err, apperr.Invalid)
	_, err = h.prompts.Save(path, Edit{Modified: 0, Text: "anything"}) // the page thought there was no copy yet
	kindOf(t, err, apperr.Conflict)
	_, err = h.prompts.Save(path, Edit{Modified: 1, Text: "anything"}) // the copy changed since the page loaded it
	kindOf(t, err, apperr.Conflict)

	saved, err = h.prompts.Save(path, Edit{Modified: modified, Text: "second version\n"})
	if err != nil {
		t.Fatal(err)
	}
	modified = saved.Document.Versions[0].Modified
	if b := h.backups(); len(b) != 1 {
		t.Errorf("backups after a second save = %v, want the first version backed up", b)
	}

	h.running = true
	_, err = h.prompts.RemoveCopy(path, Revert{Modified: modified})
	kindOf(t, err, apperr.Conflict)
	h.running = false
	doc, err := h.prompts.RemoveCopy(path, Revert{Modified: modified})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(h.mine(path)); err == nil {
		t.Error("removing left the overlay copy")
	}
	if len(doc.Versions) != 1 || source(doc.Versions[0].Layer) != "SkyrimNet’s defaults" {
		t.Errorf("versions after removing = %+v", doc.Versions)
	}
	if len(h.backups()) != 2 {
		t.Error("removing didn't back up the copy")
	}
	_, err = h.prompts.RemoveCopy(path, Revert{Modified: modified})
	kindOf(t, err, apperr.NotFound)

	// A prompt only the player has goes away entirely.
	saved, err = h.prompts.Save("mine/own.prompt", Edit{Text: "mine\n"})
	if err != nil || len(saved.Document.Versions) != 1 {
		t.Fatalf("saving a new prompt: %+v, %v", saved.Document.Versions, err)
	}
	if doc, err = h.prompts.RemoveCopy("mine/own.prompt", Revert{Modified: saved.Document.Versions[0].Modified}); err != nil || len(doc.Versions) != 0 {
		t.Errorf("after removing the only copy: %+v, %v", doc.Versions, err)
	}

	for _, bad := range []string{"../x.prompt", "a/../../x.prompt", "config/Memory.yaml", "a.txt", "/a.prompt", "characters/lydia_A2C.prompt", `a\b.prompt`} {
		if _, err := h.prompts.Save(bad, Edit{Text: "x"}); err == nil {
			t.Errorf("saved to %q", bad)
		} else {
			kindOf(t, err, apperr.Invalid)
		}
	}
	_, err = h.prompts.Read("nope.prompt")
	kindOf(t, err, apperr.NotFound)
}

func TestReadFindsEveryLayerIgnoringCase(t *testing.T) {
	h := newHarness(t)
	writeFile(t, filepath.Join(h.base, "Components", "Mood.prompt"), "shipped\r\n")
	writeFile(t, h.mine("components/mood.prompt"), "mine\n")
	doc, err := h.prompts.Read("components/MOOD.prompt")
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Versions) != 2 || doc.Versions[0].Text != "mine\n" || doc.Versions[1].Text != "shipped\n" || doc.Versions[1].Path != "Components/Mood.prompt" {
		t.Errorf("versions = %+v", doc.Versions)
	}
}
