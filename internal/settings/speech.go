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

// Universal Translator, part of SkyrimNet, rewrites what NPCs say in a
// speech pattern: one for every NPC, and one per NPC, faction, or race. Its
// rules are in config/UniversalTranslator.yaml under rules:, each with an
// id, entityType, entityId, and speechPattern. An NPC's rule is keyed by
// their bio template in lower case, the way SkyrimNet's dashboard writes
// it. Hermaeus edits an NPC's rule from their Overview, keeping the rest of
// the file as it is.

const translatorFile = "UniversalTranslator"

var templateRe = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)

// SpeechRule is an NPC's Universal Translator speech pattern.
type SpeechRule struct {
	Found    bool   `json:"found"`            // the file exists, so a pattern can be set
	Enabled  bool   `json:"enabled"`          // Universal Translator is switched on
	Pattern  string `json:"pattern"`          // "" when the NPC has no rule
	Global   string `json:"global,omitempty"` // the pattern for every NPC
	Modified int64  `json:"modified"`         // the file's time, sent back on save
}

// SpeechEdit sets an NPC's pattern; an empty one removes their rule.
type SpeechEdit struct {
	Modified int64  `json:"modified"`
	Pattern  string `json:"pattern"`
}

func (s *Settings) translatorPath() (string, error) {
	dir, err := s.configDir()
	if err != nil {
		return "", err
	}
	return filePath(dir, translatorFile), nil
}

// rulesOf returns the rules list, or nil if the file has none.
func rulesOf(doc *yaml.Node) *yaml.Node {
	if n := findNode(doc, "rules"); n != nil && n.Kind == yaml.SequenceNode {
		return n
	}
	return nil
}

func ruleField(rule *yaml.Node, key string) string {
	if n := findNode(rule, key); n != nil && n.Kind == yaml.ScalarNode {
		return n.Value
	}
	return ""
}

// npcRule finds the first rule for the template.
func npcRule(rules *yaml.Node, template string) (int, *yaml.Node) {
	if rules == nil {
		return -1, nil
	}
	for i, r := range rules.Content {
		if r.Kind == yaml.MappingNode && ruleField(r, "entityType") == "npc" && strings.EqualFold(ruleField(r, "entityId"), template) {
			return i, r
		}
	}
	return -1, nil
}

// Speech reads an NPC's speech pattern.
func (s *Settings) Speech(template string) (SpeechRule, error) {
	if !templateRe.MatchString(template) {
		return SpeechRule{}, apperr.New(apperr.Invalid, "Invalid bio template %q", template)
	}
	path, err := s.translatorPath()
	if err != nil {
		return SpeechRule{}, err
	}
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return SpeechRule{}, nil
	}
	if err != nil {
		return SpeechRule{}, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return SpeechRule{}, err
	}
	d, err := parse(b)
	if err != nil {
		return SpeechRule{}, apperr.New(apperr.Invalid, "Could not read %s: %v", filepath.Base(path), err)
	}
	out := SpeechRule{Found: true, Modified: info.ModTime().UnixMilli(), Enabled: ruleField(d.doc, "enabled") != "false"}
	rules := rulesOf(d.doc)
	if _, r := npcRule(rules, template); r != nil {
		out.Pattern = ruleField(r, "speechPattern")
	}
	if rules != nil {
		for _, r := range rules.Content {
			if ruleField(r, "entityType") == "global" {
				out.Global = ruleField(r, "speechPattern")
			}
		}
	}
	return out, nil
}

// SetSpeech sets an NPC's speech pattern, or removes their rule when the
// pattern is empty. A new rule goes after the other NPC rules and before the
// global one, as the dashboard keeps them.
func (s *Settings) SetSpeech(template string, e SpeechEdit) (SpeechRule, error) {
	if !templateRe.MatchString(template) {
		return SpeechRule{}, apperr.New(apperr.Invalid, "Invalid bio template %q", template)
	}
	pattern := strings.TrimSpace(textfile.Normalize(e.Pattern))
	what := "Saved the speech pattern for " + template
	if pattern == "" {
		what = "Removed the speech pattern for " + template
	}
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
		if info.ModTime().UnixMilli() != e.Modified {
			return apperr.New(apperr.Conflict, "Universal Translator's settings changed after you opened them. Reload to see the latest, then make your change again.")
		}
		old, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		changed := true
		updated, err := rewrite(old, translatorFile, func(doc *yaml.Node) error {
			rules := rulesOf(doc)
			i, rule := npcRule(rules, template)
			switch {
			case pattern == "" && rule == nil:
				changed = false
			case pattern == "":
				rules.Content = append(rules.Content[:i], rules.Content[i+1:]...)
			case rule != nil:
				if n := findNode(rule, "speechPattern"); n != nil {
					n.SetString(pattern)
				} else {
					rule.Content = append(rule.Content, scalar("speechPattern"), stringNode(pattern))
				}
			default:
				if rules == nil {
					rules = &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
					doc.Content = append(doc.Content, scalar("rules"), rules)
				}
				rule = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map", Content: []*yaml.Node{
					scalar("id"), scalar("rule_" + strconv.FormatInt(time.Now().UnixMilli(), 10)),
					scalar("entityType"), scalar("npc"),
					scalar("entityId"), scalar(strings.ToLower(template)),
					scalar("entityDisplayName"), scalar(template),
					scalar("speechPattern"), stringNode(pattern),
				}}
				at := len(rules.Content)
				for j, r := range rules.Content {
					if ruleField(r, "entityType") == "global" {
						at = j
						break
					}
				}
				rules.Content = append(rules.Content[:at], append([]*yaml.Node{rule}, rules.Content[at:]...)...)
			}
			return nil
		})
		if err != nil || !changed {
			return err
		}
		rel, _ := filepath.Rel(s.store.Dir(), path)
		if _, err := textfile.Backup(filepath.Join(s.store.BackupDir(), "settings", rel), old, 0o600); err != nil {
			return apperr.New(apperr.Conflict, "Could not back up %s, so nothing was changed: %v", filepath.Base(path), err)
		}
		return textfile.WriteAtomic(path, updated, info.Mode().Perm())
	})
	if err != nil {
		return SpeechRule{}, err
	}
	return s.Speech(template)
}

func stringNode(v string) *yaml.Node {
	n := &yaml.Node{}
	n.SetString(v)
	return n
}
