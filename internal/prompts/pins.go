package prompts

import (
	"slices"
	"strings"

	"hermaeus/internal/prefs"
)

// Pinned prompts sit at the top of the Prompts page, so the few the player
// keeps editing are a glance away. They're Hermaeus's own preference, kept
// with its other settings.

type PinsResponse struct {
	Pins []string `json:"pins"` // paths under prompts/, in the order they were pinned
}

// Pins lists the pinned prompts.
func (p *Prompts) Pins() PinsResponse {
	pins := p.prefs.Get().PinnedPrompts
	if pins == nil {
		pins = []string{}
	}
	return PinsResponse{Pins: pins}
}

// pinOrder ranks a prompt by when it was pinned; unpinned prompts rank last.
func (p *Prompts) pinOrder() func(path string) int {
	pins := p.prefs.Get().PinnedPrompts
	return func(path string) int {
		for i, pin := range pins {
			if strings.EqualFold(pin, path) {
				return i
			}
		}
		return len(pins)
	}
}

// SetPinned pins or unpins a prompt, and returns the pins as they are
// afterward. The prompt needn't exist: a pin outlives a plugin being
// removed and applies again when it's back.
func (p *Prompts) SetPinned(path string, pinned bool) (PinsResponse, error) {
	path, err := check(path)
	if err != nil {
		return PinsResponse{}, err
	}
	p.prefs.Update(func(v *prefs.Values) {
		v.PinnedPrompts = slices.DeleteFunc(v.PinnedPrompts, func(pin string) bool { return strings.EqualFold(pin, path) })
		if pinned {
			v.PinnedPrompts = append(v.PinnedPrompts, path)
		}
	})
	return p.Pins(), nil
}
