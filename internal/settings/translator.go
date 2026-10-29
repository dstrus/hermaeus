package settings

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"hermaeus/internal/apperr"
	"hermaeus/internal/textfile"

	"go.yaml.in/yaml/v3"
)

// Universal Translator's whole configuration, as its dashboard page shows
// it: the switch, the rules in priority order (the first that matches an
// NPC applies), the pattern for everyone the rules don't name, and the
// plugin filter, which keeps the translator away from (or only on) the
// voice lines some plugins add. It depends on SkyrimNet voicing silent
// NPCs, a setting in SkyrimNet.yaml, so that's read alongside.

// Rule is one Universal Translator rule: a speech pattern for an NPC (by
// bio template), a faction, or a race.
type Rule struct {
	ID         string `json:"id"`
	EntityType string `json:"entityType"` // npc, faction, or race
	EntityID   string `json:"entityId"`   // an NPC's bio template in lower case; a faction's or race's editor ID and the end of its form ID
	Name       string `json:"name"`       // as shown
	EditorID   string `json:"editorId,omitempty"`
	Pattern    string `json:"pattern"` // empty leaves the entity untranslated, whatever the global pattern says
	Enabled    bool   `json:"enabled"` // off, the rule is skipped and those below it can apply
}

// Translator is Universal Translator's configuration.
type Translator struct {
	Found         bool     `json:"found"` // the file exists; until SkyrimNet writes it, nothing can be changed
	Path          string   `json:"path"`  // under the SkyrimNet folder
	Modified      int64    `json:"modified"`
	Enabled       bool     `json:"enabled"`
	Rules         []Rule   `json:"rules"` // in priority order
	Global        string   `json:"global"`
	FilterMode    string   `json:"filterMode"` // blacklist or whitelist
	FilterPlugins []string `json:"filterPlugins"`
	// SkyrimNet voicing silent NPCs (dbvo.voiceSilentNPCs in SkyrimNet.yaml),
	// without which the translator has nothing to speak with.
	VoiceSilentNPCs Dependency `json:"voiceSilentNPCs"`
}

// Dependency is a setting in another file the translator needs.
type Dependency struct {
	Available bool  `json:"available"` // its file exists
	On        bool  `json:"on"`
	Modified  int64 `json:"modified"` // the file's time, for changing it
}

// TranslatorEdit is the configuration as it should be, except the
// dependency, which is changed through its own file.
type TranslatorEdit struct {
	Modified      int64    `json:"modified"`
	Enabled       bool     `json:"enabled"`
	Rules         []Rule   `json:"rules"`
	Global        string   `json:"global"`
	FilterMode    string   `json:"filterMode"`
	FilterPlugins []string `json:"filterPlugins"`
}

// Import is a YAML file of rules to add to those there are.
type Import struct {
	Modified int64  `json:"modified"`
	YAML     string `json:"yaml"`
}

type ImportResult struct {
	Translator Translator `json:"translator"`
	Added      int        `json:"added"`
	Updated    int        `json:"updated"` // rules for an entity that already had one
}

var entityTypes = map[string]bool{"npc": true, "faction": true, "race": true}

const globalType = "global"

// Translator reads the configuration.
func (s *Settings) Translator() (Translator, error) {
	path, err := s.translatorPath()
	if err != nil {
		return Translator{}, err
	}
	out := Translator{Rules: []Rule{}, FilterMode: "blacklist", FilterPlugins: []string{}}
	rel, _ := filepath.Rel(s.store.Dir(), path)
	out.Path = filepath.ToSlash(rel)
	if dep, err := s.voiceSilentNPCs(); err == nil {
		out.VoiceSilentNPCs = dep
	}
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return out, nil
	}
	if err != nil {
		return Translator{}, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return Translator{}, err
	}
	d, err := parse(b)
	if err != nil {
		return Translator{}, apperr.New(apperr.Invalid, "Could not read %s: %v", filepath.Base(path), err)
	}
	out.Found = true
	out.Modified = info.ModTime().UnixMilli()
	out.Enabled = ruleField(d.doc, "enabled") != "false"
	if ruleField(d.doc, "pluginFilterMode") == "whitelist" {
		out.FilterMode = "whitelist"
	}
	if n := findNode(d.doc, "pluginFilterPlugins"); n != nil && n.Kind == yaml.SequenceNode {
		for _, p := range n.Content {
			if p.Kind == yaml.ScalarNode && strings.TrimSpace(p.Value) != "" {
				out.FilterPlugins = append(out.FilterPlugins, p.Value)
			}
		}
	}
	if rules := rulesOf(d.doc); rules != nil {
		for _, r := range rules.Content {
			if r.Kind != yaml.MappingNode {
				continue
			}
			if ruleField(r, "entityType") == globalType {
				out.Global = ruleField(r, "speechPattern")
				continue
			}
			out.Rules = append(out.Rules, ruleOf(r))
		}
	}
	return out, nil
}

func ruleOf(r *yaml.Node) Rule {
	return Rule{
		ID:         ruleField(r, "id"),
		EntityType: ruleField(r, "entityType"),
		EntityID:   ruleField(r, "entityId"),
		Name:       ruleField(r, "entityDisplayName"),
		EditorID:   ruleField(r, "entityEditorId"),
		Pattern:    ruleField(r, "speechPattern"),
		Enabled:    ruleField(r, "enabled") != "false",
	}
}

// voiceSilentNPCs reads the setting the translator depends on.
func (s *Settings) voiceSilentNPCs() (Dependency, error) {
	dir, err := s.configDir()
	if err != nil {
		return Dependency{}, err
	}
	f, err := read(dir, "game", loadDefaults(dir))
	if err != nil {
		return Dependency{}, err
	}
	dep := Dependency{Available: f.File.Exists, Modified: f.File.Modified}
	for _, v := range f.Values {
		if strings.EqualFold(v.Key, "dbvo.voiceSilentNPCs") {
			val := v.Default
			if v.Set {
				val = v.Value
			}
			dep.On = val == true
		}
	}
	return dep, nil
}

var entityIDRe = regexp.MustCompile(`^[A-Za-z0-9_.|-]+$`)

func (r Rule) check() error {
	if !entityTypes[r.EntityType] {
		return apperr.New(apperr.Invalid, "A rule is for an NPC, a faction, or a race, not %q", r.EntityType)
	}
	if r.EntityID == "" {
		return apperr.New(apperr.Invalid, "A rule needs the NPC, faction, or race it's for")
	}
	if !entityIDRe.MatchString(r.EntityID) {
		return apperr.New(apperr.Invalid, "Invalid entity ID %q", r.EntityID)
	}
	return nil
}

// SetTranslator writes the configuration as given. Rules keep their nodes
// (by id), so one whose text didn't change is written as it was, and the
// file's formatting is kept as far as the change allows.
func (s *Settings) SetTranslator(e TranslatorEdit) (Translator, error) {
	if e.FilterMode != "blacklist" && e.FilterMode != "whitelist" {
		return Translator{}, apperr.New(apperr.Invalid, "The plugin filter is a blacklist or a whitelist, not %q", e.FilterMode)
	}
	seen := map[string]bool{}
	for i := range e.Rules {
		r := &e.Rules[i]
		r.EntityID = strings.ToLower(strings.TrimSpace(r.EntityID))
		r.Name = strings.TrimSpace(r.Name)
		r.Pattern = strings.TrimSpace(textfile.Normalize(r.Pattern))
		if err := r.check(); err != nil {
			return Translator{}, err
		}
		if r.ID == "" || seen[r.ID] {
			r.ID = newRuleID(seen)
		}
		seen[r.ID] = true
	}
	e.Global = strings.TrimSpace(textfile.Normalize(e.Global))
	return s.editTranslator("Saved Universal Translator's rules", e.Modified, func(doc *yaml.Node) error {
		setField(doc, "enabled", boolNode(e.Enabled))
		setField(doc, "pluginFilterMode", scalar(e.FilterMode))
		setField(doc, "pluginFilterPlugins", listNode(e.FilterPlugins))
		rules := rulesOf(doc)
		if rules == nil {
			rules = &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
			setField(doc, "rules", rules)
		}
		existing := map[string]*yaml.Node{}
		var global *yaml.Node
		for _, r := range rules.Content {
			if r.Kind != yaml.MappingNode {
				continue
			}
			if ruleField(r, "entityType") == globalType {
				global = r
			} else if id := ruleField(r, "id"); id != "" {
				existing[id] = r
			}
		}
		content := make([]*yaml.Node, 0, len(e.Rules)+1)
		for _, r := range e.Rules {
			n := existing[r.ID]
			if n == nil {
				n = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
			}
			setField(n, "id", scalar(r.ID))
			setField(n, "entityType", scalar(r.EntityType))
			setField(n, "entityId", scalar(r.EntityID))
			setField(n, "entityDisplayName", stringNode(r.Name))
			if r.EditorID != "" || findNode(n, "entityEditorId") != nil {
				setField(n, "entityEditorId", stringNode(r.EditorID))
			}
			setField(n, "speechPattern", stringNode(r.Pattern))
			// Only a rule that's off says so; on is what a rule without the key is.
			if !r.Enabled {
				setField(n, "enabled", boolNode(false))
			} else if findNode(n, "enabled") != nil {
				setField(n, "enabled", boolNode(true))
			}
			content = append(content, n)
		}
		if global == nil {
			global = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map", Content: []*yaml.Node{
				scalar("id"), scalar("global_rule"),
				scalar("entityType"), scalar(globalType),
				scalar("entityId"), scalar(globalType),
				scalar("entityDisplayName"), scalar("Global (All NPCs)"),
			}}
		}
		setField(global, "speechPattern", stringNode(e.Global))
		rules.Content = append(content, global)
		return nil
	})
}

// ImportRules adds the rules in a YAML file (one SkyrimNet's dashboard
// exported, or this file from another install) to those there are. A rule
// for an entity that already has one replaces that rule's pattern; the
// others go last, before the global rule. A global rule in the file sets
// the global pattern.
func (s *Settings) ImportRules(imp Import) (ImportResult, error) {
	var in struct {
		Rules []yaml.Node `yaml:"rules"`
	}
	text := textfile.Normalize(imp.YAML)
	if err := yaml.Unmarshal([]byte(text), &in); err != nil {
		// A bare list of rules counts too.
		var list []yaml.Node
		if err2 := yaml.Unmarshal([]byte(text), &list); err2 != nil {
			return ImportResult{}, apperr.New(apperr.Invalid, "Could not read the file as YAML: %v", err)
		}
		in.Rules = list
	}
	var incoming []Rule
	var global *string
	for i := range in.Rules {
		n := &in.Rules[i]
		if n.Kind != yaml.MappingNode {
			continue
		}
		if ruleField(n, "entityType") == globalType {
			g := ruleField(n, "speechPattern")
			global = &g
			continue
		}
		r := ruleOf(n)
		r.EntityID = strings.ToLower(strings.TrimSpace(r.EntityID))
		if err := r.check(); err != nil {
			return ImportResult{}, err
		}
		incoming = append(incoming, r)
	}
	if len(incoming) == 0 && global == nil {
		return ImportResult{}, apperr.New(apperr.Invalid, "The file has no Universal Translator rules (a rules: list, each with entityType, entityId, and speechPattern)")
	}
	current, err := s.Translator()
	if err != nil {
		return ImportResult{}, err
	}
	if !current.Found {
		return ImportResult{}, apperr.New(apperr.NotFound, "%s.yaml doesn't exist yet. SkyrimNet creates it the next time it starts.", translatorFile)
	}
	if current.Modified != imp.Modified {
		return ImportResult{}, apperr.New(apperr.Conflict, "Universal Translator's settings changed after you opened them. Reload to see the latest, then make your change again.")
	}
	edit := TranslatorEdit{Modified: current.Modified, Enabled: current.Enabled, Rules: current.Rules, Global: current.Global, FilterMode: current.FilterMode, FilterPlugins: current.FilterPlugins}
	result := ImportResult{}
	for _, r := range incoming {
		at := -1
		for i, have := range edit.Rules {
			if have.EntityType == r.EntityType && have.EntityID == r.EntityID {
				at = i
				break
			}
		}
		if at >= 0 {
			edit.Rules[at].Pattern = r.Pattern
			edit.Rules[at].Enabled = r.Enabled
			result.Updated++
		} else {
			r.ID = "" // given on save, so an id from elsewhere can't collide
			edit.Rules = append(edit.Rules, r)
			result.Added++
		}
	}
	if global != nil {
		edit.Global = *global
	}
	result.Translator, err = s.SetTranslator(edit)
	return result, err
}

// editTranslator makes a change to Universal Translator's file, refusing it
// if the file changed since modified, and backing the file up first.
func (s *Settings) editTranslator(what string, modified int64, change func(doc *yaml.Node) error) (Translator, error) {
	err := s.store.ChangeFiles(what, func() error {
		path, err := s.translatorPath()
		if err != nil {
			return err
		}
		info, err := os.Stat(path)
		if os.IsNotExist(err) {
			return apperr.New(apperr.NotFound, "%s.yaml doesn't exist yet. SkyrimNet creates it the next time it starts.", translatorFile)
		}
		if err != nil {
			return err
		}
		if info.ModTime().UnixMilli() != modified {
			return apperr.New(apperr.Conflict, "Universal Translator's settings changed after you opened them. Reload to see the latest, then make your change again.")
		}
		old, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		updated, err := rewrite(old, translatorFile, change)
		if err != nil {
			return err
		}
		if string(updated) == string(old) {
			return nil
		}
		rel, _ := filepath.Rel(s.store.Dir(), path)
		if _, err := textfile.Backup(filepath.Join(s.store.BackupDir(), "settings", rel), old, 0o600); err != nil {
			return apperr.New(apperr.Conflict, "Could not back up %s, so nothing was changed: %v", filepath.Base(path), err)
		}
		return textfile.WriteAtomic(path, updated, info.Mode().Perm())
	})
	if err != nil {
		return Translator{}, err
	}
	return s.Translator()
}

func newRuleID(taken map[string]bool) string {
	for t := time.Now().UnixMilli(); ; t++ {
		if id := "rule_" + strconv.FormatInt(t, 10); !taken[id] {
			return id
		}
	}
}

// setField sets key's value in a mapping, adding the key at the end if it
// isn't there.
func setField(m *yaml.Node, key string, value *yaml.Node) {
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			m.Content[i+1] = value
			return
		}
	}
	m.Content = append(m.Content, scalar(key), value)
}

func boolNode(v bool) *yaml.Node {
	return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: strconv.FormatBool(v)}
}

// listNode is a list of words, or ~ for none, as SkyrimNet writes an empty one.
func listNode(items []string) *yaml.Node {
	if len(items) == 0 {
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!null", Value: "~"}
	}
	n := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
	for _, it := range items {
		n.Content = append(n.Content, stringNode(it))
	}
	return n
}
