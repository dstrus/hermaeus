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

func TestModelTest(t *testing.T) {
	h := newHarness(t)
	var got map[string]any
	var auth string
	reply := `{"model": "google/gemma-4-31b-it", "provider": "DeepInfra", "choices": [{"finish_reason": "stop", "message": {"content": "Welcome, traveler."}}], "usage": {"prompt_tokens": 12, "completion_tokens": 4, "cost": 0.00002}}`
	status := http.StatusOK
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth = r.Header.Get("Authorization")
		json.NewDecoder(r.Body).Decode(&got)
		w.WriteHeader(status)
		fmt.Fprint(w, reply)
	}))
	defer srv.Close()
	h.config("Providers.yaml", "providers:\n  - id: openrouter\n    name: OpenRouter\n    endpoint: "+srv.URL+"/api/v1/chat/completions\n    api_key: sk-or-SECRET\n    api_style: chat\n  - id: borrower\n    name: Borrower\n    endpoint: "+srv.URL+"/v1/chat/completions\n    api_key: \"\"\n    api_key_from: openrouter\n    api_style: chat\n  - id: decisions\n    name: Decisions\n    endpoint: "+srv.URL+"/api/alpha/decisions\n    api_key: \"\"\n    api_style: decisions\n")
	h.config("OpenRouter.yaml", modelsYAML)

	r, err := h.settings.TestModel("default", ModelTest{Slot: 1, System: "You keep an inn.", Prompt: "Greet me."})
	if err != nil {
		t.Fatal(err)
	}
	if r.Reply != "Welcome, traveler." || r.PromptTokens != 12 || r.CompletionTokens != 4 || r.Cost == nil || *r.Cost != 0.00002 || r.CostFrom != "provider" || r.Upstream != "DeepInfra" || r.Model != "deepseek/deepseek-v3.2" {
		t.Errorf("result = %+v", r)
	}
	if auth != "Bearer sk-or-SECRET" {
		t.Errorf("sent %q, want the provider's key", auth)
	}
	// The slot's model and temperature, the task's cap, and both messages.
	msgs, _ := got["messages"].([]any)
	if got["model"] != "deepseek/deepseek-v3.2" || got["temperature"] != 0.7 || got["max_tokens"] != float64(testMaxTokens) || len(msgs) != 2 || got["stream"] != false {
		t.Errorf("request = %v", got)
	}
	if sent, _ := json.Marshal(r); strings.Contains(string(sent), "SECRET") {
		t.Error("the key was sent to the page")
	}

	// A provider that borrows another's key sends that key.
	h.config("OpenRouter.yaml", strings.Replace(modelsYAML, "    - provider_id: openrouter\n      temperature: 1\n", "    - provider_id: borrower\n      temperature: 1\n", 1))
	if _, err := h.settings.TestModel("default", ModelTest{Slot: 0, Prompt: "Hi."}); err != nil || auth != "Bearer sk-or-SECRET" {
		t.Errorf("borrowed key: %q, %v", auth, err)
	}

	// A provider's error, with the key kept out of it.
	status, reply = http.StatusUnauthorized, `{"error": {"message": "bad key sk-or-SECRET"}}`
	_, err = h.settings.TestModel("default", ModelTest{Slot: 0, Prompt: "Hi."})
	kindOf(t, err, apperr.Unavailable)
	if err == nil || strings.Contains(err.Error(), "SECRET") || !strings.Contains(err.Error(), "bad key") {
		t.Errorf("error = %v", err)
	}

	for _, c := range []struct {
		task string
		t    ModelTest
		kind apperr.Kind
	}{
		{"default", ModelTest{Slot: 0, Prompt: " "}, apperr.Invalid},
		{"default", ModelTest{Slot: 9, Prompt: "Hi."}, apperr.Invalid},
		{"nowhere", ModelTest{Slot: 0, Prompt: "Hi."}, apperr.NotFound},
	} {
		_, err := h.settings.TestModel(c.task, c.t)
		kindOf(t, err, c.kind)
	}
}

func TestModelTestCostFromPrices(t *testing.T) {
	h := newHarness(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/models") {
			fmt.Fprint(w, `{"data": [{"id": "google/gemma-4-31b-it", "pricing": {"prompt": "0.000001", "completion": "0.000002"}}]}`)
			return
		}
		fmt.Fprint(w, `{"choices": [{"message": {"content": [{"type": "text", "text": "Hail."}]}}], "usage": {"prompt_tokens": 1000, "completion_tokens": 500}}`)
	}))
	defer srv.Close()
	h.config("Providers.yaml", "providers:\n  - id: openrouter\n    name: Local\n    endpoint: "+srv.URL+"/v1/chat/completions\n    api_key: \"\"\n    api_style: chat\n")
	h.config("OpenRouter.yaml", modelsYAML)

	r, err := h.settings.TestModel("default", ModelTest{Slot: 0, Prompt: "Hi."})
	if err != nil || r.Reply != "Hail." || r.Cost != nil {
		t.Fatalf("before the prices are fetched: %+v, %v", r, err)
	}
	if _, err := h.settings.ModelList("openrouter", false); err != nil {
		t.Fatal(err)
	}
	r, err = h.settings.TestModel("default", ModelTest{Slot: 0, Prompt: "Hi."})
	if err != nil || r.Cost == nil || *r.Cost != 0.002 || r.CostFrom != "prices" {
		t.Errorf("by the prices: %+v, %v", r, err)
	}
}
