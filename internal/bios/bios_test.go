package bios

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"hermaeus/internal/apperr"
	"hermaeus/internal/install"
	"hermaeus/internal/layers"
	"hermaeus/internal/store"
	"hermaeus/internal/testenv"
)

const testBio = "{% block summary %}Serana is a vampire.{% endblock %}\n\n" +
	"{% block aspirations %}\n- Find her mother\n{% endblock %}\n```\n"

type harness struct {
	*testenv.Folder
	t       *testing.T
	running bool
	store   *store.Store
	library *layers.Library
	bios    *Bios
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	h := &harness{Folder: testenv.NewFolder(t), t: t}
	h.store = store.New("", func() bool { return h.running })
	if _, err := h.store.SetDir(h.Dir); err != nil {
		t.Fatal(err)
	}
	h.library = layers.New(h.store.Dir)
	h.bios = New(h.store, h.library)
	return h
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	os.MkdirAll(filepath.Dir(path), 0o755)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// writeBio puts a save bio for the test playthrough on disk.
func (h *harness) writeBio(rel, content string) string {
	h.t.Helper()
	p := filepath.Join(h.Dir, "saves", testenv.SaveID, "prompts", "characters", filepath.FromSlash(rel))
	writeFile(h.t, p, content)
	return p
}

func (h *harness) read(template string) []Version {
	h.t.Helper()
	r, err := h.bios.Read(testenv.SaveID, template)
	if err != nil {
		h.t.Fatal(err)
	}
	return r.Versions
}

func (h *harness) version(template, kind string) Version {
	h.t.Helper()
	for _, v := range h.read(template) {
		if v.Kind == kind {
			return v
		}
	}
	h.t.Fatalf("no %s bio for %s", kind, template)
	return Version{}
}

func (h *harness) backups(template string) []string {
	dir := filepath.Join(h.Dir, "hermaeus-backups", "bios", testenv.SaveID)
	var out []string
	for _, f := range install.FilesWithSuffix(dir, ".bak") {
		if strings.HasPrefix(f, template+".prompt.") {
			out = append(out, filepath.Join(dir, f))
		}
	}
	return out
}

func TestListAndReadSaveBios(t *testing.T) {
	h := newHarness(t)
	h.writeBio("serana_B74.prompt", testBio)
	h.writeBio("dynamic/serana_B74.dynamic.prompt", strings.Replace(testBio, "a vampire", "an ally", 1))
	h.writeBio("dynamic/aela_697.dynamic.prompt", testBio)

	l, err := h.bios.List(testenv.SaveID)
	if err != nil || len(l.Bios) != 2 || len(l.Bios["serana_B74"]) != 2 || len(l.Bios["aela_697"]) != 1 {
		t.Errorf("list = %v, %v", l.Bios, err)
	}
	versions := h.read("serana_B74")
	if len(versions) != 2 || versions[0].Kind != KindDynamic {
		t.Fatalf("versions = %+v; want the dynamic bio NPCs use first", versions)
	}
	if s := versions[0].Sections; len(s) != 2 || s[0].Name != "summary" || s[0].Text != "Serana is an ally." || s[1].Text != "- Find her mother" {
		t.Errorf("sections = %+v (aspirations should be trimmed)", s)
	}
	if _, err := h.bios.Read(testenv.SaveID, "../../secret"); err == nil {
		t.Error("read a path outside the characters folder")
	}
	if _, err := h.bios.Read("gone", "serana_B74"); err == nil {
		t.Error("read a playthrough that doesn't exist")
	}
}

func TestSavingSectionsChangesOnlyEditedText(t *testing.T) {
	h := newHarness(t)
	p := h.writeBio("serana_B74.prompt", testBio)
	v := h.version("serana_B74", KindBaseline)
	saved, err := h.bios.Save(testenv.SaveID, "serana_B74", KindBaseline, Edit{Modified: v.Modified, Sections: map[string]string{"aspirations": "- Find her mother\n- Cure her father  "}})
	if err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(p)
	want := strings.Replace(testBio, "- Find her mother\n", "- Find her mother\n- Cure her father\n", 1)
	if string(got) != want {
		t.Errorf("file =\n%q\nwant\n%q", got, want)
	}
	if saved.Sections[1].Text != "- Find her mother\n- Cure her father" || saved.Modified == 0 {
		t.Errorf("saved = %+v", saved)
	}
	b := h.backups("serana_B74")
	if len(b) != 1 {
		t.Fatalf("backups = %v", b)
	}
	if c, _ := os.ReadFile(b[0]); string(c) != testBio {
		t.Errorf("backup holds %q, want the original", c)
	}
}

func TestSavingKeepsWindowsLineEndings(t *testing.T) {
	h := newHarness(t)
	crlf := strings.ReplaceAll(testBio, "\n", "\r\n")
	p := h.writeBio("dynamic/serana_B74.dynamic.prompt", crlf)
	v := h.version("serana_B74", KindDynamic)
	if _, err := h.bios.Save(testenv.SaveID, "serana_B74", KindDynamic, Edit{Modified: v.Modified, Sections: map[string]string{"aspirations": "- One\n- Two"}}); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(p)
	if want := strings.Replace(crlf, "- Find her mother", "- One\r\n- Two", 1); string(got) != want {
		t.Errorf("file = %q, want %q", got, want)
	}
}

func TestSavesRefusedWhenUnsafe(t *testing.T) {
	h := newHarness(t)
	p := h.writeBio("serana_B74.prompt", testBio)
	modified := h.version("serana_B74", KindBaseline).Modified
	save := func(template string, e Edit) apperr.Kind {
		t.Helper()
		_, err := h.bios.Save(testenv.SaveID, template, KindBaseline, e)
		k, _ := apperr.KindOf(err)
		return k
	}
	x := map[string]string{"summary": "x"}
	if k := save("serana_B74", Edit{Modified: modified, Sections: map[string]string{"summary": "x", "hobbies": "y"}}); k != apperr.Invalid {
		t.Errorf("no such section: %v", k)
	}
	if k := save("../serana_B74", Edit{Modified: modified, Sections: x}); k != apperr.Invalid {
		t.Errorf("path as template: %v", k)
	}
	if k := save("aela_697", Edit{Modified: modified, Sections: x}); k != apperr.NotFound {
		t.Errorf("no such bio: %v", k)
	}
	if k := save("serana_B74", Edit{Modified: modified}); k != apperr.Invalid {
		t.Errorf("nothing to change: %v", k)
	}
	h.running = true
	if k := save("serana_B74", Edit{Modified: modified, Sections: x}); k != apperr.Conflict {
		t.Errorf("while Skyrim runs: %v", k)
	}
	h.running = false
	// SkyrimNet rewrote the file after it was opened.
	later := time.Now().Add(time.Minute)
	os.Chtimes(p, later, later)
	if k := save("serana_B74", Edit{Modified: modified, Sections: x}); k != apperr.Conflict {
		t.Errorf("changed since opened: %v", k)
	}
	if got, _ := os.ReadFile(p); string(got) != testBio {
		t.Errorf("a refused save changed the file: %q", got)
	}
	if len(h.backups("serana_B74")) != 0 {
		t.Error("a refused save made a backup")
	}
}

func TestSavingRaw(t *testing.T) {
	h := newHarness(t)
	p := h.writeBio("serana_B74.prompt", strings.ReplaceAll(testBio, "\n", "\r\n"))
	raw := "{% block summary %}Rewritten.{% endblock %}\n{% block mood %}Calm{% endblock %}\n"
	v, err := h.bios.Save(testenv.SaveID, "serana_B74", KindBaseline, Edit{Modified: h.version("serana_B74", KindBaseline).Modified, Raw: &raw})
	if err != nil {
		t.Fatal(err)
	}
	if got, _ := os.ReadFile(p); string(got) != "{% block summary %}Rewritten.{% endblock %}\r\n{% block mood %}Calm{% endblock %}\r\n" {
		t.Errorf("file = %q, want the new text with the file's CRLF line endings", got)
	}
	if !strings.Contains(v.Warning, "aspirations") || len(v.Sections) != 2 || !strings.Contains(v.Raw, "{% block mood %}") {
		t.Errorf("result = %+v", v)
	}
	before, _ := os.ReadFile(p)
	broken := "{% block summary %}Broken"
	if _, err := h.bios.Save(testenv.SaveID, "serana_B74", KindBaseline, Edit{Modified: v.Modified, Raw: &broken}); err == nil {
		t.Error("a bio with broken markup was accepted")
	}
	if after, _ := os.ReadFile(p); string(after) != string(before) {
		t.Error("a bio with broken markup was written")
	}
}

func bio(summary string) string { return "{% block summary %}" + summary + "{% endblock %}\r\n" }

// withLibrary gives the harness an MO2-style mod folder holding SkyrimNet's
// default bios, two installed plugins (one disabled) and a stray library
// folder SkyrimNet didn't install.
func (h *harness) withLibrary() {
	h.t.Helper()
	mod := filepath.Join(h.t.TempDir(), "[NoDelete] SkyrimNet Beta27")
	base := filepath.Join(mod, "SKSE", "Plugins", "SkyrimNet", "library", layers.BasePlugin, "prompts", "characters")
	for _, t := range []string{"aela_697", "lydia_A2C", "ralof_2F6", "serana_B74", "bandit_generic"} {
		writeFile(h.t, filepath.Join(base, t+".prompt"), bio("Base "+t))
	}
	chars := func(parts ...string) string {
		return filepath.Join(append(append([]string{h.Dir}, parts...), "prompts", "characters")...)
	}
	writeFile(h.t, filepath.Join(chars("library", "bios.pack"), "lydia_A2C.prompt"), bio("Pack Lydia"))
	writeFile(h.t, filepath.Join(h.Dir, "library", "bios.pack", "manifest.json"), `{"title": "Lydia Pack"}`)
	writeFile(h.t, filepath.Join(chars("library", "bios.off"), "ralof_2F6.prompt"), bio("Disabled Ralof"))
	writeFile(h.t, filepath.Join(chars("library", "bios.stray"), "aela_697.prompt"), bio("Stray Aela"))
	writeFile(h.t, filepath.Join(chars("overlay"), "aela_697.prompt"), bio("Overlay Aela"))
	writeFile(h.t, filepath.Join(h.Dir, "content-registry.json"), `{
		"plugins": {"bios.pack": {"enabled": true}, "bios.off": {"enabled": false}},
		"priority": ["bios.off", "bios.pack"]}`)
	if _, err := h.library.SetModDir(mod); err != nil {
		h.t.Fatal(err)
	}
}

func TestDefaultBioComesFromHighestLayer(t *testing.T) {
	h := newHarness(t)
	h.withLibrary()
	for template, want := range map[string]string{
		"aela_697":   "Overlay Aela|your dashboard edits|",       // overlay beats the stray folder and the base
		"lydia_A2C":  "Pack Lydia|Lydia Pack|",                   // enabled plugin beats the base
		"ralof_2F6":  "Base ralof_2F6|SkyrimNet’s default bios|", // disabled plugin skipped
		"bandit_7AF": "Base bandit_generic|SkyrimNet’s default bios|bandit_generic",
	} {
		d, err := h.bios.Default(template)
		if err != nil || d == nil {
			t.Errorf("%s: default = %v, %v", template, d, err)
			continue
		}
		if got := fmt.Sprintf("%s|%s|%s", d.Sections[0].Text, d.Source, d.Shared); got != want {
			t.Errorf("%s: got %s, want %s", template, got, want)
		}
	}
	if d, _ := h.bios.Default("nobody_000"); d != nil {
		t.Errorf("unknown template has a default bio: %+v", d)
	}
	if v := h.read("lydia_A2C"); len(v) != 1 || v[0].Kind != KindDefault {
		t.Errorf("a character with no save bio reads %+v", v)
	}
}

func TestCreateSaveBioCopiesDefault(t *testing.T) {
	h := newHarness(t)
	h.withLibrary()
	h.running = true
	if _, err := h.bios.Create(testenv.SaveID, "lydia_A2C"); err == nil {
		t.Error("created while Skyrim runs")
	}
	h.running = false
	v, err := h.bios.Create(testenv.SaveID, "lydia_A2C")
	if err != nil || v.Kind != KindBaseline {
		t.Fatalf("create = %+v, %v", v, err)
	}
	got, _ := os.ReadFile(filepath.Join(h.Dir, "saves", testenv.SaveID, "prompts", "characters", "lydia_A2C.prompt"))
	if string(got) != bio("Pack Lydia") {
		t.Errorf("save bio = %q, want the pack's bio byte for byte", got)
	}
	l, _ := h.bios.List(testenv.SaveID)
	if fmt.Sprint(l.Bios["lydia_A2C"]) != "[baseline]" {
		t.Errorf("save bios for lydia = %v", l.Bios["lydia_A2C"])
	}
	if _, err := h.bios.Create(testenv.SaveID, "lydia_A2C"); err == nil {
		t.Error("created a second baseline")
	}
	if _, err := h.bios.Create(testenv.SaveID, "nobody_000"); err == nil {
		t.Error("created a bio from nothing")
	}
	x := "x"
	if _, err := h.bios.Save(testenv.SaveID, "aela_697", KindDefault, Edit{Raw: &x}); err == nil {
		t.Error("saved over a default bio")
	}
}

func TestHistoryAndRestore(t *testing.T) {
	h := newHarness(t)
	p := h.writeBio("serana_B74.prompt", testBio)
	// SkyrimNet replaced the bio once, keeping a copy; Hermaeus then saved it twice.
	hist := filepath.Join(h.Dir, "overlay", "_history", "prompts", "characters")
	writeFile(t, filepath.Join(hist, "serana_B74.prompt.backup.1700000000"), "{% block summary %}First.{% endblock %}\r\n")
	writeFile(t, filepath.Join(hist, "serana_B74.prompt.backup.1700000500"), "{% block summary %}Second.{% endblock %}\n")
	now := time.Now()
	os.Chtimes(p, now, time.Unix(1700000500, 0)) // the save's file was written when the second copy was taken
	summarize := func() []string {
		t.Helper()
		hr, err := h.bios.History(testenv.SaveID, "serana_B74", KindBaseline)
		if err != nil {
			t.Fatal(err)
		}
		var got []string
		for _, e := range hr.Earlier {
			got = append(got, fmt.Sprintf("%s:%s:%s", e.Source, e.Playthrough, Parse(e.Raw)[0].Text))
		}
		return got
	}
	// The copy taken as this save's file was written is this playthrough's;
	// the older one can't be placed (no events to go by).
	if got, want := summarize(), []string{"skyrimnet:this:Second.", "skyrimnet:unclear:First."}; fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("history =\n %v\nwant\n %v", got, want)
	}

	for _, text := range []string{"Third.", "Fourth."} {
		v := h.version("serana_B74", KindBaseline)
		if _, err := h.bios.Save(testenv.SaveID, "serana_B74", KindBaseline, Edit{Modified: v.Modified, Sections: map[string]string{"summary": text}}); err != nil {
			t.Fatal(err)
		}
		time.Sleep(2 * time.Millisecond)
	}
	// Newest first: Hermaeus's two backups (the earlier text each time), then
	// SkyrimNet's copies, which no longer match the rewritten file's time.
	want := []string{"hermaeus:this:Third.", "hermaeus:this:Serana is a vampire.", "skyrimnet:unclear:Second.", "skyrimnet:unclear:First."}
	if got := summarize(); fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("history =\n %v\nwant\n %v", got, want)
	}
	hr, err := h.bios.History(testenv.SaveID, "serana_B74", KindBaseline)
	if err != nil {
		t.Fatal(err)
	}
	if hr.Earlier[0].Current {
		t.Error("the version before the last save is marked current")
	}

	v := h.version("serana_B74", KindBaseline)
	restored, err := h.bios.RestoreEarlier(testenv.SaveID, "serana_B74", KindBaseline, Restore{ID: hr.Earlier[2].ID, Modified: v.Modified})
	if err != nil {
		t.Fatal(err)
	}
	if restored.Sections[0].Text != "Second." {
		t.Errorf("restored = %+v", restored)
	}
	if got, _ := os.ReadFile(p); string(got) != "{% block summary %}Second.{% endblock %}\n" {
		t.Errorf("file after restore = %q", got)
	}
	if _, err := h.bios.RestoreEarlier(testenv.SaveID, "serana_B74", KindBaseline, Restore{ID: "nope", Modified: restored.Modified}); err == nil {
		t.Error("restored a version that doesn't exist")
	}
}

func TestCatalogListsEveryLayersBios(t *testing.T) {
	h := newHarness(t)
	mod := filepath.Join(t.TempDir(), "[NoDelete] SkyrimNet Beta27")
	base := filepath.Join(mod, "SKSE", "Plugins", "SkyrimNet", "library", layers.BasePlugin)
	writeFile(t, filepath.Join(base, "manifest.json"), `{"version": "0.25.3"}`)
	writeFile(t, filepath.Join(base, "prompts", "characters", "lydia_A2C.prompt"), "base")
	writeFile(t, filepath.Join(base, "prompts", "characters", "bandit_generic.prompt"), "base")
	writeFile(t, filepath.Join(base, "prompts", "characters", "dynamic", "lydia_A2C.dynamic.prompt"), "not a default")
	writeFile(t, filepath.Join(h.Dir, "library", "bios.pack", "prompts", "characters", "Lydia_A2C.prompt"), "pack")
	writeFile(t, filepath.Join(h.Dir, "library", "bios.pack", "manifest.json"), `{"title": "Lydia Pack"}`)
	writeFile(t, filepath.Join(h.Dir, "content-registry.json"), `{"priority": ["bios.pack"]}`)
	if _, err := h.library.SetModDir(mod); err != nil {
		t.Fatal(err)
	}
	c, err := h.bios.CatalogOf()
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, e := range c.Bios {
		got = append(got, e.Template+"@"+c.Layers[e.Layer].Source)
	}
	if want := "bandit_generic@SkyrimNet’s defaults,Lydia_A2C@Lydia Pack"; strings.Join(got, ",") != want {
		t.Errorf("catalog = %v, want %s", got, want)
	}
}
