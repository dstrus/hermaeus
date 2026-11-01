package settings

import (
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"hermaeus/internal/apperr"
)

const providersFull = `providers:
  - id: openrouter
    name: OpenRouter
    endpoint: https://openrouter.ai/api/v1/chat/completions
    api_key: sk-or-SECRET
    api_style: chat
    use_sse: true
    use_structured_outputs: false
    reasoning: false
    provider_settings:
      cache: true
      provider:
        ignore:
          - atlas-cloud
      route: fallback
    provider_sorting: latency
    upstream_order_mode: fleet
    cache_ttl: 5m
    builtin: true
  - id: nanogpt
    name: NanoGPT
    endpoint: https://nano-gpt.com/api/v1/chat/completions
    api_key: ""
    api_style: chat
    use_sse: false
    use_structured_outputs: false
    reasoning: false
    provider_settings: ~
    provider_sorting: default
    upstream_order_mode: fleet
    cache_ttl: 5m
    builtin: true
  - id: openrouter_decisions
    name: OpenRouter (Decisions)
    endpoint: https://openrouter.ai/api/alpha/decisions
    api_key: ""
    api_key_from: openrouter
    api_style: decisions
    use_sse: true
    use_structured_outputs: false
    reasoning: false
    provider_settings: ~
    provider_sorting: default
    upstream_order_mode: fleet
    cache_ttl: 5m
    builtin: true
`

func TestProvidersReadWithoutKeys(t *testing.T) {
	h := newHarness(t)
	r, err := h.settings.Providers()
	if err != nil || r.Found || len(r.Providers) != 0 {
		t.Fatalf("without the file: %+v, %v", r, err)
	}
	h.config("Providers.yaml", providersFull)
	h.config("OpenRouter.yaml", modelsYAML)
	r, err = h.settings.Providers()
	if err != nil || !r.Found || len(r.Providers) != 3 || r.Path != "config/Providers.yaml" {
		t.Fatalf("Providers() = %+v, %v", r, err)
	}
	or := r.Providers[0]
	if !or.HasKey || !or.Builtin || or.Style != "chat" || !or.Streaming || or.Sorting != "latency" || or.CacheTTL != "5m" || !strings.Contains(or.Settings, `"route":"fallback"`) || strings.Join(or.Tasks, ",") != "default,memory,sever_background" {
		t.Errorf("openrouter = %+v", or)
	}
	if d := r.Providers[2]; d.HasKey || d.KeyFrom != "openrouter" || d.Style != "decisions" || d.Settings != "" || len(d.Tasks) != 0 {
		t.Errorf("decisions = %+v", d)
	}
	if sent, _ := json.Marshal(r); strings.Contains(string(sent), "SECRET") {
		t.Error("a key was sent to the page")
	}
}

func TestProvidersChange(t *testing.T) {
	h := newHarness(t)
	path := h.config("Providers.yaml", providersFull)
	h.config("OpenRouter.yaml", modelsYAML)
	past := time.Now().Add(-time.Hour)
	os.Chtimes(path, past, past)
	r, _ := h.settings.Providers()
	edit := func(p ProviderDetails) ProviderEdit {
		return ProviderEdit{Modified: r.Modified, Name: p.Name, Endpoint: p.Endpoint, Style: p.Style, Streaming: p.Streaming, StructuredOutputs: p.StructuredOutputs, Reasoning: p.Reasoning, Settings: p.Settings, Sorting: p.Sorting, CacheTTL: p.CacheTTL}
	}

	// Switching streaming off changes one line; the key stays.
	e := edit(r.Providers[1])
	e.Streaming = true
	r, err := h.settings.SetProvider("nanogpt", e)
	if err != nil {
		t.Fatal(err)
	}
	got := h.read("Providers.yaml")
	if want := strings.Replace(providersFull, "    use_sse: false\n", "    use_sse: true\n", 1); got != want {
		t.Errorf("after one switch:\n%s\nwant\n%s", got, want)
	}

	// The key: set, never read back, then cleared.
	r, err = h.settings.SetKey("nanogpt", KeyEdit{Modified: r.Modified, Key: " sk-nano-NEW\n"})
	if err != nil || !r.Providers[1].HasKey {
		t.Fatalf("after setting a key: %+v, %v", r.Providers[1], err)
	}
	// Written in place, so it keeps the quotes the empty key had.
	if got := h.read("Providers.yaml"); !strings.Contains(got, "    api_key: \"sk-nano-NEW\"\n") && !strings.Contains(got, "    api_key: sk-nano-NEW\n") {
		t.Errorf("key not written:\n%s", got)
	}
	if sent, _ := json.Marshal(r); strings.Contains(string(sent), "NEW") {
		t.Error("the key was sent to the page")
	}
	r, err = h.settings.SetKey("nanogpt", KeyEdit{Modified: r.Modified})
	if err != nil || r.Providers[1].HasKey {
		t.Fatalf("after clearing: %+v, %v", r.Providers[1], err)
	}
	if got := h.read("Providers.yaml"); !strings.Contains(got, "    api_key: \"\"\n    api_style: chat\n    use_sse: true\n") {
		t.Errorf("key not cleared:\n%s", got)
	}
	_, err = h.settings.SetKey("nanogpt", KeyEdit{Modified: r.Modified, Key: "two words"})
	kindOf(t, err, apperr.Invalid)

	// A new provider, then its settings as JSON, then its removal.
	r, err = h.settings.AddProvider(ProviderEdit{Modified: r.Modified, Name: "My Local LLM", Endpoint: "http://localhost:1234/v1/chat/completions", Style: "chat", Streaming: true})
	if err != nil || len(r.Providers) != 4 || r.Providers[3].ID != "my_local_llm" || r.Providers[3].Builtin || r.Providers[3].HasKey {
		t.Fatalf("after adding: %+v, %v", r.Providers, err)
	}
	got = h.read("Providers.yaml")
	if !strings.Contains(got, "  - id: my_local_llm\n    name: My Local LLM\n    endpoint: http://localhost:1234/v1/chat/completions\n    api_key: \"\"\n    api_style: chat\n    use_sse: true\n") || !strings.Contains(got, "    provider_settings: ~\n    provider_sorting: default\n    cache_ttl: 5m\n    builtin: false\n") {
		t.Errorf("added =\n%s", got)
	}
	e = edit(r.Providers[3])
	e.Settings = `{"route": "fallback", "cache": true}`
	e.CacheTTL = "1h"
	r, err = h.settings.SetProvider("my_local_llm", e)
	if err != nil || !strings.Contains(r.Providers[3].Settings, `"cache":true`) || r.Providers[3].CacheTTL != "1h" {
		t.Fatalf("after settings: %+v, %v", r.Providers[3], err)
	}
	if got := h.read("Providers.yaml"); !strings.Contains(got, "    provider_settings:\n      cache: true\n      route: fallback\n") {
		t.Errorf("settings =\n%s", got)
	}
	r, err = h.settings.RemoveProvider("my_local_llm", r.Modified)
	if err != nil || len(r.Providers) != 3 {
		t.Fatalf("after removing: %d providers, %v", len(r.Providers), err)
	}

	// What's refused: a built-in, one in use, bad fields, a stale file.
	_, err = h.settings.RemoveProvider("nanogpt", r.Modified)
	kindOf(t, err, apperr.Invalid)
	h.config("OpenRouter.yaml", strings.Replace(modelsYAML, "provider_id: nanogpt", "provider_id: nanogpt", 1))
	r2, _ := h.settings.AddProvider(ProviderEdit{Modified: r.Modified, Name: "Used", Endpoint: "https://x.example/v1/chat/completions", Style: "chat"})
	h.config("OpenRouter.yaml", strings.Replace(modelsYAML, "provider_id: nanogpt", "provider_id: used", 1))
	_, err = h.settings.RemoveProvider("used", r2.Modified)
	kindOf(t, err, apperr.Conflict)
	if err == nil || !strings.Contains(err.Error(), "used by Combat.") {
		t.Errorf("in use: %v, want the task named", err)
	}
	r = r2
	e = edit(r.Providers[0])
	e.Endpoint = "openrouter.ai"
	_, err = h.settings.SetProvider("openrouter", e)
	kindOf(t, err, apperr.Invalid)
	e = edit(r.Providers[0])
	e.Settings = "not json"
	_, err = h.settings.SetProvider("openrouter", e)
	kindOf(t, err, apperr.Invalid)
	e = edit(r.Providers[0])
	e.Style = "grpc"
	_, err = h.settings.SetProvider("openrouter", e)
	kindOf(t, err, apperr.Invalid)
	e = edit(r.Providers[0])
	e.Modified = 1
	_, err = h.settings.SetProvider("openrouter", e)
	kindOf(t, err, apperr.Conflict)
	_, err = h.settings.SetProvider("nowhere", edit(r.Providers[0]))
	kindOf(t, err, apperr.NotFound)
}

func TestProviderSlug(t *testing.T) {
	for in, want := range map[string]string{"My Local LLM": "my_local_llm", "  Groq! ": "groq", "---": "provider", "X.ai": "x_ai"} {
		if got := slug(in); got != want {
			t.Errorf("slug(%q) = %q, want %q", in, got, want)
		}
	}
}
