package settings

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"hermaeus/internal/apperr"
	"hermaeus/internal/textfile"

	"go.yaml.in/yaml/v3"
)

// Edit changes some settings in one file. Modified is the file's time when
// the page loaded it, so a change made since (by SkyrimNet, say) isn't
// overwritten.
type Edit struct {
	Modified int64    `json:"modified"`
	Changes  []Change `json:"changes"`
}

type Change struct {
	Key   string `json:"key"`
	Value any    `json:"value,omitempty"`
	Reset bool   `json:"reset,omitempty"` // back to the recorded default instead
}

// Save applies an edit to one settings file, after backing the file up.
func (s *Settings) Save(name string, e Edit) (FileValues, error) {
	if !validName(name) {
		return FileValues{}, apperr.New(apperr.Invalid, "Invalid settings name %q", name)
	}
	if len(e.Changes) == 0 {
		return FileValues{}, apperr.New(apperr.Invalid, "Nothing to change")
	}
	keys := make([]string, len(e.Changes))
	for i, c := range e.Changes {
		keys[i] = c.Key
	}
	var out FileValues
	err := s.store.ChangeFiles(fmt.Sprintf("Saved %s settings: %s", name, strings.Join(keys, ", ")), func() error { // names only, never values
		dir, err := s.configDir()
		if err != nil {
			return err
		}
		path := filePath(dir, name)
		info, err := os.Stat(path)
		if os.IsNotExist(err) {
			return apperr.New(apperr.NotFound, "%s doesn't exist yet. SkyrimNet creates it the next time it starts.", filepath.Base(path))
		}
		if err != nil {
			return err
		}
		if info.ModTime().UnixMilli() != e.Modified {
			return apperr.New(apperr.Conflict, "These settings changed after you opened them. Reload to see the latest, then make your changes again.")
		}
		old, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		defs, _ := s.loadDefaults(dir).of(name)
		updated, err := apply(old, name, defs, e.Changes)
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(s.store.Dir(), path)
		if _, err := textfile.Backup(filepath.Join(s.store.BackupDir(), "settings", rel), old, 0o600); err != nil { // may hold API keys
			return fmt.Errorf("could not back up the settings, so nothing was changed: %w", err)
		}
		if err := textfile.WriteAtomic(path, updated, info.Mode().Perm()); err != nil {
			return err
		}
		out, err = read(dir, name, s.loadDefaults(dir))
		out.File.Pinned = s.pinned(name)
		return err
	})
	return out, err
}

// apply returns the file with the changes made, the way it was written
// (see yaml.go).
func apply(old []byte, name string, defs map[string]any, changes []Change) ([]byte, error) {
	return rewrite(old, name, func(doc *yaml.Node) error {
		for _, c := range changes {
			if err := applyChange(doc, name, defs, c); err != nil {
				return err
			}
		}
		return nil
	})
}

// EditYAML changes a YAML file's document and returns the file written the
// way it was: values swapped in place where only values changed, otherwise
// the changed lines merged in, so comments and formatting stay.
func EditYAML(old []byte, change func(doc *yaml.Node) error) ([]byte, error) {
	return rewrite(old, "the file", change)
}

// rewrite returns the file with change made to its document, written the
// way it was: values swapped in place where only values changed, otherwise
// the changed lines merged in (see yaml.go).
func rewrite(old []byte, name string, change func(doc *yaml.Node) error) ([]byte, error) {
	untouched, err := parse(old)
	if err != nil {
		return nil, apperr.New(apperr.Invalid, "Could not read %s: %v", name, err)
	}
	edited, _ := parse(old)
	if err := change(edited.doc); err != nil {
		return nil, err
	}
	if patched, ok := patchValues(string(old), untouched.doc, edited.doc); ok {
		return []byte(patched), nil
	}
	before, err := untouched.encode()
	if err != nil {
		return nil, err
	}
	after, err := edited.encode()
	if err != nil {
		return nil, err
	}
	return []byte(keepFormatting(string(old), string(before), string(after))), nil
}

func applyChange(doc *yaml.Node, file string, defs map[string]any, c Change) error {
	node := findNode(doc, c.Key)
	def, hasDef := defs[c.Key]
	if node == nil && !hasDef {
		return apperr.New(apperr.Invalid, "There's no setting %s", c.Key)
	}
	var current any
	if node != nil {
		node.Decode(&current)
	}
	if containsSecret(current) || containsSecret(def) {
		return apperr.New(apperr.Invalid, "%s holds API keys, so it can't be edited here. Use SkyrimNet's dashboard.", c.Key)
	}
	value := c.Value
	if c.Reset {
		if !hasDef {
			return apperr.New(apperr.Invalid, "%s has no recorded default", c.Key)
		}
		value = cleanDefault(def)
	}
	like := current
	if node == nil || current == nil {
		like = def
	}
	value, err := matchKind(c.Key, value, like)
	if err != nil {
		return err
	}
	if isSecret(file, c.Key) {
		if _, ok := value.(string); !ok && value != nil {
			return apperr.New(apperr.Invalid, "%s must be text", c.Key)
		}
	}
	var fresh yaml.Node
	if err := fresh.Encode(value); err != nil {
		return err
	}
	if node == nil {
		parent := doc
		parts := strings.Split(c.Key, ".")
		for _, part := range parts[:len(parts)-1] {
			next := findNode(parent, part)
			if next == nil {
				next = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
				parent.Content = append(parent.Content, scalar(part), next)
			}
			if next.Kind != yaml.MappingNode {
				return apperr.New(apperr.Invalid, "Can't add %s: %s isn't a group of settings", c.Key, part)
			}
			parent = next
		}
		parent.Content = append(parent.Content, scalar(parts[len(parts)-1]), &fresh)
		return nil
	}
	// Keep the keys of groups in the order the file had them: a value from
	// the page arrives as JSON, whose order Go doesn't keep.
	orderLike(&fresh, node)
	// Keep a quoted string quoted, and comments where they were.
	if node.Kind == yaml.ScalarNode && fresh.Kind == yaml.ScalarNode && node.Tag == fresh.Tag {
		fresh.Style = node.Style
	}
	fresh.HeadComment, fresh.LineComment, fresh.FootComment = node.HeadComment, node.LineComment, node.FootComment
	*node = fresh
	return nil
}

// cleanDefault undoes the single-precision rounding in recorded defaults, so
// resetting writes 0.7, not 0.699999988079071.
func cleanDefault(v any) any {
	switch x := v.(type) {
	case float64:
		if x == math.Trunc(x) {
			return x
		}
		f, _ := strconv.ParseFloat(strconv.FormatFloat(float64(float32(x)), 'g', -1, 32), 64)
		return f
	case []any:
		out := make([]any, len(x))
		for i, e := range x {
			out[i] = cleanDefault(e)
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, e := range x {
			out[k] = cleanDefault(e)
		}
		return out
	}
	return v
}

// matchKind converts a value from the page to the kind of setting it
// replaces, refusing a different kind: SkyrimNet expects each setting's
// type. JSON numbers arrive as float64, so a whole one replacing an integer
// setting becomes an integer.
func matchKind(key string, v, like any) (any, error) {
	kind := func(x any) string {
		switch x.(type) {
		case nil:
			return ""
		case bool:
			return "a true/false switch"
		case int, int64, uint64, float64:
			return "a number"
		case string:
			return "text"
		case []any:
			return "a list"
		case map[string]any:
			return "a group of settings"
		}
		return fmt.Sprintf("%T", x)
	}
	if like == nil || v == nil {
		return v, nil
	}
	if kind(v) != kind(like) {
		return nil, apperr.New(apperr.Invalid, "%s must be %s", key, kind(like))
	}
	if f, ok := v.(float64); ok {
		if _, wholeOnly := like.(int); wholeOnly {
			if f != math.Trunc(f) {
				return nil, apperr.New(apperr.Invalid, "%s must be a whole number", key)
			}
			return int64(f), nil
		}
	}
	return v, nil
}

// orderLike puts the keys of each group in fresh in the order they have in
// old, the value it replaces; keys old doesn't have follow, as they were.
// A list's entries follow old's entry at the same place, or its first.
func orderLike(fresh, old *yaml.Node) {
	if fresh == nil || old == nil {
		return
	}
	switch {
	case fresh.Kind == yaml.MappingNode && old.Kind == yaml.MappingNode:
		at := map[string]int{}
		for i := 0; i+1 < len(old.Content); i += 2 {
			at[old.Content[i].Value] = i / 2
		}
		type pair struct{ k, v *yaml.Node }
		pairs := make([]pair, 0, len(fresh.Content)/2)
		for i := 0; i+1 < len(fresh.Content); i += 2 {
			pairs = append(pairs, pair{fresh.Content[i], fresh.Content[i+1]})
		}
		rank := func(p pair) int {
			if n, ok := at[p.k.Value]; ok {
				return n
			}
			return len(at)
		}
		sort.SliceStable(pairs, func(i, j int) bool { return rank(pairs[i]) < rank(pairs[j]) })
		fresh.Content = fresh.Content[:0]
		for _, p := range pairs {
			if n, ok := at[p.k.Value]; ok {
				orderLike(p.v, old.Content[n*2+1])
			}
			fresh.Content = append(fresh.Content, p.k, p.v)
		}
	case fresh.Kind == yaml.SequenceNode && old.Kind == yaml.SequenceNode && len(old.Content) > 0:
		for i, e := range fresh.Content {
			like := old.Content[0]
			if i < len(old.Content) {
				like = old.Content[i]
			}
			orderLike(e, like)
		}
	}
}
