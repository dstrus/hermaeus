package settings

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

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
	if _, err := s.editTranslator(what, e.Modified, func(doc *yaml.Node) error {
		rules := rulesOf(doc)
		i, rule := npcRule(rules, template)
		switch {
		case pattern == "" && rule == nil:
			// nothing to remove
		case pattern == "":
			rules.Content = append(rules.Content[:i], rules.Content[i+1:]...)
		case rule != nil:
			setField(rule, "speechPattern", stringNode(pattern))
		default:
			if rules == nil {
				rules = &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
				doc.Content = append(doc.Content, scalar("rules"), rules)
			}
			rule = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map", Content: []*yaml.Node{
				scalar("id"), scalar(newRuleID(nil)),
				scalar("entityType"), scalar("npc"),
				scalar("entityId"), scalar(strings.ToLower(template)),
				scalar("entityDisplayName"), scalar(template),
				scalar("speechPattern"), stringNode(pattern),
			}}
			at := len(rules.Content)
			for j, r := range rules.Content {
				if ruleField(r, "entityType") == globalType {
					at = j
					break
				}
			}
			rules.Content = append(rules.Content[:at], append([]*yaml.Node{rule}, rules.Content[at:]...)...)
		}
		return nil
	}); err != nil {
		return SpeechRule{}, err
	}
	return s.Speech(template)
}

// stringNode is a text value; one with line breaks is quoted rather than
// written as a block, so it reads as one line.
func stringNode(v string) *yaml.Node {
	n := &yaml.Node{}
	n.SetString(v)
	if strings.ContainsAny(v, "\n\r\t") {
		n.Style = yaml.DoubleQuotedStyle
	}
	return n
}
