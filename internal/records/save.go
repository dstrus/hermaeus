package records

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"hermaeus/internal/apperr"
	"hermaeus/internal/layers"
	"hermaeus/internal/settings"
	"hermaeus/internal/textfile"

	"go.yaml.in/yaml/v3"
)

// RecordEdit is a record's YAML, as the player's copy should have it.
type RecordEdit struct {
	Modified int64  `json:"modified"` // the copy's time when read, 0 for none yet
	Text     string `json:"text"`
}

type Switch struct {
	Modified int64 `json:"modified"`
	Enabled  bool  `json:"enabled"`
}

type Revert struct {
	Modified int64 `json:"modified"`
}

// recordPath checks a record's path: <kind>/<…>.yaml.
func recordPath(rel string) (string, error) {
	rel, err := checkPath(rel)
	if err != nil {
		return "", err
	}
	root, rest, _ := strings.Cut(rel, "/")
	if !isKind(root) || rest == "" || !strings.HasSuffix(strings.ToLower(rel), ".yaml") {
		return "", apperr.New(apperr.Invalid, "Invalid record path %q", rel)
	}
	return rel, nil
}

// SaveRecord saves a record's YAML as the player's copy, which replaces the
// same file in every layer below. It's refused if SkyrimNet would reject or
// misread it.
func (r *Records) SaveRecord(rel string, e RecordEdit) ([]Version, error) {
	rel, err := recordPath(rel)
	if err != nil {
		return nil, err
	}
	text := textfile.Normalize(e.Text)
	if err := checkRecord(rel, text); err != nil {
		return nil, err
	}
	if err := r.saveCopy(rel, e.Modified, "Saved your copy of "+rel, func([]byte) (string, error) { return text, nil }); err != nil {
		return nil, err
	}
	return r.Versions(rel)
}

// SetEnabled switches a record on or off in the player's copy, changing
// only that line of the version it replaces.
func (r *Records) SetEnabled(rel string, s Switch) ([]Version, error) {
	rel, err := recordPath(rel)
	if err != nil {
		return nil, err
	}
	what := "Switched on " + rel
	if !s.Enabled {
		what = "Switched off " + rel
	}
	err = r.saveCopy(rel, s.Modified, what, func(from []byte) (string, error) {
		if len(from) == 0 {
			return "", apperr.New(apperr.NotFound, "No layer has %s", rel)
		}
		out, err := settings.EditYAML(from, func(doc *yaml.Node) error {
			if doc.Kind != yaml.MappingNode {
				return apperr.New(apperr.Invalid, "%s isn't a set of fields, so it can't be switched here", rel)
			}
			v := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: fmt.Sprint(s.Enabled)}
			for i := 0; i+1 < len(doc.Content); i += 2 {
				if doc.Content[i].Value == "enabled" {
					doc.Content[i+1] = v
					return nil
				}
			}
			doc.Content = append(doc.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "enabled"}, v)
			return nil
		})
		if err != nil {
			return "", err
		}
		return textfile.Normalize(string(out)), nil
	})
	if err != nil {
		return nil, err
	}
	return r.Versions(rel)
}

// RevertRecord removes the player's copy, so the version below applies
// again. A backup is kept.
func (r *Records) RevertRecord(rel string, rv Revert) ([]Version, error) {
	rel, err := recordPath(rel)
	if err != nil {
		return nil, err
	}
	err = r.store.ChangeFiles("Removed your copy of "+rel, func() error {
		overlay, err := r.overlay()
		if err != nil {
			return err
		}
		dest := layers.Lookup(overlay, rel)
		if dest == "" {
			return apperr.New(apperr.NotFound, "You don't have your own copy of %s", rel)
		}
		info, err := os.Stat(dest)
		if err != nil {
			return err
		}
		if info.ModTime().UnixMilli() != rv.Modified {
			return apperr.New(apperr.Conflict, "Your copy changed after you opened it. Reload to see the latest.")
		}
		old, err := os.ReadFile(dest)
		if err != nil {
			return err
		}
		if err := r.backup(overlay, dest, old); err != nil {
			return err
		}
		return os.Remove(dest)
	})
	if err != nil {
		return nil, err
	}
	v, err := r.Versions(rel)
	if kind, ok := apperr.KindOf(err); ok && kind == apperr.NotFound {
		return []Version{}, nil
	}
	return v, err
}

func (r *Records) overlay() (string, error) {
	lay := r.library.Layers()
	if len(lay) == 0 || lay[0].Kind != "overlay" {
		return "", apperr.New(apperr.Unavailable, "No SkyrimNet folder chosen yet")
	}
	return lay[0].Root, nil
}

// saveCopy writes the player's copy of rel. build gets the version it
// replaces (their copy, the highest layer's file, or nil for a new one) and
// returns the text with \n line endings; the copy keeps the line endings of
// what it replaces. modified is the copy's time when read, 0 for none.
func (r *Records) saveCopy(rel string, modified int64, what string, build func(from []byte) (string, error)) error {
	return r.store.ChangeFiles(what, func() error {
		overlay, err := r.overlay()
		if err != nil {
			return err
		}
		dest := layers.Lookup(overlay, rel)
		var old, from []byte
		if dest != "" {
			info, err := os.Stat(dest)
			if err != nil {
				return err
			}
			if info.ModTime().UnixMilli() != modified {
				return apperr.New(apperr.Conflict, "Your copy changed after you opened it. Reload to see the latest, then make your change again.")
			}
			if old, err = os.ReadFile(dest); err != nil {
				return err
			}
			from = old
		} else {
			if modified != 0 {
				return apperr.New(apperr.Conflict, "Your copy was removed after you opened it. Reload to see the latest.")
			}
			dest = filepath.Join(overlay, filepath.FromSlash(rel))
			for _, l := range r.library.Layers()[1:] {
				if p := layers.Lookup(l.Root, rel); p != "" {
					from, _ = os.ReadFile(p)
					break
				}
			}
		}
		text, err := build(from)
		if err != nil {
			return err
		}
		if from != nil {
			text = textfile.MatchLineEndings(text, string(from))
		}
		if old != nil {
			if err := r.backup(overlay, dest, old); err != nil {
				return err
			}
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		return textfile.WriteAtomic(dest, []byte(text), 0o644)
	})
}

// backup keeps the player's copy before it's changed or removed, under
// <backups>/overlay/<path>.
func (r *Records) backup(overlay, path string, old []byte) error {
	rel, err := filepath.Rel(overlay, path)
	if err != nil {
		return err
	}
	if _, err := textfile.Backup(filepath.Join(r.store.BackupDir(), "overlay", rel), old, 0o644); err != nil {
		return apperr.New(apperr.Conflict, "Could not back up your copy, so nothing was changed: %v", err)
	}
	return nil
}

var yamlLineRe = regexp.MustCompile(`^yaml: line (\d+): `)

func yamlError(err error) string {
	return yamlLineRe.ReplaceAllString(err.Error(), "line $1: ")
}

// checkRecord refuses a record SkyrimNet would reject or misread: invalid
// YAML, an identity field that doesn't match the file's name, or a missing
// required field.
func checkRecord(rel, text string) error {
	var m map[string]any
	if err := yaml.Unmarshal([]byte(text), &m); err != nil {
		return apperr.New(apperr.Invalid, "Not valid YAML: %s", yamlError(err))
	}
	if m == nil {
		return apperr.New(apperr.Invalid, "A record is a set of fields")
	}
	root, _, _ := strings.Cut(rel, "/")
	stem, _, _ := strings.Cut(filepath.Base(rel), ".")
	str := func(k string) string { s, _ := m[k].(string); return strings.TrimSpace(s) }
	need := func(fields ...string) error {
		for _, f := range fields {
			if str(f) == "" {
				return apperr.New(apperr.Invalid, "%s needs a %s", filepath.Base(rel), f)
			}
		}
		return nil
	}
	ident := func(field string) error {
		if err := need(field); err != nil {
			return err
		}
		if !strings.EqualFold(str(field), stem) {
			return apperr.New(apperr.Invalid, "The %s must match the file's name, %q; SkyrimNet skips the file otherwise", field, stem)
		}
		return nil
	}
	kind := str("kind")
	switch root {
	case "triggers":
		if err := ident("name"); err != nil {
			return err
		}
		resp, _ := m["response"].(map[string]any)
		if resp == nil || resp["type"] == nil || fmt.Sprint(resp["type"]) == "" {
			return apperr.New(apperr.Invalid, "A trigger needs a response type")
		}
		if m["eventCriteria"] == nil {
			return apperr.New(apperr.Invalid, "A trigger needs an event to respond to")
		}
	case "actions":
		if err := ident("name"); err != nil {
			return err
		}
		params, _ := m["parameterMapping"].([]any)
		for i, p := range params {
			pm, _ := p.(map[string]any)
			if t := fmt.Sprint(pm["type"]); pm == nil || (t != "speaker" && t != "dynamic" && t != "static") {
				return apperr.New(apperr.Invalid, "Parameter %d's type must be speaker, dynamic, or static", i+1)
			}
		}
	case "entities":
		return need("entityName")
	case "voice_effects":
		if err := ident("id"); err != nil {
			return err
		}
		raw, _ := m["stages"].([]any)
		var stages []Stage
		for i, st := range raw {
			sm, _ := st.(map[string]any)
			if sm == nil {
				return apperr.New(apperr.Invalid, "Stage %d isn't a set of fields", i+1)
			}
			params, _ := sm["params"].(map[string]any)
			stages = append(stages, Stage{Type: fmt.Sprint(sm["type"]), Params: params})
			if sm["type"] == nil {
				stages[i].Type = ""
			}
		}
		if err := checkStages(stages); err != nil {
			return err
		}
	case "filters":
		switch kind {
		case "dialogue_rule", "tts_rule":
			if err := ident("id"); err != nil {
				return err
			}
			if p := str("pattern"); p == "" || len(p) > 1024 {
				return apperr.New(apperr.Invalid, "A rule needs a pattern of at most 1024 characters")
			}
		case "actor", "memory":
			for _, k := range []string{"FactionWhitelist", "FactionBlacklist", "RaceWhitelist", "RaceBlacklist", "GenderWhitelist", "GenderBlacklist"} {
				if v, ok := m[k]; ok && v != nil {
					if _, isList := v.([]any); !isList {
						return apperr.New(apperr.Invalid, "%s must be a list", k)
					}
				}
			}
		default:
			return apperr.New(apperr.Invalid, "A filter's kind must be actor, memory, dialogue_rule, or tts_rule")
		}
	case "dialogue_actions":
		switch kind {
		case "instruction":
			if err := ident("key"); err != nil {
				return err
			}
			return need("text")
		case "lists":
		default:
			return apperr.New(apperr.Invalid, "A dialogue action's kind must be lists or instruction")
		}
	}
	return nil
}
