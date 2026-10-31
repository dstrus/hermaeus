package records

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"hermaeus/internal/apperr"
	"hermaeus/internal/layers"
	"hermaeus/internal/settings"

	"go.yaml.in/yaml/v3"
)

// Voice effects are recipes that change how a voice sounds, such as a
// draugr's or a ghost's: a chain of stages applied to speech in order, and
// optionally a sample that replaces the speaker's own for cloning. Each is
// a record in voice_effects/, named by its id. A recipe doesn't say whom
// it's for: a trigger whose response is activate_voice_effect puts it on
// the trigger's target, a mod's script can with SkyrimNet's SetVoiceEffect,
// and SkyrimNet itself uses narrator_effect and player_thought_effect. The
// assignments live in SkyrimNet's save data, not in these files.

// StageParam is one setting of a kind of stage.
type StageParam struct {
	Name    string   `json:"name"` // as the recipe has it
	Label   string   `json:"label"`
	Unit    string   `json:"unit,omitempty"`
	Whole   bool     `json:"whole,omitempty"`   // written as a whole number, as SkyrimNet's recipes write it
	Suggest []string `json:"suggest,omitempty"` // text, with these values seen in SkyrimNet's recipes and dashboard
	Default any      `json:"default"`
}

// StageKind is a kind of stage, with its settings. The types, their
// settings, and their defaults are those SkyrimNet's dashboard offers when
// a stage is added; the limits SkyrimNet puts on each setting aren't
// published, so Hermaeus neither shows nor checks them.
type StageKind struct {
	Type   string       `json:"type"`
	Label  string       `json:"label"`
	What   string       `json:"what"`
	Params []StageParam `json:"params"`
}

func p(name, label, unit string, def any) StageParam {
	return StageParam{Name: name, Label: label, Unit: unit, Default: def}
}

func whole(name, label, unit string, def int) StageParam {
	return StageParam{Name: name, Label: label, Unit: unit, Whole: true, Default: def}
}

// StageKinds are the stages SkyrimNet's dashboard offers, in its order.
var StageKinds = []StageKind{
	{"pitch_shift", "Pitch shift", "Raises or lowers the voice.", []StageParam{p("semitones", "Semitones", "semitones", 0.0)}},
	{"formant_shift", "Formant shift", "Changes the size the voice sounds, as from a larger or smaller throat, without changing its pitch.", []StageParam{p("semitones", "Semitones", "semitones", 0.0)}},
	{"speed", "Speed", "Speeds speech up or slows it down. Higher is faster.", []StageParam{p("ratio", "Ratio", "", 1.0)}},
	{"bitcrush", "Bitcrusher", "Coarsens the sound, like old or broken audio.", []StageParam{whole("bits", "Bits", "", 16), whole("rate_div", "Rate divider", "", 1)}},
	{"distortion", "Distortion", "Roughens the voice.", []StageParam{p("drive", "Drive", "", 0.0), p("mix", "Mix", "", 1.0)}},
	{"reverb", "Reverb", "Places the voice in a room, a hall, or a cave.", []StageParam{p("room_size", "Room size", "", 0.5), p("decay", "Decay", "", 0.5), p("wet", "Wet", "", 0.3)}},
	{"chorus", "Chorus", "Layers slightly shifted copies of the voice, as if several speak at once.", []StageParam{p("rate_hz", "Rate", "Hz", 1.5), p("depth_ms", "Depth", "ms", 3.0), whole("voices", "Voices", "", 2)}},
	{"tremolo", "Tremolo", "Makes the voice waver in loudness.", []StageParam{p("rate_hz", "Rate", "Hz", 5.0), p("depth", "Depth", "", 0.5)}},
	{"ringmod", "Ring modulation", "Gives the voice a metallic, inhuman ring.", []StageParam{p("freq_hz", "Frequency", "Hz", 440.0), p("mix", "Mix", "", 1.0)}},
	{"echo", "Echo", "Repeats the voice after a delay.", []StageParam{whole("delay_ms", "Delay", "ms", 200), p("feedback", "Feedback", "", 0.3), p("wet", "Wet", "", 0.5)}},
	{"eq", "Equalizer", "Raises or cuts part of the sound, around a frequency.", []StageParam{
		{Name: "mode", Label: "Mode", Suggest: []string{"band", "low"}, Default: "band"},
		p("freq_hz", "Frequency", "Hz", 1000.0),
		p("q", "Width", "", 1.0),
		p("gain_db", "Gain", "dB", 0.0),
	}},
	{"whisper", "Whisper", "Mixes in breath noise, for a hushed or ghostly voice.", []StageParam{p("noise_db", "Noise", "dB", -20.0), p("mix", "Mix", "", 0.5)}},
	{"gain", "Gain", "Makes the voice louder or quieter.", []StageParam{p("gain_db", "Gain", "dB", 0.0)}},
}

func stageKind(t string) *StageKind {
	for i := range StageKinds {
		if StageKinds[i].Type == t {
			return &StageKinds[i]
		}
	}
	return nil
}

// Stage is one step of a recipe.
type Stage struct {
	Type   string         `json:"type"`
	Params map[string]any `json:"params"`
}

// VoiceEffect is a recipe, as the highest layer that has it says.
type VoiceEffect struct {
	ID        string     `json:"id"`
	Path      string     `json:"path"`
	Name      string     `json:"name"`
	Enabled   bool       `json:"enabled"`
	Override  string     `json:"override,omitempty"` // its switch in content-registry.json, when it has one
	Sample    string     `json:"sample"`             // "" for none
	Stages    []Stage    `json:"stages"`
	Providers []Provider `json:"providers"` // highest priority first
	Error     string     `json:"error,omitempty"`
	// Triggers that switch the effect on or, naming it, off.
	Triggers []EffectTrigger `json:"triggers"`
}

// EffectTrigger is a trigger that names a voice effect.
type EffectTrigger struct {
	Path      string `json:"path"`
	Name      string `json:"name"`
	Source    string `json:"source"`
	Activates bool   `json:"activates"` // or deactivates
	Enabled   bool   `json:"enabled"`
}

type VoiceEffectList struct {
	Layers  []layers.Layer `json:"layers"`
	Effects []VoiceEffect  `json:"effects"` // by name
	Kinds   []StageKind    `json:"kinds"`
}

// VoiceEffects lists every recipe in any layer.
func (r *Records) VoiceEffects() VoiceEffectList {
	c := r.Content()
	out := VoiceEffectList{Layers: c.Layers, Effects: []VoiceEffect{}, Kinds: StageKinds}
	byEffect := map[string][]EffectTrigger{}
	for _, t := range r.Triggers().Triggers {
		if t.EffectID == "" || (t.ResponseType != "activate_voice_effect" && t.ResponseType != "deactivate_voice_effect") {
			continue
		}
		byEffect[t.EffectID] = append(byEffect[t.EffectID], EffectTrigger{Path: t.Path, Name: t.Name, Source: t.Source, Activates: t.ResponseType == "activate_voice_effect", Enabled: t.Enabled})
	}
	for _, rec := range c.Records {
		if rec.Kind != "voice_effects" {
			continue
		}
		id := strings.TrimSuffix(filepath.Base(rec.Path), filepath.Ext(rec.Path))
		e := VoiceEffect{ID: id, Path: rec.Path, Name: rec.Name, Enabled: rec.Enabled, Override: rec.Override, Stages: []Stage{}, Providers: rec.Providers, Error: rec.Error, Triggers: byEffect[id]}
		if e.Triggers == nil {
			e.Triggers = []EffectTrigger{}
		}
		top := rec.Providers[0]
		if b, err := os.ReadFile(layers.Lookup(c.Layers[top.Layer].Root, top.Path)); err == nil && e.Error == "" {
			var doc struct {
				Name   string  `yaml:"name"`
				Sample *string `yaml:"sample"`
				Stages []Stage `yaml:"stages"`
			}
			if err := yaml.Unmarshal(b, &doc); err != nil {
				e.Error = yamlError(err)
			} else {
				if doc.Name != "" {
					e.Name = doc.Name
				}
				if doc.Sample != nil {
					e.Sample = *doc.Sample
				}
				for _, st := range doc.Stages {
					if st.Params == nil {
						st.Params = map[string]any{}
					}
					e.Stages = append(e.Stages, st)
				}
			}
		}
		out.Effects = append(out.Effects, e)
	}
	sortBy(out.Effects, func(e VoiceEffect) string { return strings.ToLower(e.Name) })
	return out
}

func sortBy[T any](s []T, key func(T) string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && key(s[j]) < key(s[j-1]); j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}

// EffectEdit is a recipe as the player's copy should have it.
type EffectEdit struct {
	Modified int64   `json:"modified"` // the copy's time when read, 0 for none yet
	Name     string  `json:"name"`
	Sample   string  `json:"sample"` // "" for none
	Stages   []Stage `json:"stages"`
}

var effectID = regexp.MustCompile(`^[a-z0-9][a-z0-9_]{0,63}$`)

// SaveEffect writes the player's copy of a recipe, or a new one. Only the
// name, sample, and stages are rewritten; the rest of the file, its
// comments among them, stays as it was.
func (r *Records) SaveEffect(id string, e EffectEdit) (VoiceEffectList, error) {
	if !effectID.MatchString(id) {
		return VoiceEffectList{}, apperr.New(apperr.Invalid, "An effect’s id is up to 64 lowercase letters, digits, and underscores, starting with a letter or digit, such as frost_atronach")
	}
	if strings.TrimSpace(e.Name) == "" {
		return VoiceEffectList{}, apperr.New(apperr.Invalid, "An effect needs a name")
	}
	if err := checkStages(e.Stages); err != nil {
		return VoiceEffectList{}, err
	}
	rel := "voice_effects/" + id + ".yaml"
	err := r.saveCopy(rel, e.Modified, "Saved voice effect "+id, func(from []byte) (string, error) {
		if len(from) == 0 {
			from = []byte("id: " + id + "\n")
		}
		out, err := settings.EditYAML(from, func(doc *yaml.Node) error {
			if doc.Kind != yaml.MappingNode {
				return apperr.New(apperr.Invalid, "%s isn’t a set of fields", rel)
			}
			set := func(key string, value any) error {
				var n yaml.Node
				if err := n.Encode(value); err != nil {
					return err
				}
				for i := 0; i+1 < len(doc.Content); i += 2 {
					if doc.Content[i].Value == key {
						old := doc.Content[i+1]
						if old.Kind == yaml.ScalarNode && n.Kind == yaml.ScalarNode && old.Tag == n.Tag {
							n.Style = old.Style
						}
						n.HeadComment, n.LineComment, n.FootComment = old.HeadComment, old.LineComment, old.FootComment
						doc.Content[i+1] = &n
						return nil
					}
				}
				doc.Content = append(doc.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}, &n)
				return nil
			}
			var sample any
			if s := strings.TrimSpace(e.Sample); s != "" {
				sample = s
			}
			stages := make([]any, len(e.Stages))
			for i, st := range e.Stages {
				stages[i] = stageNode(st)
			}
			for _, kv := range []struct {
				k string
				v any
			}{{"name", strings.TrimSpace(e.Name)}, {"sample", sample}, {"stages", stages}} {
				if err := set(kv.k, kv.v); err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return "", err
		}
		text := string(out)
		if err := checkRecord(rel, text); err != nil {
			return "", err
		}
		return text, nil
	})
	if err != nil {
		return VoiceEffectList{}, err
	}
	return r.VoiceEffects(), nil
}

// stageNode is a stage as a recipe writes it: type, then its params in the
// kind's order, numbers as the kind takes them.
func stageNode(st Stage) *yaml.Node {
	params := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
	add := func(name string, v any) {
		var n yaml.Node
		switch x := v.(type) {
		case float64:
			if k := stageKind(st.Type); k != nil {
				for _, pm := range k.Params {
					if pm.Name == name && pm.Whole {
						n.Encode(int64(math.Round(x)))
						goto done
					}
				}
			}
			// A float as SkyrimNet's recipes write one: 2.0, not 2.
			if x == math.Trunc(x) && math.Abs(x) < 1e15 {
				n = yaml.Node{Kind: yaml.ScalarNode, Tag: "!!float", Value: fmt.Sprintf("%.1f", x)}
			} else {
				n.Encode(x)
			}
		default:
			n.Encode(x)
		}
	done:
		params.Content = append(params.Content, &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: name}, &n)
	}
	seen := map[string]bool{}
	if k := stageKind(st.Type); k != nil {
		for _, pm := range k.Params {
			if v, ok := st.Params[pm.Name]; ok {
				add(pm.Name, v)
				seen[pm.Name] = true
			}
		}
	}
	var rest []string
	for name := range st.Params {
		if !seen[name] {
			rest = append(rest, name)
		}
	}
	sortBy(rest, func(s string) string { return s })
	for _, name := range rest {
		add(name, st.Params[name])
	}
	return &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map", Content: []*yaml.Node{
		{Kind: yaml.ScalarNode, Tag: "!!str", Value: "type"}, {Kind: yaml.ScalarNode, Tag: "!!str", Value: st.Type},
		{Kind: yaml.ScalarNode, Tag: "!!str", Value: "params"}, params,
	}}
}

// checkStages refuses a stage with no type. What else SkyrimNet accepts
// in a stage isn't published, so it isn't checked here.
func checkStages(stages []Stage) error {
	for i, st := range stages {
		if strings.TrimSpace(st.Type) == "" {
			return apperr.New(apperr.Invalid, "Stage %d needs a type", i+1)
		}
	}
	return nil
}

// effectIDOf is the voice effect a trigger's response names.
func effectIDOf(m map[string]any) string {
	rs, _ := m["response"].(map[string]any)
	s, _ := rs["effectId"].(string)
	return strings.TrimSpace(s)
}

func toFloat(v any) (float64, bool) {
	switch x := v.(type) {
	case int:
		return float64(x), true
	case int64:
		return float64(x), true
	case float64:
		return x, !math.IsNaN(x) && !math.IsInf(x, 0)
	}
	return 0, false
}
