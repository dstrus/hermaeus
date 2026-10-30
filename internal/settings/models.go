package settings

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"hermaeus/internal/apperr"
	"hermaeus/internal/textfile"

	"go.yaml.in/yaml/v3"
)

// Which model does each of SkyrimNet's jobs. They're all in one file,
// OpenRouter.yaml, whatever the provider: the Dialogue task is the file's
// root, and the rest are under variants:. A task's rotation is its
// model_name (a comma-separated list) paired by position with its
// model_params (one map per slot: the provider, a temperature, a role, and
// so on); default_params and timeout apply to the whole task. Providers
// themselves are in Providers.yaml, and are read here by name only.

const modelsFile = "OpenRouter"

// Task is one job SkyrimNet sends to a model, with its rotation of models.
type Task struct {
	ID    string `json:"id"` // default for Dialogue; otherwise the variant's name
	Label string `json:"label"`
	What  string `json:"what"`
	Known bool   `json:"known"` // one of SkyrimNet's own tasks; others come from plugins

	Slots        []Slot `json:"slots"` // in rotation order
	Params       Params `json:"params"`
	Timeout      *int   `json:"timeout,omitempty"`      // seconds; the Dialogue task's applies to those without one
	RotationMode string `json:"rotationMode,omitempty"` // sequential or random; "" follows the Dialogue task
}

// Slot is one model in a task's rotation, with what it overrides.
type Slot struct {
	Model             string   `json:"model"`
	Provider          string   `json:"provider"`         // a provider's id
	Origin            *int     `json:"origin,omitempty"` // its position when read, so its settings follow it when the rotation is reordered; none for a new slot
	Enabled           bool     `json:"enabled"`
	Role              string   `json:"role,omitempty"` // backup: only when no primary answers
	Weight            *float64 `json:"weight,omitempty"`
	Temperature       *float64 `json:"temperature,omitempty"`
	RequestTimeout    *int     `json:"requestTimeout,omitempty"`
	ConnectTimeout    *int     `json:"connectTimeout,omitempty"`
	MaxRetries        *int     `json:"maxRetries,omitempty"`
	ReasoningEffort   string   `json:"reasoningEffort,omitempty"`
	StructuredOutputs *bool    `json:"structuredOutputs,omitempty"`
}

// Params are a task's sampling settings, for every slot that doesn't say otherwise.
type Params struct {
	Temperature      *float64 `json:"temperature,omitempty"`
	MaxTokens        *int     `json:"maxTokens,omitempty"`
	TopP             *float64 `json:"topP,omitempty"`
	TopK             *int     `json:"topK,omitempty"`
	FrequencyPenalty *float64 `json:"frequencyPenalty,omitempty"`
	PresencePenalty  *float64 `json:"presencePenalty,omitempty"`
	StopSequences    []string `json:"stopSequences"`
}

// Provider is one of the providers in Providers.yaml, without its key.
type Provider struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Style   string `json:"style"` // chat, decisions, and so on
	HasKey  bool   `json:"hasKey"`
	Builtin bool   `json:"builtin"`
}

type Models struct {
	Found     bool       `json:"found"` // OpenRouter.yaml exists
	Path      string     `json:"path"`
	Modified  int64      `json:"modified"`
	Tasks     []Task     `json:"tasks"`
	Providers []Provider `json:"providers"`
}

// TaskEdit is a task's rotation and settings as they should be.
type TaskEdit struct {
	Modified     int64  `json:"modified"`
	Slots        []Slot `json:"slots"`
	Params       Params `json:"params"`
	Timeout      *int   `json:"timeout,omitempty"`
	RotationMode string `json:"rotationMode,omitempty"`
}

// SkyrimNet's own tasks, in the order its dashboard lists them.
var knownTasks = []struct{ id, label, what string }{
	{"default", "Dialogue", "What NPCs say in conversation, and any task without a model of its own."},
	{"combat", "Combat", "What NPCs say and do while fighting."},
	{"CharacterProfileGeneration", "Character profiles", "Writing and rewriting NPCs’ bios."},
	{"DiaryGeneration", "Diary", "NPCs’ diary entries."},
	{"memory", "Memory", "Turning events into memories, with how much they matter."},
	{"gamemaster_evaluation", "GameMaster", "Reading a scene and deciding who speaks or acts."},
	{"action_evaluation", "Action evaluation", "Which actions an NPC may take in response to an event."},
	{"meta", "Meta", "Quick sorting and routing: a fast, cheap model."},
	{"decisions", "Decisions", "Typed questions (an NPC’s mood, who speaks next) answered with probabilities. Needs a provider with the Decisions or System One API style; sampling settings don’t apply."},
	{"vision", "Vision", "Describing screenshots for OmniSight."},
	{"vision_quality", "Vision (quality)", "Describing the OmniSight captures you take yourself, with a stronger model."},
	{"UniversalTranslator", "Universal Translator", "Rewriting NPCs’ lines in their speech patterns."},
	{"AgentDefault", "AI assistant", "The dashboard’s tool-calling assistant."},
}

func knownTask(id string) (int, bool) {
	for i, t := range knownTasks {
		if t.id == id {
			return i, true
		}
	}
	return -1, false
}

// taskLabel names a plugin's task from its id: sever_background → Sever background.
func taskLabel(id string) string {
	words := strings.ReplaceAll(id, "_", " ")
	if words == "" {
		return id
	}
	return strings.ToUpper(words[:1]) + words[1:]
}

var taskIDRe = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)

func (s *Settings) modelsPath() (string, error) {
	dir, err := s.configDir()
	if err != nil {
		return "", err
	}
	return filePath(dir, modelsFile), nil
}

// Models reads every task's models, and the providers they can use.
func (s *Settings) Models() (Models, error) {
	path, err := s.modelsPath()
	if err != nil {
		return Models{}, err
	}
	out := Models{Tasks: []Task{}, Providers: []Provider{}}
	rel, _ := filepath.Rel(s.store.Dir(), path)
	out.Path = filepath.ToSlash(rel)
	if providers, err := s.providers(); err == nil {
		out.Providers = providers
	}
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return out, nil
	}
	if err != nil {
		return Models{}, err
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return Models{}, err
	}
	d, err := parse(b)
	if err != nil {
		return Models{}, apperr.New(apperr.Invalid, "Could not read %s: %v", filepath.Base(path), err)
	}
	out.Found = true
	out.Modified = info.ModTime().UnixMilli()
	root := findNode(d.doc, "openrouter")
	if root == nil || root.Kind != yaml.MappingNode {
		return out, nil
	}
	out.Tasks = append(out.Tasks, taskOf("default", root))
	if variants := findNode(root, "variants"); variants != nil && variants.Kind == yaml.MappingNode {
		var others []Task
		for i := 0; i+1 < len(variants.Content); i += 2 {
			if variants.Content[i+1].Kind == yaml.MappingNode {
				others = append(others, taskOf(variants.Content[i].Value, variants.Content[i+1]))
			}
		}
		// SkyrimNet's own tasks in its order, then plugins' by name.
		sort.SliceStable(others, func(i, j int) bool {
			a, aKnown := knownTask(others[i].ID)
			b, bKnown := knownTask(others[j].ID)
			if aKnown != bKnown {
				return aKnown
			}
			if aKnown {
				return a < b
			}
			return strings.ToLower(others[i].Label) < strings.ToLower(others[j].Label)
		})
		out.Tasks = append(out.Tasks, others...)
	}
	return out, nil
}

func taskOf(id string, n *yaml.Node) Task {
	t := Task{ID: id, Label: taskLabel(id), Slots: []Slot{}, Params: Params{StopSequences: []string{}}}
	if i, ok := knownTask(id); ok {
		t.Label, t.What, t.Known = knownTasks[i].label, knownTasks[i].what, true
	}
	names := splitModels(ruleField(n, "model_name"))
	params := findNode(n, "model_params")
	for i, name := range names {
		at := i
		slot := Slot{Model: name, Enabled: true, Origin: &at}
		if params != nil && params.Kind == yaml.SequenceNode && i < len(params.Content) && params.Content[i].Kind == yaml.MappingNode {
			p := params.Content[i]
			slot.Provider = ruleField(p, "provider_id")
			slot.Enabled = ruleField(p, "enabled") != "false"
			slot.Role = ruleField(p, "routing_role")
			slot.Weight = floatField(p, "rotation_weight")
			slot.Temperature = floatField(p, "temperature")
			slot.RequestTimeout = intField(p, "request_timeout")
			slot.ConnectTimeout = intField(p, "connect_timeout")
			slot.MaxRetries = intField(p, "max_retries")
			slot.ReasoningEffort = ruleField(p, "reasoning_effort")
			slot.StructuredOutputs = boolField(p, "use_structured_outputs")
		}
		t.Slots = append(t.Slots, slot)
	}
	if dp := findNode(n, "default_params"); dp != nil && dp.Kind == yaml.MappingNode {
		t.Params = Params{
			Temperature:      floatField(dp, "temperature"),
			MaxTokens:        intField(dp, "max_tokens"),
			TopP:             floatField(dp, "top_p"),
			TopK:             intField(dp, "top_k"),
			FrequencyPenalty: floatField(dp, "frequency_penalty"),
			PresencePenalty:  floatField(dp, "presence_penalty"),
			StopSequences:    []string{},
		}
		if ss := findNode(dp, "stop_sequences"); ss != nil && ss.Kind == yaml.SequenceNode {
			for _, v := range ss.Content {
				if v.Kind == yaml.ScalarNode {
					t.Params.StopSequences = append(t.Params.StopSequences, v.Value)
				}
			}
		}
	}
	t.Timeout = intField(n, "timeout")
	t.RotationMode = ruleField(n, "rotation_mode")
	return t
}

// splitModels reads a model_name list: "a, b" is two models.
func splitModels(s string) []string {
	var out []string
	for _, m := range strings.Split(s, ",") {
		if m = strings.TrimSpace(m); m != "" {
			out = append(out, m)
		}
	}
	return out
}

func floatField(m *yaml.Node, key string) *float64 {
	n := findNode(m, key)
	if n == nil || n.Kind != yaml.ScalarNode || n.Tag == "!!null" {
		return nil
	}
	var f float64
	if err := n.Decode(&f); err != nil {
		return nil
	}
	return &f
}

func intField(m *yaml.Node, key string) *int {
	f := floatField(m, key)
	if f == nil {
		return nil
	}
	i := int(*f)
	return &i
}

func boolField(m *yaml.Node, key string) *bool {
	switch ruleField(m, key) {
	case "true":
		t := true
		return &t
	case "false":
		f := false
		return &f
	}
	return nil
}

// providers lists Providers.yaml's providers, never their keys.
func (s *Settings) providers() ([]Provider, error) {
	dir, err := s.configDir()
	if err != nil {
		return nil, err
	}
	nodes, err := providerNodes(dir)
	if err != nil {
		return nil, err
	}
	out := make([]Provider, 0, len(nodes))
	for _, p := range nodes {
		out = append(out, providerOf(p))
	}
	return out, nil
}

// providerNodes reads Providers.yaml's providers: list.
func providerNodes(configDir string) ([]*yaml.Node, error) {
	b, err := os.ReadFile(filePath(configDir, "Providers"))
	if err != nil {
		return nil, err
	}
	d, err := parse(b)
	if err != nil {
		return nil, err
	}
	var out []*yaml.Node
	if list := findNode(d.doc, "providers"); list != nil && list.Kind == yaml.SequenceNode {
		for _, p := range list.Content {
			if p.Kind == yaml.MappingNode {
				out = append(out, p)
			}
		}
	}
	return out, nil
}

func providerOf(p *yaml.Node) Provider {
	return Provider{
		ID:      ruleField(p, "id"),
		Name:    ruleField(p, "name"),
		Style:   ruleField(p, "api_style"),
		HasKey:  strings.TrimSpace(ruleField(p, "api_key")) != "",
		Builtin: ruleField(p, "builtin") == "true",
	}
}

var reasoningEfforts = map[string]bool{"": true, "off": true, "minimal": true, "low": true, "medium": true, "high": true, "default": true}

// SetTask writes one task's rotation and settings. Each slot keeps the
// model_params node it was read from, so what this page doesn't manage (a
// slot's provider settings, say) stays with it.
func (s *Settings) SetTask(id string, e TaskEdit) (Models, error) {
	if !taskIDRe.MatchString(id) {
		return Models{}, apperr.New(apperr.Invalid, "Invalid task %q", id)
	}
	if e.RotationMode != "" && e.RotationMode != "sequential" && e.RotationMode != "random" {
		return Models{}, apperr.New(apperr.Invalid, "The rotation order is sequential or random, not %q", e.RotationMode)
	}
	providers, _ := s.providers()
	for i := range e.Slots {
		sl := &e.Slots[i]
		sl.Model = strings.TrimSpace(sl.Model)
		sl.Provider = strings.TrimSpace(sl.Provider)
		if sl.Model == "" {
			return Models{}, apperr.New(apperr.Invalid, "Every slot in the rotation names a model")
		}
		if strings.Contains(sl.Model, ",") {
			return Models{}, apperr.New(apperr.Invalid, "A model's name can't hold a comma: SkyrimNet lists the rotation with them")
		}
		if sl.Provider == "" {
			return Models{}, apperr.New(apperr.Invalid, "Every slot in the rotation names a provider")
		}
		if len(providers) > 0 {
			found := false
			for _, p := range providers {
				found = found || p.ID == sl.Provider
			}
			if !found {
				return Models{}, apperr.New(apperr.Invalid, "There's no provider %q in Providers.yaml", sl.Provider)
			}
		}
		if sl.Role != "" && sl.Role != "backup" {
			return Models{}, apperr.New(apperr.Invalid, "A slot is primary or a backup, not %q", sl.Role)
		}
		if !reasoningEfforts[sl.ReasoningEffort] {
			return Models{}, apperr.New(apperr.Invalid, "Unknown reasoning effort %q", sl.ReasoningEffort)
		}
	}
	err := s.store.ChangeFiles("Saved the models for "+id, func() error {
		path, err := s.modelsPath()
		if err != nil {
			return err
		}
		info, err := os.Stat(path)
		if os.IsNotExist(err) {
			return apperr.New(apperr.NotFound, "%s.yaml doesn't exist yet. SkyrimNet creates it the next time it starts.", modelsFile)
		}
		if err != nil {
			return err
		}
		if info.ModTime().UnixMilli() != e.Modified {
			return apperr.New(apperr.Conflict, "The models changed after you opened them. Reload to see the latest, then make your change again.")
		}
		old, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		updated, err := rewrite(old, modelsFile, func(doc *yaml.Node) error {
			root := findNode(doc, "openrouter")
			if root == nil || root.Kind != yaml.MappingNode {
				return apperr.New(apperr.Invalid, "%s.yaml has no openrouter: section", modelsFile)
			}
			task := root
			if id != "default" {
				variants := findNode(root, "variants")
				if variants == nil || variants.Kind != yaml.MappingNode {
					return apperr.New(apperr.NotFound, "There's no task %q", id)
				}
				if task = findNode(variants, id); task == nil || task.Kind != yaml.MappingNode {
					return apperr.New(apperr.NotFound, "There's no task %q", id)
				}
			}
			writeTask(task, e)
			return nil
		})
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
		return Models{}, err
	}
	return s.Models()
}

func writeTask(task *yaml.Node, e TaskEdit) {
	names := make([]string, len(e.Slots))
	for i, sl := range e.Slots {
		names[i] = sl.Model
	}
	setField(task, "model_name", stringNode(strings.Join(names, ", ")))

	params := findNode(task, "model_params")
	if params == nil || params.Kind != yaml.SequenceNode {
		params = &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
		setField(task, "model_params", params)
	}
	content := make([]*yaml.Node, len(e.Slots))
	used := map[int]bool{}
	for i, sl := range e.Slots {
		var p *yaml.Node
		if o := sl.Origin; o != nil && *o >= 0 && *o < len(params.Content) && params.Content[*o].Kind == yaml.MappingNode && !used[*o] {
			p = params.Content[*o]
			used[*o] = true
		} else {
			p = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		}
		setField(p, "provider_id", scalar(sl.Provider))
		setOrDrop(p, "enabled", boolValue(!sl.Enabled, false))
		setOrDrop(p, "routing_role", stringValue(sl.Role))
		setOrDrop(p, "rotation_weight", floatValue(sl.Weight))
		setOrDrop(p, "temperature", floatValue(sl.Temperature))
		setOrDrop(p, "request_timeout", intValue(sl.RequestTimeout))
		setOrDrop(p, "connect_timeout", intValue(sl.ConnectTimeout))
		setOrDrop(p, "max_retries", intValue(sl.MaxRetries))
		setOrDrop(p, "reasoning_effort", stringValue(sl.ReasoningEffort))
		setOrDrop(p, "use_structured_outputs", boolPtrValue(sl.StructuredOutputs))
		content[i] = p
	}
	params.Content = content

	dp := findNode(task, "default_params")
	if dp == nil || dp.Kind != yaml.MappingNode {
		dp = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		setField(task, "default_params", dp)
	}
	// A setting left empty stays as a key with no value, the way SkyrimNet
	// writes the ones it doesn't set.
	setOrNull(dp, "temperature", floatValue(e.Params.Temperature))
	setOrNull(dp, "max_tokens", intValue(e.Params.MaxTokens))
	setOrNull(dp, "top_p", floatValue(e.Params.TopP))
	setOrNull(dp, "top_k", intValue(e.Params.TopK))
	setOrNull(dp, "frequency_penalty", floatValue(e.Params.FrequencyPenalty))
	setOrNull(dp, "presence_penalty", floatValue(e.Params.PresencePenalty))
	if len(e.Params.StopSequences) > 0 {
		setField(dp, "stop_sequences", listNode(e.Params.StopSequences))
	} else {
		setOrNull(dp, "stop_sequences", nil)
	}

	setOrDrop(task, "timeout", intValue(e.Timeout))
	setOrDrop(task, "rotation_mode", stringValue(e.RotationMode))
}

// setOrDrop sets a key, or removes it when there's no value.
func setOrDrop(m *yaml.Node, key string, value *yaml.Node) {
	if value != nil {
		setField(m, key, value)
		return
	}
	for i := 0; i+1 < len(m.Content); i += 2 {
		if m.Content[i].Value == key {
			m.Content = append(m.Content[:i], m.Content[i+2:]...)
			return
		}
	}
}

// setOrNull sets a key, or leaves it with no value (~) when there's none.
func setOrNull(m *yaml.Node, key string, value *yaml.Node) {
	if value == nil {
		value = &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!null", Value: "~"}
	}
	setField(m, key, value)
}

func stringValue(s string) *yaml.Node {
	if s == "" {
		return nil
	}
	return stringNode(s)
}

func floatValue(f *float64) *yaml.Node {
	if f == nil {
		return nil
	}
	var n yaml.Node
	n.Encode(*f)
	return &n
}

func intValue(i *int) *yaml.Node {
	if i == nil {
		return nil
	}
	var n yaml.Node
	n.Encode(*i)
	return &n
}

// boolValue is the key's node when set is true, else none.
func boolValue(set, v bool) *yaml.Node {
	if !set {
		return nil
	}
	return boolNode(v)
}

func boolPtrValue(b *bool) *yaml.Node {
	if b == nil {
		return nil
	}
	return boolNode(*b)
}
