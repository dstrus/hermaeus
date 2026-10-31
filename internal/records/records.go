// Package records reads SkyrimNet's actions and triggers: the YAML files in
// each content layer's actions/ and triggers/ folders. A file in a higher
// layer replaces the one with the same name below it, as the player's own
// edits (overlay/) replace a plugin's, and a plugin's replace SkyrimNet's.
// SkyrimNet also has a few actions built in, which aren't files.
package records

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"hermaeus/internal/layers"
	"hermaeus/internal/store"

	"go.yaml.in/yaml/v3"
)

type Records struct {
	store   *store.Store
	library *layers.Library
}

func New(s *store.Store, l *layers.Library) *Records {
	return &Records{store: s, library: l}
}

// Origin is where a record comes from.
type Origin struct {
	ID        string `json:"id"` // the file's name, without .yaml
	Layer     string `json:"layer"`
	LayerKind string `json:"layerKind"` // overlay, pack, or base
	Source    string `json:"source"`    // the layer, for people
	Path      string `json:"path"`      // under the layer's folder
	Replaces  string `json:"replaces,omitempty"`
}

// Param is one of an action's parameters, which the model fills in.
type Param struct {
	Name        string `json:"name"`
	Type        string `json:"type"` // speaker, target, dynamic…
	Description string `json:"description,omitempty"`
	Optional    bool   `json:"optional,omitempty"`
}

// Action is something an NPC can do when the model picks it.
type Action struct {
	Origin      `tstype:",extends"` // its fields sent flat, as JSON embeds them
	Name        string              `json:"name"`
	Description string              `json:"description"` // a template, read by the model
	Category    string              `json:"category,omitempty"`
	Enabled     bool                `json:"enabled"`            // its switch in content-registry.json, or else its file
	Override    string              `json:"override,omitempty"` // that switch, "enabled" or "disabled", when it has one
	Priority    *int                `json:"priority,omitempty"`
	Params      []Param             `json:"params"`
	Rules       int                 `json:"rules"` // eligibility rules: who may take it, and when
	Quest       string              `json:"quest,omitempty"`
	Script      string              `json:"script,omitempty"`
	Function    string              `json:"function,omitempty"`
	Event       string              `json:"event,omitempty"` // the event text recorded when it runs
	Confirm     bool                `json:"confirm,omitempty"`
}

// Condition is one test on an event's fields.
type Condition struct {
	Field    string `json:"field"`
	Operator string `json:"operator"`
	Value    string `json:"value"`
}

// Trigger has SkyrimNet respond when an event happens.
type Trigger struct {
	Origin       `tstype:",extends"` // its fields sent flat, as JSON embeds them
	Name         string              `json:"name"`
	Description  string              `json:"description"`
	Enabled      bool                `json:"enabled"`            // its switch in content-registry.json, or else its file
	Override     string              `json:"override,omitempty"` // that switch, when it has one
	EventType    string              `json:"eventType"`
	Conditions   []Condition         `json:"conditions"`
	Audience     string              `json:"audience,omitempty"`
	ResponseType string              `json:"responseType,omitempty"`
	Content      string              `json:"content,omitempty"` // a template
	TargetScope  string              `json:"targetScope,omitempty"`
	Cooldown     *float64            `json:"cooldown,omitempty"` // seconds
	Priority     *int                `json:"priority,omitempty"`
	Probability  *float64            `json:"probability,omitempty"`
	Interrupt    bool                `json:"interrupt,omitempty"`
	EffectID     string              `json:"effectId,omitempty"` // the voice effect an activate_voice_effect response switches on
}

type ActionList struct {
	Actions []Action `json:"actions"` // by layer, highest last as the dashboard shows plugins, then by name
}

type TriggerList struct {
	Triggers []Trigger `json:"triggers"`
}

// files reads every record in a folder across the layers, the highest
// layer's for each name, noting the layer below it that it replaces.
func (r *Records) files(folder string, read func(o Origin, b []byte)) {
	type found struct {
		o Origin
		b []byte
	}
	var all []found
	seen := map[string]int{} // folded id → index into all
	for _, l := range r.library.Layers() {
		dir := layers.Lookup(l.Root, folder)
		if dir == "" {
			continue
		}
		filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() || filepath.Ext(path) != ".yaml" { // SkyrimNet skips .YAML
				return nil
			}
			id := strings.TrimSuffix(d.Name(), filepath.Ext(d.Name()))
			key := strings.ToLower(id)
			if i, ok := seen[key]; ok {
				if all[i].o.Replaces == "" {
					all[i].o.Replaces = l.Source
				}
				return nil
			}
			b, err := os.ReadFile(path)
			if err != nil {
				return nil
			}
			rel, _ := filepath.Rel(l.Root, path)
			seen[key] = len(all)
			all = append(all, found{Origin{ID: id, Layer: l.ID, LayerKind: l.Kind, Source: l.Source, Path: filepath.ToSlash(rel)}, b})
			return nil
		})
	}
	for _, f := range all {
		read(f.o, f.b)
	}
}

func str(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case nil:
		return ""
	}
	b, _ := yaml.Marshal(v)
	return strings.TrimSpace(string(b))
}

func intPtr(v any) *int {
	switch x := v.(type) {
	case int:
		return &x
	case float64:
		i := int(x)
		return &i
	}
	return nil
}

func floatPtr(v any) *float64 {
	switch x := v.(type) {
	case int:
		f := float64(x)
		return &f
	case float64:
		return &x
	}
	return nil
}

// on reads an enabled flag; a record without one is on.
func on(v any) bool {
	b, ok := v.(bool)
	return !ok || b
}

// Actions lists every action in a file.
func (r *Records) Actions() ActionList {
	out := ActionList{Actions: []Action{}}
	acts := r.library.Activations()
	r.files("actions", func(o Origin, b []byte) {
		var m map[string]any
		if yaml.Unmarshal(b, &m) != nil || m == nil {
			return
		}
		a := Action{Origin: o, Name: str(m["name"]), Description: str(m["description"]), Category: str(m["customCategory"]), Override: acts[strings.ToLower(o.Path)],
			Priority: intPtr(m["defaultPriority"]), Params: []Param{}, Quest: str(m["questEditorId"]), Script: str(m["scriptName"]), Function: str(m["executionFunctionName"]),
			Event: str(m["eventString"])}
		a.Enabled = layers.Activation(a.Override, on(m["enabled"]))
		if a.Name == "" {
			a.Name = o.ID
		}
		if c, ok := m["useConfirmationWindow"].(bool); ok {
			a.Confirm = c
		}
		if rules, ok := m["eligibilityRules"].([]any); ok {
			a.Rules = len(rules)
		}
		if ps, ok := m["parameterMapping"].([]any); ok {
			for _, p := range ps {
				if pm, ok := p.(map[string]any); ok {
					opt, _ := pm["optional"].(bool)
					a.Params = append(a.Params, Param{Name: str(pm["name"]), Type: str(pm["type"]), Description: str(pm["description"]), Optional: opt})
				}
			}
		}
		out.Actions = append(out.Actions, a)
	})
	rank := r.order()
	sort.SliceStable(out.Actions, func(i, j int) bool {
		return byLayer(rank, out.Actions[i].Origin, out.Actions[i].Name, out.Actions[j].Origin, out.Actions[j].Name)
	})
	return out
}

// Triggers lists every trigger.
func (r *Records) Triggers() TriggerList {
	out := TriggerList{Triggers: []Trigger{}}
	acts := r.library.Activations()
	r.files("triggers", func(o Origin, b []byte) {
		var m map[string]any
		if yaml.Unmarshal(b, &m) != nil || m == nil {
			return
		}
		t := Trigger{Origin: o, Name: str(m["name"]), Description: str(m["description"]), Override: acts[strings.ToLower(o.Path)], Conditions: []Condition{},
			Audience: str(m["audience"]), Cooldown: floatPtr(m["cooldownSeconds"]), Priority: intPtr(m["priority"]), Probability: floatPtr(m["probability"])}
		t.Enabled = layers.Activation(t.Override, on(m["enabled"]))
		if t.Name == "" {
			t.Name = o.ID
		}
		if i, ok := m["interrupt"].(bool); ok {
			t.Interrupt = i
		}
		if ec, ok := m["eventCriteria"].(map[string]any); ok {
			t.EventType = str(ec["eventType"])
			if cs, ok := ec["schemaConditions"].([]any); ok {
				for _, c := range cs {
					if cm, ok := c.(map[string]any); ok {
						t.Conditions = append(t.Conditions, Condition{Field: str(cm["fieldPath"]), Operator: str(cm["operator"]), Value: str(cm["value"])})
					}
				}
			}
		}
		if rs, ok := m["response"].(map[string]any); ok {
			t.ResponseType, t.Content, t.TargetScope = str(rs["type"]), str(rs["content"]), str(rs["targetScope"])
		}
		t.EffectID = effectIDOf(m)
		out.Triggers = append(out.Triggers, t)
	})
	rank := r.order()
	sort.SliceStable(out.Triggers, func(i, j int) bool {
		return byLayer(rank, out.Triggers[i].Origin, out.Triggers[i].Name, out.Triggers[j].Origin, out.Triggers[j].Name)
	})
	return out
}

// order ranks layers as the Plugins page shows them, lowest priority first.
func (r *Records) order() map[string]int {
	lay := r.library.Layers()
	rank := map[string]int{}
	for i, l := range lay {
		rank[l.ID] = len(lay) - i
	}
	return rank
}

// byLayer orders records by their layer's rank, then by name.
func byLayer(rank map[string]int, a Origin, an string, b Origin, bn string) bool {
	if rank[a.Layer] != rank[b.Layer] {
		return rank[a.Layer] < rank[b.Layer]
	}
	return strings.ToLower(an) < strings.ToLower(bn)
}
