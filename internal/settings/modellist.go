package settings

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"hermaeus/internal/apperr"
)

// A provider's own list of models, fetched over the Internet on request:
// never when a page loads, only when the player asks. The list comes from
// the provider's models endpoint, found beside its chat endpoint in
// Providers.yaml (…/v1/chat/completions → …/v1/models), with the provider's
// key from the same file when it has one. The key stays here; the page
// gets the list.

// ModelInfo is one model as the provider describes it.
type ModelInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name,omitempty"`
	Context int    `json:"context,omitempty"` // tokens
	// Prices per million tokens, when the provider says.
	Prompt     *float64 `json:"prompt,omitempty"`
	Completion *float64 `json:"completion,omitempty"`
}

type ModelList struct {
	Provider string      `json:"provider"`
	Fetched  int64       `json:"fetched"` // Unix ms
	Models   []ModelInfo `json:"models"`  // by id
}

// modelLists is what's been fetched this session, by provider.
type modelLists struct {
	mu     sync.Mutex
	lists  map[string]ModelList
	client *http.Client
}

// ModelList lists a provider's models, from the provider unless fetched
// already this session and not asked to refresh.
func (s *Settings) ModelList(provider string, refresh bool) (ModelList, error) {
	if !refresh {
		s.lists.mu.Lock()
		list, ok := s.lists.lists[provider]
		s.lists.mu.Unlock()
		if ok {
			return list, nil
		}
	}
	if sp, ok := signInProvider(provider); ok {
		return ModelList{}, apperr.New(apperr.Invalid, "%s is reached through SkyrimNet’s own Fateless sign-in, which Hermaeus doesn’t have, so its models can’t be listed here", sp.Name)
	}
	providers, err := s.providerDetails()
	if err != nil {
		return ModelList{}, err
	}
	var p providerDetail
	found := false
	for _, c := range providers {
		if c.ID == provider {
			p, found = c, true
			break
		}
	}
	if !found {
		return ModelList{}, apperr.New(apperr.NotFound, "There's no provider %q in Providers.yaml", provider)
	}
	if p.Style != "" && p.Style != "chat" {
		return ModelList{}, apperr.New(apperr.Invalid, "%s has no model list: its API style is %s", p.Name, p.Style)
	}
	listURL, err := modelsURL(p.Endpoint)
	if err != nil {
		return ModelList{}, apperr.New(apperr.Invalid, "Can't tell where %s lists its models: %v", p.Name, err)
	}
	models, err := s.fetchModels(listURL, p.Key)
	if err != nil {
		return ModelList{}, apperr.New(apperr.Unavailable, "Could not fetch %s's models: %v", p.Name, err)
	}
	list := ModelList{Provider: provider, Fetched: time.Now().UnixMilli(), Models: models}
	s.lists.mu.Lock()
	if s.lists.lists == nil {
		s.lists.lists = map[string]ModelList{}
	}
	s.lists.lists[provider] = list
	s.lists.mu.Unlock()
	return list, nil
}

// modelsURL finds the models endpoint beside a chat endpoint.
func modelsURL(endpoint string) (string, error) {
	u, err := url.Parse(strings.TrimSpace(endpoint))
	if err != nil || u.Scheme == "" || u.Host == "" {
		return "", fmt.Errorf("the endpoint %q isn't a URL", endpoint)
	}
	path := strings.TrimSuffix(u.Path, "/")
	for _, suffix := range []string{"/chat/completions", "/completions"} {
		if base, ok := strings.CutSuffix(path, suffix); ok {
			u.Path = base + "/models"
			u.RawQuery = ""
			return u.String(), nil
		}
	}
	return "", fmt.Errorf("the endpoint %q doesn't end in /chat/completions", endpoint)
}

// fetchModels asks the provider, reading the OpenAI-style list
// ({"data": [{"id": …}]}) that OpenRouter and its kind return.
func (s *Settings) fetchModels(listURL, key string) ([]ModelInfo, error) {
	client := s.lists.client
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, listURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Hermaeus")
	if key != "" {
		req.Header.Set("Authorization", "Bearer "+key)
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(io.LimitReader(res.Body, 32<<20))
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("the provider answered %s", res.Status)
	}
	var in struct {
		Data []struct {
			ID            string `json:"id"`
			Name          string `json:"name"`
			ContextLength any    `json:"context_length"`
			Pricing       struct {
				Prompt     any `json:"prompt"`
				Completion any `json:"completion"`
			} `json:"pricing"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &in); err != nil {
		return nil, fmt.Errorf("the answer wasn't a model list")
	}
	out := make([]ModelInfo, 0, len(in.Data))
	for _, m := range in.Data {
		if m.ID == "" {
			continue
		}
		info := ModelInfo{ID: m.ID, Name: m.Name}
		if n, ok := number(m.ContextLength); ok {
			info.Context = int(n)
		}
		// Prices come per token; per million reads as providers quote them.
		if n, ok := number(m.Pricing.Prompt); ok {
			n *= 1e6
			info.Prompt = &n
		}
		if n, ok := number(m.Pricing.Completion); ok {
			n *= 1e6
			info.Completion = &n
		}
		out = append(out, info)
	}
	sort.Slice(out, func(i, j int) bool { return strings.ToLower(out[i].ID) < strings.ToLower(out[j].ID) })
	return out, nil
}

// number reads a JSON number, or a number written as text, as OpenRouter writes prices.
func number(v any) (float64, bool) {
	switch x := v.(type) {
	case float64:
		return x, true
	case string:
		n, err := strconv.ParseFloat(strings.TrimSpace(x), 64)
		return n, err == nil
	}
	return 0, false
}

// providerDetail is a provider with what's needed to ask it: never sent to the page.
type providerDetail struct {
	Provider
	Endpoint string
	Key      string
	KeyFrom  string // the provider whose key it borrows while its own is blank
}

func (s *Settings) providerDetails() ([]providerDetail, error) {
	dir, err := s.configDir()
	if err != nil {
		return nil, err
	}
	nodes, err := providerNodes(dir)
	if err != nil {
		return nil, err
	}
	out := make([]providerDetail, 0, len(nodes))
	for _, p := range nodes {
		out = append(out, providerDetail{Provider: providerOf(p), Endpoint: ruleField(p, "endpoint"), Key: strings.TrimSpace(ruleField(p, "api_key")), KeyFrom: strings.TrimSpace(ruleField(p, "api_key_from"))})
	}
	return out, nil
}
