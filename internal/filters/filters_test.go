package filters

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
	t       *testing.T
	running bool
	filters *Filters
}

func write(t *testing.T, path, content string) {
	t.Helper()
	os.MkdirAll(filepath.Dir(path), 0o755)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// newHarness has SkyrimNet's shipped lists in an MO2-style mod folder, a
// plugin's list in the SkyrimNet folder, and no lists of the player's.
func newHarness(t *testing.T) *harness {
	t.Helper()
	h := &harness{Folder: testenv.NewFolder(t), t: t}
	s := store.New("", func() bool { return h.running })
	if _, err := s.SetDir(h.Dir); err != nil {
		t.Fatal(err)
	}
	mod := filepath.Join(t.TempDir(), "SkyrimNet")
	base := filepath.Join(mod, "SKSE", "Plugins", "SkyrimNet", "library", layers.BasePlugin)
	write(t, filepath.Join(base, "manifest.json"), `{"version": "0.28"}`)
	os.MkdirAll(filepath.Join(base, "prompts"), 0o755) // what marks SkyrimNet's defaults
	write(t, filepath.Join(base, "filters", "animals_no_memories.yaml"), "kind: memory\nname: Animals form no memories\ndescription: Cats, dogs and wolves don't form memories.\nRaceBlacklist:\n  - ACLRaceCat\n  - DogRace\n  - WolfRace\n")
	write(t, filepath.Join(base, "filters", "mannequins_no_memories.yaml"), "kind: memory\nname: Mannequins form no memories\nRaceBlacklist:\n  - ManakinRace\n")
	write(t, filepath.Join(base, "filters", "animals_dont_speak.yaml"), "kind: actor\nname: Animals don't speak\nRaceBlacklist:\n  - DogRace\n")
	write(t, filepath.Join(base, "filters", "combat_grunts.yaml"), "kind: dialogue_rule\nid: combat_grunts\npattern: x\n")
	plugin := filepath.Join(h.Dir, "library", "someone.dogs")
	write(t, filepath.Join(plugin, "manifest.json"), `{"version": "1.0", "title": "Dogs Remember"}`)
	write(t, filepath.Join(plugin, "filters", "dog_memories.yaml"), "kind: memory\nname: Dogs remember\nRaceWhitelist:\n  - DogRace\n")
	write(t, filepath.Join(h.Dir, "content-registry.json"), `{"schema_version": 1, "plugins": {"someone.dogs": {"enabled": true}}, "priority": ["someone.dogs"]}`)
	lib := layers.New(s.Dir)
	if _, err := lib.SetModDir(mod); err != nil {
		t.Fatal(err)
	}
	h.filters = New(s, lib)
	return h
}

func kindOf(t *testing.T, err error, want apperr.Kind) {
	t.Helper()
	if kind, ok := apperr.KindOf(err); !ok || kind != want {
		t.Errorf("error = %v, want kind %v", err, want)
	}
}

func names(r ListsResponse) string {
	var out []string
	for _, l := range r.Lists {
		out = append(out, l.ID+"@"+l.LayerKind)
	}
	return strings.Join(out, " ")
}

func TestListsAddUp(t *testing.T) {
	h := newHarness(t)
	r, err := h.filters.Lists("memory")
	if err != nil {
		t.Fatal(err)
	}
	if got := names(r); got != "animals_no_memories@base dog_memories@pack mannequins_no_memories@base" {
		t.Errorf("lists = %s", got)
	}
	if r.Lists[1].Source != "Dogs Remember" || r.Lists[0].Description == "" {
		t.Errorf("lists = %+v", r.Lists)
	}
	bl := r.InEffect["RaceBlacklist"]
	if len(bl) != 4 || bl[0].Value != "ACLRaceCat" || bl[2].Value != "ManakinRace" || bl[1].From[0].Name != "Animals form no memories" {
		t.Errorf("race blacklist = %+v", bl)
	}
	if wl := r.InEffect["RaceWhitelist"]; len(wl) != 1 || wl[0].From[0].ID != "dog_memories" {
		t.Errorf("race whitelist = %+v", wl)
	}
	if len(r.InEffect["GenderBlacklist"]) != 0 {
		t.Errorf("gender blacklist = %+v", r.InEffect["GenderBlacklist"])
	}
	a, _ := h.filters.Lists("actor")
	if names(a) != "animals_dont_speak@base" {
		t.Errorf("actor lists = %s", names(a))
	}
	_, err = h.filters.Lists("dialogue_rule")
	kindOf(t, err, apperr.Invalid)
}

func TestYourListsAndCopies(t *testing.T) {
	h := newHarness(t)
	mine := filepath.Join(h.Dir, "overlay", "filters")

	// A new list gets an id from its name, free in every layer.
	r, err := h.filters.Lists("memory")
	if err != nil {
		t.Fatal(err)
	}
	created, list, err := h.filters.Create("memory", ListEdit{Name: "No Bandits", Entries: map[string][]string{"FactionBlacklist": {"BanditFaction", " banditfaction ", ""}}})
	if err != nil || list.ID != "no_bandits_memory" || list.LayerKind != "overlay" || created.Lists[0].ID != "no_bandits_memory" {
		t.Fatalf("created %+v, %v", list, err)
	}
	b, _ := os.ReadFile(filepath.Join(mine, "no_bandits_memory.yaml"))
	if string(b) != "kind: memory\nname: No Bandits\nFactionBlacklist:\n  - BanditFaction\n" {
		t.Errorf("new list =\n%s", b)
	}
	_, again, _ := h.filters.Create("memory", ListEdit{Name: "No bandits"})
	if again.ID != "no_bandits_memory_2" {
		t.Errorf("second id = %s", again.ID)
	}

	// Editing a built-in list saves a copy over it, keeping what this page
	// doesn't manage; removing the copy brings the original back.
	write(t, filepath.Join(mine, "unused.txt"), "")
	r, err = h.filters.Save("memory", "animals_no_memories", ListEdit{Name: "Animals form no memories", Entries: map[string][]string{"RaceBlacklist": {"DogRace", "WolfRace"}}})
	if err != nil {
		t.Fatal(err)
	}
	var copy List
	for _, l := range r.Lists {
		if l.ID == "animals_no_memories" {
			copy = l
		}
	}
	if copy.LayerKind != "overlay" || copy.Replaces != "SkyrimNet’s defaults" || len(copy.Entries["RaceBlacklist"]) != 2 {
		t.Fatalf("copy = %+v", copy)
	}
	for _, e := range r.InEffect["RaceBlacklist"] {
		if e.Value == "ACLRaceCat" {
			t.Error("the copy didn't replace the original")
		}
	}
	past := time.Now().Add(-time.Hour)
	os.Chtimes(filepath.Join(mine, "animals_no_memories.yaml"), past, past)
	_, err = h.filters.Save("memory", "animals_no_memories", ListEdit{Modified: 1, Name: "x"})
	kindOf(t, err, apperr.Conflict)
	r, err = h.filters.Remove("memory", "animals_no_memories", Removal{Modified: past.UnixMilli()})
	if err != nil {
		t.Fatal(err)
	}
	if len(r.InEffect["RaceBlacklist"]) != 4 {
		t.Errorf("after removing the copy: %+v", r.InEffect["RaceBlacklist"])
	}
	var bak []string
	filepath.WalkDir(h.filters.store.BackupDir(), func(path string, d os.DirEntry, err error) error {
		if err == nil && strings.Contains(path, filepath.Join("overlay", "filters", "animals_no_memories.yaml")) {
			bak = append(bak, path)
		}
		return nil
	})
	if len(bak) != 1 { // the copy, kept when it was removed; creating it backed up nothing
		t.Errorf("backups = %v", bak)
	}

	// What's refused.
	_, err = h.filters.Remove("memory", "mannequins_no_memories", Removal{})
	kindOf(t, err, apperr.NotFound)
	_, err = h.filters.Save("memory", "../x", ListEdit{Name: "x"})
	kindOf(t, err, apperr.Invalid)
	_, err = h.filters.Save("memory", "x", ListEdit{})
	kindOf(t, err, apperr.Invalid)
	h.running = true
	_, _, err = h.filters.Create("memory", ListEdit{Name: "y"})
	kindOf(t, err, apperr.Conflict)
}
