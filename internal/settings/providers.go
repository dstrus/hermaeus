package settings

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"hermaeus/internal/apperr"
	"hermaeus/internal/textfile"

	"go.yaml.in/yaml/v3"
)

// The providers SkyrimNet can send a task to, from Providers.yaml: each
// with an endpoint, an API style, a key, and how to talk to it. A key is
// set and cleared here but never read back: the page only learns whether
// one is filled in. The shipped Decisions provider borrows OpenRouter's key
// (api_key_from) while its own is blank.

const providersFile = "Providers"

// ProviderDetails is a provider as shown, without its key.
type ProviderDetails struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	Style             string   `json:"style"`
	HasKey            bool     `json:"hasKey"`
	Builtin           bool     `json:"builtin"`
	Endpoint          string   `json:"endpoint"`
	KeyFrom           string   `json:"keyFrom,omitempty"` // the provider whose key it borrows while its own is blank
	Streaming         bool     `json:"streaming"`         // use_sse
	StructuredOutputs bool     `json:"structuredOutputs"`
	Reasoning         bool     `json:"reasoning"`
	Settings          string   `json:"settings"` // provider_settings as JSON, "" for none
	Sorting           string   `json:"sorting"`  // provider_sorting: how OpenRouter picks among its upstreams
	CacheTTL          string   `json:"cacheTTL"`
	Tasks             []string `json:"tasks"` // the tasks with a slot on it, by id
	taskNames         []string // the same, by name, for messages
}

type ProvidersResponse struct {
	Found     bool              `json:"found"`
	Path      string            `json:"path"`
	Modified  int64             `json:"modified"`
	Providers []ProviderDetails `json:"providers"`
}

// ProviderEdit is a provider as it should be, except its key.
type ProviderEdit struct {
	Modified          int64  `json:"modified"`
	Name              string `json:"name"`
	Endpoint          string `json:"endpoint"`
	Style             string `json:"style"`
	Streaming         bool   `json:"streaming"`
	StructuredOutputs bool   `json:"structuredOutputs"`
	Reasoning         bool   `json:"reasoning"`
	Settings          string `json:"settings"`
	Sorting           string `json:"sorting"`
	CacheTTL          string `json:"cacheTTL"`
}

// KeyEdit sets a provider's key; empty clears it.
type KeyEdit struct {
	Modified int64  `json:"modified"`
	Key      string `json:"key"`
}

var (
	apiStyles    = map[string]bool{"chat": true, "decisions": true, "systemone": true}
	sortings     = map[string]bool{"": true, "default": true, "price": true, "throughput": true, "latency": true}
	cacheTTLs    = map[string]bool{"": true, "5m": true, "1h": true}
	providerIDRe = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]*$`)
)

func (s *Settings) providersPath() (string, error) {
	dir, err := s.configDir()
	if err != nil {
		return "", err
	}
	return filePath(dir, providersFile), nil
}

// Providers reads every provider.
func (s *Settings) Providers() (ProvidersResponse, error) {
	path, err := s.providersPath()
	if err != nil {
		return ProvidersResponse{}, err
	}
	out := ProvidersResponse{Providers: []ProviderDetails{}}
	rel, _ := filepath.Rel(s.store.Dir(), path)
	out.Path = filepath.ToSlash(rel)
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return out, nil
	}
	if err != nil {
		return ProvidersResponse{}, err
	}
	dir, _ := s.configDir()
	nodes, err := providerNodes(dir)
	if err != nil {
		return ProvidersResponse{}, apperr.New(apperr.Invalid, "Could not read %s: %v", filepath.Base(path), err)
	}
	out.Found = true
	out.Modified = info.ModTime().UnixMilli()
	// Which tasks use each provider, so one in use can't be removed.
	used := map[string][]Task{}
	if models, err := s.Models(); err == nil {
		for _, t := range models.Tasks {
			seen := map[string]bool{}
			for _, sl := range t.Slots {
				if !seen[sl.Provider] {
					seen[sl.Provider] = true
					used[sl.Provider] = append(used[sl.Provider], t)
				}
			}
		}
	}
	for _, n := range nodes {
		base := providerOf(n)
		p := ProviderDetails{
			ID:                base.ID,
			Name:              base.Name,
			Style:             base.Style,
			HasKey:            base.HasKey,
			Builtin:           base.Builtin,
			Endpoint:          ruleField(n, "endpoint"),
			KeyFrom:           ruleField(n, "api_key_from"),
			Streaming:         ruleField(n, "use_sse") != "false",
			StructuredOutputs: ruleField(n, "use_structured_outputs") == "true",
			Reasoning:         ruleField(n, "reasoning") == "true",
			Sorting:           ruleField(n, "provider_sorting"),
			CacheTTL:          ruleField(n, "cache_ttl"),
			Tasks:             []string{},
		}
		if p.Style == "" {
			p.Style = "chat"
		}
		if settings := findNode(n, "provider_settings"); settings != nil && settings.Kind == yaml.MappingNode && len(settings.Content) > 0 {
			var v any
			if settings.Decode(&v) == nil {
				if b, err := json.Marshal(v); err == nil {
					p.Settings = string(b)
				}
			}
		}
		for _, t := range used[p.ID] {
			p.Tasks = append(p.Tasks, t.ID)
			p.taskNames = append(p.taskNames, t.label)
		}
		out.Providers = append(out.Providers, p)
	}
	return out, nil
}

func (e *ProviderEdit) check() error {
	e.Name = strings.TrimSpace(e.Name)
	e.Endpoint = strings.TrimSpace(e.Endpoint)
	e.Settings = strings.TrimSpace(e.Settings)
	if e.Name == "" {
		return apperr.New(apperr.Invalid, "A provider needs a name")
	}
	if u, err := url.Parse(e.Endpoint); err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return apperr.New(apperr.Invalid, "The endpoint is the full URL SkyrimNet sends requests to, such as https://openrouter.ai/api/v1/chat/completions")
	}
	if !apiStyles[e.Style] {
		return apperr.New(apperr.Invalid, "The API style is chat, decisions, or systemone, not %q", e.Style)
	}
	if !sortings[e.Sorting] {
		return apperr.New(apperr.Invalid, "Unknown provider sorting %q", e.Sorting)
	}
	if !cacheTTLs[e.CacheTTL] {
		return apperr.New(apperr.Invalid, "A cached prompt is kept for 5m or 1h, not %q", e.CacheTTL)
	}
	if e.Settings != "" {
		var v map[string]any
		if err := json.Unmarshal([]byte(e.Settings), &v); err != nil {
			return apperr.New(apperr.Invalid, "The provider settings must be a JSON object: %v", err)
		}
	}
	return nil
}

// SetProvider changes a provider, leaving its key as it is.
func (s *Settings) SetProvider(id string, e ProviderEdit) (ProvidersResponse, error) {
	if err := e.check(); err != nil {
		return ProvidersResponse{}, err
	}
	return s.editProviders("Saved the provider "+id, e.Modified, func(list *yaml.Node) error {
		n := providerNode(list, id)
		if n == nil {
			return apperr.New(apperr.NotFound, "There's no provider %q", id)
		}
		writeProvider(n, e)
		return nil
	})
}

// AddProvider adds a provider, named by its id made from its name.
func (s *Settings) AddProvider(e ProviderEdit) (ProvidersResponse, error) {
	if err := e.check(); err != nil {
		return ProvidersResponse{}, err
	}
	return s.editProviders("Added the provider "+e.Name, e.Modified, func(list *yaml.Node) error {
		id := slug(e.Name)
		for i := 2; providerNode(list, id) != nil; i++ {
			id = fmt.Sprintf("%s_%d", slug(e.Name), i)
		}
		n := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map", Content: []*yaml.Node{
			scalar("id"), scalar(id),
			scalar("name"), stringNode(e.Name),
			scalar("endpoint"), stringNode(e.Endpoint),
			scalar("api_key"), {Kind: yaml.ScalarNode, Tag: "!!str", Value: "", Style: yaml.DoubleQuotedStyle},
			scalar("api_style"), scalar(e.Style),
			scalar("use_sse"), boolNode(e.Streaming),
			scalar("use_structured_outputs"), boolNode(e.StructuredOutputs),
			scalar("reasoning"), boolNode(e.Reasoning),
			scalar("provider_settings"), settingsNode(e.Settings),
			scalar("provider_sorting"), scalar(orDefault(e.Sorting, "default")),
			scalar("cache_ttl"), scalar(orDefault(e.CacheTTL, "5m")),
			scalar("builtin"), boolNode(false),
		}}
		list.Content = append(list.Content, n)
		return nil
	})
}

// RemoveProvider removes a provider the player added, when no task uses it.
func (s *Settings) RemoveProvider(id string, modified int64) (ProvidersResponse, error) {
	current, err := s.Providers()
	if err != nil {
		return ProvidersResponse{}, err
	}
	for _, p := range current.Providers {
		if p.ID != id {
			continue
		}
		if p.Builtin {
			return ProvidersResponse{}, apperr.New(apperr.Invalid, "%s came with SkyrimNet, so it stays; clear its key instead", p.Name)
		}
		if len(p.Tasks) > 0 {
			return ProvidersResponse{}, apperr.New(apperr.Conflict, "%s is still used by %s. Give those tasks another provider first.", p.Name, strings.Join(p.taskNames, ", "))
		}
	}
	return s.editProviders("Removed the provider "+id, modified, func(list *yaml.Node) error {
		for i, n := range list.Content {
			if ruleField(n, "id") == id {
				list.Content = append(list.Content[:i], list.Content[i+1:]...)
				return nil
			}
		}
		return apperr.New(apperr.NotFound, "There's no provider %q", id)
	})
}

// SetKey sets a provider's key, or clears it. The key is written and
// never read back.
func (s *Settings) SetKey(id string, e KeyEdit) (ProvidersResponse, error) {
	key := strings.TrimSpace(e.Key)
	if strings.ContainsAny(key, "\n\r\t ") {
		return ProvidersResponse{}, apperr.New(apperr.Invalid, "A key is one line with no spaces")
	}
	what := "Set the key for " + id
	if key == "" {
		what = "Cleared the key for " + id
	}
	return s.editProviders(what, e.Modified, func(list *yaml.Node) error {
		n := providerNode(list, id)
		if n == nil {
			return apperr.New(apperr.NotFound, "There's no provider %q", id)
		}
		v := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: key}
		if key == "" {
			v.Style = yaml.DoubleQuotedStyle // api_key: "" as SkyrimNet writes an empty one
		}
		setField(n, "api_key", v)
		return nil
	})
}

func writeProvider(n *yaml.Node, e ProviderEdit) {
	setField(n, "name", stringNode(e.Name))
	setField(n, "endpoint", stringNode(e.Endpoint))
	setField(n, "api_style", scalar(e.Style))
	setField(n, "use_sse", boolNode(e.Streaming))
	setField(n, "use_structured_outputs", boolNode(e.StructuredOutputs))
	setField(n, "reasoning", boolNode(e.Reasoning))
	setField(n, "provider_settings", settingsNode(e.Settings))
	if e.Sorting != "" || findNode(n, "provider_sorting") != nil {
		setField(n, "provider_sorting", scalar(orDefault(e.Sorting, "default")))
	}
	if e.CacheTTL != "" || findNode(n, "cache_ttl") != nil {
		setField(n, "cache_ttl", scalar(orDefault(e.CacheTTL, "5m")))
	}
}

// settingsNode is provider_settings from JSON: a map, or ~ for none.
func settingsNode(settings string) *yaml.Node {
	if settings == "" {
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!null", Value: "~"}
	}
	var v map[string]any
	if json.Unmarshal([]byte(settings), &v) != nil || len(v) == 0 {
		return &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!null", Value: "~"}
	}
	var n yaml.Node
	n.Encode(v)
	return &n
}

func orDefault(v, def string) string {
	if v == "" {
		return def
	}
	return v
}

func providerNode(list *yaml.Node, id string) *yaml.Node {
	for _, n := range list.Content {
		if n.Kind == yaml.MappingNode && ruleField(n, "id") == id {
			return n
		}
	}
	return nil
}

// slug makes a provider id from its name: "My Local LLM" → my_local_llm.
func slug(name string) string {
	var b strings.Builder
	last := '_'
	for _, r := range strings.ToLower(name) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			last = r
		default:
			if last != '_' {
				b.WriteRune('_')
				last = '_'
			}
		}
	}
	id := strings.Trim(b.String(), "_")
	if !providerIDRe.MatchString(id) {
		return "provider"
	}
	return id
}

// editProviders changes Providers.yaml's providers: list, refusing the
// change if the file changed since modified, and backing it up first.
func (s *Settings) editProviders(what string, modified int64, change func(list *yaml.Node) error) (ProvidersResponse, error) {
	err := s.store.ChangeFiles(what, func() error {
		path, err := s.providersPath()
		if err != nil {
			return err
		}
		info, err := os.Stat(path)
		if os.IsNotExist(err) {
			return apperr.New(apperr.NotFound, "%s.yaml doesn't exist yet. SkyrimNet creates it the next time it starts.", providersFile)
		}
		if err != nil {
			return err
		}
		if info.ModTime().UnixMilli() != modified {
			return apperr.New(apperr.Conflict, "The providers changed after you opened them. Reload to see the latest, then make your change again.")
		}
		old, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		updated, err := rewrite(old, providersFile, func(doc *yaml.Node) error {
			list := findNode(doc, "providers")
			if list == nil || list.Kind != yaml.SequenceNode {
				list = &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq"}
				setField(doc, "providers", list)
			}
			return change(list)
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
		return ProvidersResponse{}, err
	}
	return s.Providers()
}
