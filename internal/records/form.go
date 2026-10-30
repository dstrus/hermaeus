package records

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"strings"

	"hermaeus/internal/apperr"
	"hermaeus/internal/layers"
	"hermaeus/internal/settings"
	"hermaeus/internal/textfile"

	"go.yaml.in/yaml/v3"
)

// A record's form edits the record as a JSON object, its fields in the
// file's order. Saving merges the object back into the YAML it came from,
// changing only what differs: fields keep their order, values that didn't
// change keep how they were written, and comments stay where they were.

// RecordForm is a record for its form: its copies by layer, and the one in
// use as an object, or why it can't be read as one.
type RecordForm struct {
	Versions []Version      `json:"versions"`
	Data     map[string]any `json:"data" tstype:"{ [key: string]: unknown } | null"` // nil when it isn't a set of fields
	Error    string         `json:"error,omitempty"`
}

// FormEdit is a record as its form leaves it.
type FormEdit struct {
	Modified int64           `json:"modified"` // the player's copy's time when read, 0 for none yet
	Data     json.RawMessage `json:"data" tstype:"{ [key: string]: unknown }"`
	New      bool            `json:"new,omitempty"` // a new record: refused if any layer has the file
}

// Form reads a record for its form.
func (r *Records) Form(rel string) (RecordForm, error) {
	rel, err := recordPath(rel)
	if err != nil {
		return RecordForm{}, err
	}
	v, err := r.Versions(rel)
	if err != nil {
		return RecordForm{}, err
	}
	return formOf(v), nil
}

// formOf decodes the copy in use. The object is sent as ordered JSON (see
// MarshalJSON on orderedForm), so its fields keep the file's order.
func formOf(v []Version) RecordForm {
	out := RecordForm{Versions: v}
	if len(v) == 0 || v[0].Text == nil {
		return out
	}
	var doc yaml.Node
	if err := yaml.Unmarshal([]byte(*v[0].Text), &doc); err != nil {
		out.Error = yamlError(err)
		return out
	}
	if len(doc.Content) == 0 || doc.Content[0].Kind != yaml.MappingNode {
		out.Error = "It isn’t a set of fields"
		return out
	}
	var m map[string]any
	if err := doc.Content[0].Decode(&m); err != nil {
		out.Error = err.Error()
		return out
	}
	out.Data = m
	return out
}

// MarshalJSON sends the form with its object's fields in the file's order,
// which a Go map loses.
func (f RecordForm) MarshalJSON() ([]byte, error) {
	var data []byte
	if f.Data == nil || len(f.Versions) == 0 || f.Versions[0].Text == nil {
		data = []byte("null")
	} else {
		var doc yaml.Node
		if err := yaml.Unmarshal([]byte(*f.Versions[0].Text), &doc); err != nil || len(doc.Content) == 0 {
			data = []byte("null")
		} else {
			var buf bytes.Buffer
			if err := nodeToJSON(&buf, doc.Content[0]); err != nil {
				return nil, err
			}
			data = buf.Bytes()
		}
	}
	type plain struct {
		Versions []Version       `json:"versions"`
		Data     json.RawMessage `json:"data"`
		Error    string          `json:"error,omitempty"`
	}
	return json.Marshal(plain{f.Versions, data, f.Error})
}

// SaveForm saves a record's form as the player's copy, or as a new record.
func (r *Records) SaveForm(rel string, e FormEdit) (RecordForm, error) {
	rel, err := recordPath(rel)
	if err != nil {
		return RecordForm{}, err
	}
	values, err := jsonToNode(e.Data)
	if err != nil {
		return RecordForm{}, apperr.New(apperr.Invalid, "Invalid record: %v", err)
	}
	if values.Kind != yaml.MappingNode {
		return RecordForm{}, apperr.New(apperr.Invalid, "A record is a set of fields")
	}
	if e.New {
		for _, l := range r.layerList() {
			if layers.Lookup(l.Root, rel) != "" {
				return RecordForm{}, apperr.New(apperr.Conflict, "%s already exists, in %s. Choose another name, or edit that one.", rel, l.Source)
			}
		}
	}
	err = r.saveCopy(rel, e.Modified, "Saved your copy of "+rel, func(from []byte) (string, error) {
		fresh := len(from) == 0
		out, err := settings.EditYAML(from, func(doc *yaml.Node) error {
			if doc.Kind != yaml.MappingNode {
				return apperr.New(apperr.Invalid, "The version being replaced isn’t a set of fields, so it can’t be edited here")
			}
			mergeNode(doc, values)
			return nil
		})
		if err != nil {
			if _, ok := apperr.KindOf(err); ok {
				return "", err
			}
			return "", apperr.New(apperr.Invalid, "The version being replaced isn’t valid YAML, so it can’t be edited here: %v", err)
		}
		text := textfile.Normalize(string(out))
		if fresh && !strings.HasSuffix(text, "\n") {
			text += "\n"
		}
		return text, checkRecord(rel, text)
	})
	if err != nil {
		return RecordForm{}, err
	}
	v, err := r.Versions(rel)
	if err != nil {
		return RecordForm{}, err
	}
	return formOf(v), nil
}

// ---------- YAML and JSON, keeping order ----------

// nodeToJSON writes a YAML node as JSON, keeping mapping keys in order.
func nodeToJSON(buf *bytes.Buffer, n *yaml.Node) error {
	switch n.Kind {
	case yaml.AliasNode:
		return nodeToJSON(buf, n.Alias)
	case yaml.MappingNode:
		buf.WriteByte('{')
		for i := 0; i+1 < len(n.Content); i += 2 {
			if i > 0 {
				buf.WriteByte(',')
			}
			k, _ := json.Marshal(n.Content[i].Value)
			buf.Write(k)
			buf.WriteByte(':')
			if err := nodeToJSON(buf, n.Content[i+1]); err != nil {
				return err
			}
		}
		buf.WriteByte('}')
	case yaml.SequenceNode:
		buf.WriteByte('[')
		for i, c := range n.Content {
			if i > 0 {
				buf.WriteByte(',')
			}
			if err := nodeToJSON(buf, c); err != nil {
				return err
			}
		}
		buf.WriteByte(']')
	default:
		var v any
		if err := n.Decode(&v); err != nil {
			return err
		}
		if f, ok := v.(float64); ok && (math.IsNaN(f) || math.IsInf(f, 0)) {
			v = nil
		}
		b, err := json.Marshal(v)
		if err != nil {
			return err
		}
		buf.Write(b)
	}
	return nil
}

// jsonToNode reads JSON into a YAML node, keeping object keys in order.
func jsonToNode(data []byte) (*yaml.Node, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	n, err := readJSONValue(dec)
	if err != nil {
		return nil, err
	}
	if _, err := dec.Token(); err != io.EOF {
		return nil, fmt.Errorf("unexpected data after the record")
	}
	return n, nil
}

func readJSONValue(dec *json.Decoder) (*yaml.Node, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	switch v := tok.(type) {
	case json.Delim:
		if v == '{' {
			m := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
			for dec.More() {
				k, err := dec.Token()
				if err != nil {
					return nil, err
				}
				val, err := readJSONValue(dec)
				if err != nil {
					return nil, err
				}
				m.Content = append(m.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: k.(string)}, val)
			}
			_, err := dec.Token()
			return m, err
		}
		seq := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
		for dec.More() {
			val, err := readJSONValue(dec)
			if err != nil {
				return nil, err
			}
			seq.Content = append(seq.Content, val)
		}
		_, err := dec.Token()
		return seq, err
	case string:
		n := &yaml.Node{}
		n.SetString(v)
		return n, nil
	case json.Number:
		tag := "!!int"
		if strings.ContainsAny(string(v), ".eE") {
			tag = "!!float"
		}
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: tag, Value: string(v)}, nil
	case bool:
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: fmt.Sprint(v)}, nil
	default:
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!null", Value: "null"}, nil
	}
}

// mergeNode makes old hold new's values while keeping what it can of old:
// fields in their order, scalars that didn't change as they were written,
// comments where they were, and list items merged one by one. Fields only
// in new go at the end; fields only in old are removed.
func mergeNode(old, new *yaml.Node) {
	switch {
	case old.Kind == yaml.MappingNode && new.Kind == yaml.MappingNode:
		want := map[string]*yaml.Node{}
		var order []string
		for i := 0; i+1 < len(new.Content); i += 2 {
			want[new.Content[i].Value] = new.Content[i+1]
			order = append(order, new.Content[i].Value)
		}
		var kept []*yaml.Node
		seen := map[string]bool{}
		for i := 0; i+1 < len(old.Content); i += 2 {
			k, v := old.Content[i], old.Content[i+1]
			if nv, ok := want[k.Value]; ok && !seen[k.Value] {
				seen[k.Value] = true
				mergeNode(v, nv)
				kept = append(kept, k, v)
			}
		}
		for _, k := range order {
			if !seen[k] {
				kept = append(kept, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: k}, want[k])
			}
		}
		old.Content = kept
	case old.Kind == yaml.SequenceNode && new.Kind == yaml.SequenceNode:
		for i, nv := range new.Content {
			if i < len(old.Content) {
				mergeNode(old.Content[i], nv)
			} else {
				old.Content = append(old.Content, nv)
			}
		}
		old.Content = old.Content[:len(new.Content)]
	case old.Kind == yaml.ScalarNode && new.Kind == yaml.ScalarNode && sameScalar(old, new):
		// unchanged: keep how it was written
	default:
		head, line, foot := old.HeadComment, old.LineComment, old.FootComment
		// A changed text keeps its quoting, or its block style for several
		// lines, as SkyrimNet's prompts and descriptions are often written.
		style := new.Style
		if old.Kind == yaml.ScalarNode && new.Kind == yaml.ScalarNode && old.Tag == new.Tag {
			oldBlock := old.Style&(yaml.LiteralStyle|yaml.FoldedStyle) != 0
			if oldBlock == strings.Contains(new.Value, "\n") {
				style = old.Style
			}
		}
		*old = *new
		old.Style = style
		old.HeadComment, old.LineComment, old.FootComment = head, line, foot
	}
}

// sameScalar compares two scalars by value: 2 and 2.0 are the same number.
func sameScalar(a, b *yaml.Node) bool {
	var x, y any
	if a.Decode(&x) != nil || b.Decode(&y) != nil {
		return false
	}
	if fx, ok := toFloat(x); ok {
		fy, ok := toFloat(y)
		return ok && fx == fy
	}
	return fmt.Sprint(x) == fmt.Sprint(y) && (x == nil) == (y == nil)
}
