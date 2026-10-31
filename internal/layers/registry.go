package layers

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// A content record can be switched on or off in content-registry.json's
// items, by its path ("triggers/conduit_weapon_cast_narration.yaml"), without
// touching its file: "enabled" or "disabled" wins over the file's own
// enabled:, and a record with no entry follows its file. SkyrimNet's
// dashboard switches records this way, and keeps an entry only while it
// differs from the file.
const (
	Enabled  = "enabled"
	Disabled = "disabled"
)

// Activations reads the registry's switched records, by their folded path.
// A registry that can't be read switches nothing.
func (l *Library) Activations() map[string]string {
	out := map[string]string{}
	sky := l.skyrimNetDir()
	if sky == "" {
		return out
	}
	b, err := os.ReadFile(filepath.Join(sky, registryFile))
	if err != nil {
		return out
	}
	var reg struct {
		Items map[string]any `json:"items"`
	}
	if json.Unmarshal(b, &reg) != nil {
		return out
	}
	for path, v := range reg.Items {
		if word := activationWord(v); word != "" {
			out[strings.ToLower(path)] = word
		}
	}
	return out
}

// activationWord reads an item's value: SkyrimNet's word, or a plain true
// or false.
func activationWord(v any) string {
	switch x := v.(type) {
	case string:
		if x == Enabled || x == Disabled {
			return x
		}
	case bool:
		if x {
			return Enabled
		}
		return Disabled
	}
	return ""
}

// Activation is whether a record is on: its registry entry, or else its
// file's own enabled.
func Activation(word string, fileOn bool) bool {
	switch word {
	case Enabled:
		return true
	case Disabled:
		return false
	}
	return fileOn
}

// EncodeRegistry writes the registry as SkyrimNet does: sorted keys and
// two-space indents, ending as the old file did.
func EncodeRegistry(raw map[string]any, old []byte) ([]byte, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(raw); err != nil {
		return nil, err
	}
	out := buf.Bytes()
	if !bytes.HasSuffix(old, []byte("\n")) {
		out = bytes.TrimRight(out, "\n")
	}
	return out, nil
}

// SetActivation sets a record's entry in the registry's items: the word, or
// none for "" (the record follows its file). Any entry for the same path in
// another case is replaced.
func SetActivation(raw map[string]any, path, word string) {
	items, _ := raw["items"].(map[string]any)
	if items == nil {
		items = map[string]any{}
	}
	for k := range items {
		if strings.EqualFold(k, path) {
			delete(items, k)
		}
	}
	if word != "" {
		items[path] = word
	}
	raw["items"] = items
}
