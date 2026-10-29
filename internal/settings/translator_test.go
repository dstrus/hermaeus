package settings

import (
	"os"
	"strings"
	"testing"
	"time"

	"hermaeus/internal/apperr"
)

const translatorYAML = "enabled: true\r\npluginFilterMode: blacklist\r\npluginFilterPlugins: ~\r\nrules:\r\n" +
	"  - id: rule_1\r\n    entityType: npc\r\n    entityId: aya_81a\r\n    entityDisplayName: aya_81A\r\n    speechPattern: Fast, clipped, and punchy.\r\n" +
	"  - id: rule_2\r\n    entityType: npc\r\n    entityId: ria_6d8\r\n    entityDisplayName: ria_6D8\r\n    speechPattern: Earnest.\r\n" +
	"  - id: global_rule\r\n    entityType: global\r\n    entityId: global\r\n    entityDisplayName: Global (All NPCs)\r\n    speechPattern: \"\"\r\n"

func TestTranslatorReadsTheFile(t *testing.T) {
	h := newHarness(t)
	tr, err := h.settings.Translator()
	if err != nil || tr.Found || len(tr.Rules) != 0 || tr.FilterMode != "blacklist" {
		t.Fatalf("without the file: %+v, %v", tr, err)
	}
	_, err = h.settings.SetTranslator(TranslatorEdit{FilterMode: "blacklist"})
	kindOf(t, err, apperr.NotFound)

	h.config("UniversalTranslator.yaml", translatorYAML)
	h.config("SkyrimNet.yaml", "dbvo:\n  enabled: true\n  voiceSilentNPCs: false\n")
	tr, err = h.settings.Translator()
	if err != nil || !tr.Found || !tr.Enabled || tr.Path != "config/UniversalTranslator.yaml" {
		t.Fatalf("Translator() = %+v, %v", tr, err)
	}
	if len(tr.Rules) != 2 || tr.Rules[0].ID != "rule_1" || tr.Rules[0].EntityID != "aya_81a" || tr.Rules[0].Name != "aya_81A" || !tr.Rules[0].Enabled || tr.Rules[1].Pattern != "Earnest." {
		t.Errorf("rules = %+v", tr.Rules)
	}
	if tr.Global != "" || len(tr.FilterPlugins) != 0 {
		t.Errorf("global = %q, plugins = %v", tr.Global, tr.FilterPlugins)
	}
	if !tr.VoiceSilentNPCs.Available || tr.VoiceSilentNPCs.On {
		t.Errorf("voiceSilentNPCs = %+v", tr.VoiceSilentNPCs)
	}
}

func TestTranslatorSavesMinimally(t *testing.T) {
	h := newHarness(t)
	path := h.config("UniversalTranslator.yaml", translatorYAML)
	past := time.Now().Add(-time.Hour)
	os.Chtimes(path, past, past)
	tr, _ := h.settings.Translator()
	edit := func(tr Translator) TranslatorEdit { // with its own copy of the rules
		return TranslatorEdit{Modified: tr.Modified, Enabled: tr.Enabled, Rules: append([]Rule{}, tr.Rules...), Global: tr.Global, FilterMode: tr.FilterMode, FilterPlugins: tr.FilterPlugins}
	}

	// Changing one pattern changes one line, and keeps the line endings.
	e := edit(tr)
	e.Rules[1].Pattern = "Earnest and quick."
	tr, err := h.settings.SetTranslator(e)
	if err != nil {
		t.Fatal(err)
	}
	got := h.read("UniversalTranslator.yaml")
	want := strings.Replace(translatorYAML, "speechPattern: Earnest.", "speechPattern: Earnest and quick.", 1)
	if got != want {
		t.Errorf("after one pattern:\n%q\nwant\n%q", got, want)
	}

	// Reordering moves the rules' blocks whole; the global rule stays last.
	e = edit(tr)
	e.Rules[0], e.Rules[1] = e.Rules[1], e.Rules[0]
	tr, err = h.settings.SetTranslator(e)
	if err != nil || tr.Rules[0].ID != "rule_2" {
		t.Fatalf("after reordering: %+v, %v", tr.Rules, err)
	}
	got = h.read("UniversalTranslator.yaml")
	if !strings.HasPrefix(got, "enabled: true\r\npluginFilterMode: blacklist\r\npluginFilterPlugins: ~\r\nrules:\r\n  - id: rule_2\r\n") || strings.Index(got, "rule_1") > strings.Index(got, "global_rule") || !strings.HasSuffix(got, "speechPattern: \"\"\r\n") {
		t.Errorf("after reordering =\n%q", got)
	}

	// A new rule, switched off, for a faction; the global pattern; the
	// filter as a whitelist of two plugins; the switch off.
	e = edit(tr)
	e.Enabled = false
	e.Global = "Plain Skyrim speech."
	e.FilterMode = "whitelist"
	e.FilterPlugins = []string{"Skyrim.esm", "Dawnguard.esm"}
	e.Rules = append(e.Rules, Rule{EntityType: "faction", EntityID: "ThievesGuildFaction_8AE", Name: "Thieves Guild", EditorID: "ThievesGuildFaction", Pattern: "Sly.", Enabled: false})
	tr, err = h.settings.SetTranslator(e)
	if err != nil || tr.Enabled || tr.Global != "Plain Skyrim speech." || tr.FilterMode != "whitelist" || len(tr.FilterPlugins) != 2 || len(tr.Rules) != 3 {
		t.Fatalf("after adding: %+v, %v", tr, err)
	}
	added := tr.Rules[2]
	if !strings.HasPrefix(added.ID, "rule_") || added.EntityID != "thievesguildfaction_8ae" || added.Enabled {
		t.Errorf("added = %+v", added)
	}
	got = h.read("UniversalTranslator.yaml")
	for _, want := range []string{"enabled: false\r\npluginFilterMode: whitelist\r\npluginFilterPlugins:\r\n  - Skyrim.esm\r\n  - Dawnguard.esm\r\n", "    entityType: faction\r\n    entityId: thievesguildfaction_8ae\r\n    entityDisplayName: Thieves Guild\r\n    entityEditorId: ThievesGuildFaction\r\n    speechPattern: Sly.\r\n    enabled: false\r\n  - id: global_rule", "speechPattern: Plain Skyrim speech.\r\n"} {
		if !strings.Contains(got, want) {
			t.Errorf("after adding, missing %q in\n%q", want, got)
		}
	}
	if strings.Count(got, "enabled: false") != 2 {
		t.Errorf("only the rule that's off says so:\n%q", got)
	}

	// Removing a rule and emptying the filter; a rule switched back on
	// loses nothing but its enabled: false.
	e = edit(tr)
	e.Rules = e.Rules[1:]
	e.Rules[1].Enabled = true
	e.FilterPlugins = nil
	tr, err = h.settings.SetTranslator(e)
	if err != nil || len(tr.Rules) != 2 || tr.Rules[0].ID != "rule_1" || !tr.Rules[1].Enabled {
		t.Fatalf("after removing: %+v, %v", tr.Rules, err)
	}
	got = h.read("UniversalTranslator.yaml")
	if strings.Contains(got, "rule_2") || !strings.Contains(got, "pluginFilterPlugins: ~\r\n") || !strings.Contains(got, "speechPattern: Sly.\r\n    enabled: true\r\n") {
		t.Errorf("after removing =\n%q", got)
	}

	// What's refused.
	e = edit(tr)
	e.Rules[0].EntityType = "voice"
	_, err = h.settings.SetTranslator(e)
	kindOf(t, err, apperr.Invalid)
	e = edit(tr)
	e.FilterMode = "greylist"
	_, err = h.settings.SetTranslator(e)
	kindOf(t, err, apperr.Invalid)
	e = edit(tr)
	e.Modified = 1
	_, err = h.settings.SetTranslator(e)
	kindOf(t, err, apperr.Conflict)
	h.running = true
	_, err = h.settings.SetTranslator(edit(tr))
	kindOf(t, err, apperr.Conflict)
}

func TestTranslatorImportsRules(t *testing.T) {
	h := newHarness(t)
	h.config("UniversalTranslator.yaml", translatorYAML)
	tr, _ := h.settings.Translator()
	res, err := h.settings.ImportRules(Import{Modified: tr.Modified, YAML: "rules:\n  - id: rule_1\n    entityType: npc\n    entityId: RIA_6D8\n    speechPattern: Imported.\n  - entityType: race\n    entityId: nordrace_3ab\n    entityDisplayName: Nords\n    speechPattern: Hearty.\n  - entityType: global\n    speechPattern: Everyone.\n"})
	if err != nil || res.Added != 1 || res.Updated != 1 {
		t.Fatalf("import = %+v, %v", res, err)
	}
	tr = res.Translator
	if len(tr.Rules) != 3 || tr.Rules[1].Pattern != "Imported." || tr.Rules[2].EntityID != "nordrace_3ab" || tr.Rules[2].ID == "rule_1" || tr.Global != "Everyone." {
		t.Errorf("after import: %+v global %q", tr.Rules, tr.Global)
	}
	// A bare list works too; an empty file is refused.
	res, err = h.settings.ImportRules(Import{Modified: tr.Modified, YAML: "- entityType: npc\n  entityId: serana\n  speechPattern: Dry.\n"})
	if err != nil || res.Added != 1 || len(res.Translator.Rules) != 4 {
		t.Fatalf("bare list import = %+v, %v", res, err)
	}
	_, err = h.settings.ImportRules(Import{Modified: res.Translator.Modified, YAML: "enabled: true\n"})
	kindOf(t, err, apperr.Invalid)
	_, err = h.settings.ImportRules(Import{Modified: 1, YAML: "- entityType: npc\n  entityId: x\n"})
	kindOf(t, err, apperr.Conflict)
}
