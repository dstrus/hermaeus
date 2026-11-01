package settings

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"hermaeus/internal/apperr"
)

func toJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

const modelsYAML = `openrouter:
  model_name: google/gemma-4-31b-it, deepseek/deepseek-v3.2
  api_key: sk-or-SECRET
  timeout: 30
  rotation_mode: sequential
  default_params:
    temperature: 0.7
    max_tokens: 4096
    top_p: ~
    top_k: ~
    frequency_penalty: ~
    presence_penalty: ~
    stop_sequences: ~
  provider_settings: "{\"route\": \"fallback\"}"
  model_params:
    - provider_id: openrouter
      temperature: 1
    - provider_id: openrouter
      temperature: 0.7
  variants:
    sever_background:
      model_name: google/gemma-4-31b-it
      model_params:
        - provider_id: openrouter
    memory:
      model_name: deepseek/deepseek-v3.2, qwen/qwen3-235b-a22b-2507
      timeout: 60
      default_params:
        temperature: 0.7
        max_tokens: 4000
        top_p: ~
        top_k: ~
        frequency_penalty: ~
        presence_penalty: ~
        stop_sequences: ~
      model_params:
        - provider_id: openrouter
          request_timeout: 60
          provider_settings: "{\"cache\": true}"
        - provider_id: openrouter
          routing_role: backup
          temperature: 0.7
    combat:
      model_name: google/gemma-4-26b-a4b-it
      model_params:
        - provider_id: nanogpt
          temperature: 1
active_preset_id: ""
`

const providersYAML = "providers:\n  - id: openrouter\n    name: OpenRouter\n    api_key: sk-or-SECRET\n    api_style: chat\n    builtin: true\n  - id: nanogpt\n    name: NanoGPT\n    api_key: \"\"\n    api_style: chat\n    builtin: true\n"

func TestModelsReadsTheTasks(t *testing.T) {
	h := newHarness(t)
	m, err := h.settings.Models()
	if err != nil || m.Found || len(m.Tasks) != 0 {
		t.Fatalf("without the file: %+v, %v", m, err)
	}
	h.config("OpenRouter.yaml", modelsYAML)
	h.config("Providers.yaml", providersYAML)
	m, err = h.settings.Models()
	if err != nil || !m.Found || m.Path != "config/OpenRouter.yaml" {
		t.Fatalf("Models() = %+v, %v", m, err)
	}
	ids := make([]string, len(m.Tasks))
	for i, task := range m.Tasks {
		ids[i] = task.ID
	}
	if strings.Join(ids, " ") != "default combat memory sever_background" {
		t.Errorf("tasks in order = %v", ids)
	}
	d := m.Tasks[0]
	if d.label != "Dialogue" || !d.Known || len(d.Slots) != 2 || d.Slots[0].Model != "google/gemma-4-31b-it" || d.Slots[0].Provider != "openrouter" || *d.Slots[0].Temperature != 1 || *d.Slots[1].Temperature != 0.7 {
		t.Errorf("default = %+v", d)
	}
	if d.Timeout == nil || *d.Timeout != 30 || d.RotationMode != "sequential" || d.Params.MaxTokens == nil || *d.Params.MaxTokens != 4096 || d.Params.TopP != nil || len(d.Params.StopSequences) != 0 {
		t.Errorf("default settings = %+v %+v", d.Timeout, d.Params)
	}
	mem := m.Tasks[2]
	if mem.label != "Memory" || len(mem.Slots) != 2 || mem.Slots[0].RequestTimeout == nil || *mem.Slots[0].RequestTimeout != 60 || mem.Slots[1].Role != "backup" || !mem.Slots[1].Enabled {
		t.Errorf("memory = %+v", mem)
	}
	if sev := m.Tasks[3]; sev.Known || sev.label != "Sever background" || sev.Timeout != nil || sev.RotationMode != "" {
		t.Errorf("plugin task = %+v", sev)
	}
	if len(m.Providers) != 2 || m.Providers[0].Name != "OpenRouter" || !m.Providers[0].HasKey || m.Providers[1].HasKey {
		t.Errorf("providers = %+v", m.Providers)
	}
	if sent := toJSON(t, m); strings.Contains(sent, "SECRET") {
		t.Errorf("a key was sent: %s", sent)
	}
}

func TestModelsSaveMinimally(t *testing.T) {
	h := newHarness(t)
	path := h.config("OpenRouter.yaml", modelsYAML)
	h.config("Providers.yaml", providersYAML)
	past := time.Now().Add(-time.Hour)
	os.Chtimes(path, past, past)
	m, _ := h.settings.Models()
	edit := func(task Task) TaskEdit {
		slots := make([]Slot, len(task.Slots))
		copy(slots, task.Slots)
		return TaskEdit{Modified: m.Modified, Slots: slots, Params: task.Params, Timeout: task.Timeout, RotationMode: task.RotationMode}
	}

	// Swapping one model changes the model_name line only.
	e := edit(m.Tasks[0])
	e.Slots[1].Model = "deepseek/deepseek-v4"
	m, err := h.settings.SetTask("default", e)
	if err != nil {
		t.Fatal(err)
	}
	got := h.read("OpenRouter.yaml")
	want := strings.Replace(modelsYAML, "google/gemma-4-31b-it, deepseek/deepseek-v3.2\n  api_key", "google/gemma-4-31b-it, deepseek/deepseek-v4\n  api_key", 1)
	if got != want {
		t.Errorf("after one model:\n%s\nwant\n%s", got, want)
	}

	// Reordering memory's rotation keeps each slot's own settings with it,
	// including one this page doesn't manage.
	e = edit(m.Tasks[2])
	e.Slots[0], e.Slots[1] = e.Slots[1], e.Slots[0]
	e.Slots[1].Role = ""
	e.Slots[0].Role = "backup"
	m, err = h.settings.SetTask("memory", e)
	if err != nil {
		t.Fatal(err)
	}
	got = h.read("OpenRouter.yaml")
	if !strings.Contains(got, "      model_name: qwen/qwen3-235b-a22b-2507, deepseek/deepseek-v3.2\n") ||
		!strings.Contains(got, "      model_params:\n        - provider_id: openrouter\n          routing_role: backup\n          temperature: 0.7\n        - provider_id: openrouter\n          request_timeout: 60\n          provider_settings: \"{\\\"cache\\\": true}\"\n") {
		t.Errorf("after reordering =\n%s", got)
	}
	if mem := m.Tasks[2]; mem.Slots[0].Role != "backup" || mem.Slots[1].Role != "" || mem.Slots[1].RequestTimeout == nil {
		t.Errorf("after reordering, read back = %+v", mem.Slots)
	}

	// Adding a slot, a weight, random order, a stop sequence, and dropping
	// the task's timeout; an untouched setting stays ~.
	e = edit(m.Tasks[1])
	w := 2.0
	tk := 8000
	e.Slots = append(e.Slots, Slot{Model: "mistral/large", Provider: "openrouter", Enabled: false, Weight: &w, ReasoningEffort: "low"})
	e.RotationMode = "random"
	e.Params.MaxTokens = &tk
	e.Params.StopSequences = []string{"\n\n", "END"}
	m, err = h.settings.SetTask("combat", e)
	if err != nil {
		t.Fatal(err)
	}
	got = h.read("OpenRouter.yaml")
	for _, want := range []string{
		"      model_name: google/gemma-4-26b-a4b-it, mistral/large\n",
		"        - provider_id: openrouter\n          enabled: false\n          rotation_weight: 2\n          reasoning_effort: low\n",
		"      rotation_mode: random\n",
		"        max_tokens: 8000\n        top_p: ~\n",
		"        stop_sequences:\n          - \"\\n\\n\"\n          - END\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("after adding, missing %q in\n%s", want, got)
		}
	}
	if c := m.Tasks[1]; len(c.Slots) != 2 || c.Slots[1].Enabled || *c.Slots[1].Weight != 2 || c.RotationMode != "random" || len(c.Params.StopSequences) != 2 {
		t.Errorf("combat read back = %+v", c)
	}
	// The secret and the preset id outside the tasks are untouched.
	if !strings.Contains(got, "  api_key: sk-or-SECRET\n") || !strings.HasSuffix(got, "active_preset_id: \"\"\n") {
		t.Errorf("the rest of the file changed:\n%s", got)
	}

	// What's refused.
	e = edit(m.Tasks[0])
	e.Slots[0].Provider = "nowhere"
	_, err = h.settings.SetTask("default", e)
	kindOf(t, err, apperr.Invalid)
	e = edit(m.Tasks[0])
	e.Slots[0].Model = ""
	_, err = h.settings.SetTask("default", e)
	kindOf(t, err, apperr.Invalid)
	e = edit(m.Tasks[0])
	e.Slots[0].Role = "tertiary"
	_, err = h.settings.SetTask("default", e)
	kindOf(t, err, apperr.Invalid)
	_, err = h.settings.SetTask("nothing", edit(m.Tasks[0]))
	kindOf(t, err, apperr.NotFound)
	e = edit(m.Tasks[0])
	e.Modified = 1
	_, err = h.settings.SetTask("default", e)
	kindOf(t, err, apperr.Conflict)
	h.running = true
	_, err = h.settings.SetTask("default", edit(m.Tasks[0]))
	kindOf(t, err, apperr.Conflict)
}

// From Beta 29, SkyrimNet makes Fateless's providers itself from its
// sign-in: a task may use them although Providers.yaml doesn't list them,
// and Hermaeus says why it can't test or list them.
func TestFatelessProviders(t *testing.T) {
	h := newHarness(t)
	h.config("OpenRouter.yaml", modelsYAML)
	h.config("Providers.yaml", providersYAML)
	m, _ := h.settings.Models()
	for _, p := range m.Providers {
		if p.SignIn {
			t.Fatalf("Fateless listed for a SkyrimNet that doesn't know it: %+v", p)
		}
	}

	h.config("defaults_manifest.json", `{"fields": {"FatelessTTS": {"model": "fateless/speech-plus"}}}`)
	m, _ = h.settings.Models()
	var ids []string
	for _, p := range m.Providers {
		ids = append(ids, p.ID)
	}
	if got := strings.Join(ids, " "); got != "openrouter nanogpt fateless fateless_decisions" {
		t.Fatalf("providers = %s", got)
	}
	task := m.Tasks[0]
	slots := append([]Slot{}, task.Slots...)
	slots[0].Provider, slots[0].Model = "fateless", "fateless/dialogue"
	if _, err := h.settings.SetTask("default", TaskEdit{Modified: m.Modified, Slots: slots, Params: task.Params, Timeout: task.Timeout, RotationMode: task.RotationMode}); err != nil {
		t.Fatalf("a task on Fateless was refused: %v", err)
	}
	_, err := h.settings.ModelList("fateless", false)
	kindOf(t, err, apperr.Invalid)
	_, err = h.settings.TestModel("default", ModelTest{Slot: 0, Prompt: "Hello."})
	kindOf(t, err, apperr.Invalid)
}
