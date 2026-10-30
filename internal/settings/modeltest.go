package settings

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"hermaeus/internal/apperr"
)

// A test of a task's model: one prompt sent, on request, to one of the
// task's slots, as an OpenAI-style chat request with the slot's model and
// the task's sampling settings, using the provider's key from
// Providers.yaml (or the key it borrows). The key stays here; the page gets
// the reply, the tokens it took, and what it cost. Only providers with the
// chat API style can be tested.

// ModelTest is a test to send: the slot, by its place in the task's
// rotation, and the prompt.
type ModelTest struct {
	Slot   int    `json:"slot"`
	System string `json:"system,omitempty"` // instructions, sent before the prompt
	Prompt string `json:"prompt"`
}

// ModelTestResult is what came back.
type ModelTestResult struct {
	Model            string   `json:"model"`              // as asked for
	Answered         string   `json:"answered,omitempty"` // the model that answered, when the provider says
	Upstream         string   `json:"upstream,omitempty"` // OpenRouter's upstream provider, when it says
	Reply            string   `json:"reply"`
	Reasoning        string   `json:"reasoning,omitempty"`
	FinishReason     string   `json:"finishReason,omitempty"`
	PromptTokens     int      `json:"promptTokens"`
	CompletionTokens int      `json:"completionTokens"`
	Cost             *float64 `json:"cost,omitempty"`     // in US dollars
	CostFrom         string   `json:"costFrom,omitempty"` // provider: as billed; prices: from its model list's prices
	Millis           int64    `json:"millis"`
	MaxTokens        int      `json:"maxTokens"` // the cap the test was sent with
}

// testMaxTokens caps a test's reply, so a test stays cheap whatever the
// task allows.
const testMaxTokens = 1024

// TestModel sends a test to a task's slot.
func (s *Settings) TestModel(taskID string, t ModelTest) (ModelTestResult, error) {
	if strings.TrimSpace(t.Prompt) == "" {
		return ModelTestResult{}, apperr.New(apperr.Invalid, "Write something to send")
	}
	m, err := s.Models()
	if err != nil {
		return ModelTestResult{}, err
	}
	var task *Task
	for i := range m.Tasks {
		if m.Tasks[i].ID == taskID {
			task = &m.Tasks[i]
		}
	}
	if task == nil {
		return ModelTestResult{}, apperr.New(apperr.NotFound, "There's no task %q", taskID)
	}
	if t.Slot < 0 || t.Slot >= len(task.Slots) {
		return ModelTestResult{}, apperr.New(apperr.Invalid, "%s has no model %d", task.Label, t.Slot+1)
	}
	slot := task.Slots[t.Slot]
	if strings.TrimSpace(slot.Model) == "" {
		return ModelTestResult{}, apperr.New(apperr.Invalid, "That slot names no model")
	}
	providers, err := s.providerDetails()
	if err != nil {
		return ModelTestResult{}, err
	}
	p, ok := findProvider(providers, slot.Provider)
	if !ok {
		return ModelTestResult{}, apperr.New(apperr.NotFound, "There's no provider %q in Providers.yaml", slot.Provider)
	}
	if p.Style != "" && p.Style != "chat" {
		return ModelTestResult{}, apperr.New(apperr.Invalid, "%s uses the %s API style, which a test can’t send to. Only chat providers can be tested.", p.Name, p.Style)
	}
	key := p.Key
	if key == "" && p.KeyFrom != "" {
		if lender, ok := findProvider(providers, p.KeyFrom); ok {
			key = lender.Key
		}
	}
	endpoint, err := url.Parse(strings.TrimSpace(p.Endpoint))
	if err != nil || (endpoint.Scheme != "http" && endpoint.Scheme != "https") || endpoint.Host == "" {
		return ModelTestResult{}, apperr.New(apperr.Invalid, "%s’s endpoint isn’t a web address", p.Name)
	}

	// The request, as SkyrimNet's chat requests are made: the slot's
	// temperature, or the task's; the task's other sampling settings.
	maxTokens := testMaxTokens
	if task.Params.MaxTokens != nil && *task.Params.MaxTokens > 0 && *task.Params.MaxTokens < maxTokens {
		maxTokens = *task.Params.MaxTokens
	}
	messages := []map[string]string{}
	if sys := strings.TrimSpace(t.System); sys != "" {
		messages = append(messages, map[string]string{"role": "system", "content": sys})
	}
	messages = append(messages, map[string]string{"role": "user", "content": t.Prompt})
	body := map[string]any{"model": slot.Model, "messages": messages, "max_tokens": maxTokens, "stream": false}
	temp := slot.Temperature
	if temp == nil {
		temp = task.Params.Temperature
	}
	if temp != nil {
		body["temperature"] = *temp
	}
	if task.Params.TopP != nil {
		body["top_p"] = *task.Params.TopP
	}
	if task.Params.TopK != nil {
		body["top_k"] = *task.Params.TopK
	}
	if task.Params.FrequencyPenalty != nil {
		body["frequency_penalty"] = *task.Params.FrequencyPenalty
	}
	if task.Params.PresencePenalty != nil {
		body["presence_penalty"] = *task.Params.PresencePenalty
	}
	if len(task.Params.StopSequences) > 0 {
		body["stop"] = task.Params.StopSequences
	}

	timeout := 60 * time.Second
	if slot.RequestTimeout != nil && *slot.RequestTimeout > 0 {
		timeout = time.Duration(*slot.RequestTimeout) * time.Second
	} else if task.Timeout != nil && *task.Timeout > 0 {
		timeout = time.Duration(*task.Timeout) * time.Second
	}
	timeout = min(timeout, 180*time.Second)

	data, _ := json.Marshal(body)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), bytes.NewReader(data))
	if err != nil {
		return ModelTestResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Hermaeus")
	if key != "" {
		req.Header.Set("Authorization", "Bearer "+key)
	}
	client := s.lists.client
	if client == nil {
		client = &http.Client{}
	}
	start := time.Now()
	res, err := client.Do(req)
	if err != nil {
		return ModelTestResult{}, apperr.New(apperr.Unavailable, "Couldn’t reach %s: %s", p.Name, hideKey(err.Error(), key))
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(res.Body, 4<<20))
	millis := time.Since(start).Milliseconds()
	if err != nil {
		return ModelTestResult{}, apperr.New(apperr.Unavailable, "%s’s answer was cut off: %s", p.Name, hideKey(err.Error(), key))
	}
	var in struct {
		Model    string `json:"model"`
		Provider string `json:"provider"`
		Choices  []struct {
			FinishReason string `json:"finish_reason"`
			Message      struct {
				Content   any    `json:"content"`
				Reasoning string `json:"reasoning"`
			} `json:"message"`
		} `json:"choices"`
		Usage struct {
			PromptTokens     int      `json:"prompt_tokens"`
			CompletionTokens int      `json:"completion_tokens"`
			Cost             *float64 `json:"cost"`
		} `json:"usage"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	parsed := json.Unmarshal(raw, &in) == nil
	if res.StatusCode != http.StatusOK || (parsed && in.Error != nil) {
		msg := res.Status
		if parsed && in.Error != nil && in.Error.Message != "" {
			msg = in.Error.Message
		}
		return ModelTestResult{}, apperr.New(apperr.Unavailable, "%s answered with an error: %s", p.Name, hideKey(msg, key))
	}
	if !parsed || len(in.Choices) == 0 {
		return ModelTestResult{}, apperr.New(apperr.Unavailable, "%s’s answer wasn’t a chat reply", p.Name)
	}
	out := ModelTestResult{
		Model:            slot.Model,
		Answered:         in.Model,
		Upstream:         in.Provider,
		Reply:            contentText(in.Choices[0].Message.Content),
		Reasoning:        in.Choices[0].Message.Reasoning,
		FinishReason:     in.Choices[0].FinishReason,
		PromptTokens:     in.Usage.PromptTokens,
		CompletionTokens: in.Usage.CompletionTokens,
		Millis:           millis,
		MaxTokens:        maxTokens,
	}
	if in.Usage.Cost != nil {
		out.Cost, out.CostFrom = in.Usage.Cost, "provider"
	} else if cost, ok := s.priceOf(slot.Provider, slot.Model, out.PromptTokens, out.CompletionTokens); ok {
		out.Cost, out.CostFrom = &cost, "prices"
	}
	return out, nil
}

func findProvider(list []providerDetail, id string) (providerDetail, bool) {
	for _, p := range list {
		if p.ID == id {
			return p, true
		}
	}
	return providerDetail{}, false
}

// contentText reads a reply's content: text, or a list of parts with text.
func contentText(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case []any:
		var b strings.Builder
		for _, part := range x {
			if m, ok := part.(map[string]any); ok {
				if t, ok := m["text"].(string); ok {
					b.WriteString(t)
				}
			}
		}
		return b.String()
	}
	return ""
}

// priceOf is what a reply cost by the model's prices, when its provider's
// model list has been fetched this session and lists them.
func (s *Settings) priceOf(provider, model string, prompt, completion int) (float64, bool) {
	s.lists.mu.Lock()
	list, ok := s.lists.lists[provider]
	s.lists.mu.Unlock()
	if !ok {
		return 0, false
	}
	for _, m := range list.Models {
		if m.ID == model && m.Prompt != nil && m.Completion != nil {
			return (float64(prompt)**m.Prompt + float64(completion)**m.Completion) / 1e6, true
		}
	}
	return 0, false
}

// hideKey keeps a key out of a message, should a provider repeat it.
func hideKey(msg, key string) string {
	if len(key) >= 8 {
		msg = strings.ReplaceAll(msg, key, "(your key)")
	}
	return fmt.Sprint(msg)
}
