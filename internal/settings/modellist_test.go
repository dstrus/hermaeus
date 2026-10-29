package settings

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"hermaeus/internal/apperr"
)

func TestModelListFromTheProvider(t *testing.T) {
	h := newHarness(t)
	calls := 0
	var auth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		auth = r.Header.Get("Authorization")
		if r.URL.Path != "/api/v1/models" {
			http.NotFound(w, r)
			return
		}
		fmt.Fprint(w, `{"data": [
			{"id": "zeta/z-1", "name": "Z One", "context_length": 32000, "pricing": {"prompt": "0.0000015", "completion": "0.000006"}},
			{"id": "alpha/a-2", "name": "A Two", "context_length": 128000, "pricing": {"prompt": 0, "completion": 0}},
			{"id": ""}
		]}`)
	}))
	defer srv.Close()
	h.config("Providers.yaml", "providers:\n  - id: openrouter\n    name: OpenRouter\n    endpoint: "+srv.URL+"/api/v1/chat/completions\n    api_key: sk-or-SECRET\n    api_style: chat\n  - id: decisions\n    name: Decisions\n    endpoint: "+srv.URL+"/api/alpha/decisions\n    api_key: \"\"\n    api_style: decisions\n")

	list, err := h.settings.ModelList("openrouter", false)
	if err != nil {
		t.Fatal(err)
	}
	if auth != "Bearer sk-or-SECRET" {
		t.Errorf("sent %q, want the provider's key", auth)
	}
	if len(list.Models) != 2 || list.Models[0].ID != "alpha/a-2" || list.Models[1].Name != "Z One" || list.Models[1].Context != 32000 || *list.Models[1].Prompt != 1.5 || *list.Models[1].Completion != 6 {
		t.Errorf("models = %+v", list.Models)
	}
	if list.Models[0].Prompt == nil || *list.Models[0].Prompt != 0 {
		t.Errorf("a free model's price = %v", list.Models[0].Prompt)
	}
	if sent, _ := json.Marshal(list); strings.Contains(string(sent), "SECRET") {
		t.Error("the key was sent to the page")
	}
	// Cached until asked to refresh.
	if _, err := h.settings.ModelList("openrouter", false); err != nil || calls != 1 {
		t.Errorf("second read: calls = %d, %v", calls, err)
	}
	if _, err := h.settings.ModelList("openrouter", true); err != nil || calls != 2 {
		t.Errorf("refresh: calls = %d, %v", calls, err)
	}
	_, err = h.settings.ModelList("decisions", false)
	kindOf(t, err, apperr.Invalid)
	_, err = h.settings.ModelList("nowhere", false)
	kindOf(t, err, apperr.NotFound)
}

func TestModelsURL(t *testing.T) {
	for in, want := range map[string]string{
		"https://openrouter.ai/api/v1/chat/completions":         "https://openrouter.ai/api/v1/models",
		"https://integrate.api.nvidia.com/v1/chat/completions/": "https://integrate.api.nvidia.com/v1/models",
		"http://localhost:1234/v1/completions":                  "http://localhost:1234/v1/models",
	} {
		if got, err := modelsURL(in); err != nil || got != want {
			t.Errorf("modelsURL(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	for _, in := range []string{"", "not a url", "https://example.com/decisions"} {
		if _, err := modelsURL(in); err == nil {
			t.Errorf("modelsURL(%q) accepted", in)
		}
	}
}
