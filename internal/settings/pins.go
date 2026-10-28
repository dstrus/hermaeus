package settings

import (
	"slices"
	"strings"

	"hermaeus/internal/apperr"
	"hermaeus/internal/prefs"
)

// Pinned settings files sit at the top of the Settings page, so the few
// the player keeps coming back to (SkyrimNet, Providers, a voice service)
// are a glance away. They're Hermaeus's own preference, kept with its
// other settings rather than in SkyrimNet's files.

type PinsResponse struct {
	Pins []string `json:"pins"` // file names, in the order they were pinned
}

// Pins lists the pinned files.
func (s *Settings) Pins() PinsResponse {
	pins := s.prefs.Get().PinnedSettings
	if pins == nil {
		pins = []string{}
	}
	return PinsResponse{Pins: pins}
}

// pinOrder ranks a file by when it was pinned; unpinned files rank last.
func (s *Settings) pinOrder() func(name string) int {
	pins := s.prefs.Get().PinnedSettings
	return func(name string) int {
		for i, p := range pins {
			if strings.EqualFold(p, name) {
				return i
			}
		}
		return len(pins)
	}
}

func (s *Settings) pinned(name string) bool {
	return slices.ContainsFunc(s.prefs.Get().PinnedSettings, func(p string) bool { return strings.EqualFold(p, name) })
}

// SetPinned pins or unpins a file, and returns the pins as they are
// afterward. A file needn't exist yet to be pinned: its defaults are listed
// either way.
func (s *Settings) SetPinned(name string, pinned bool) (PinsResponse, error) {
	if !validName(name) {
		return PinsResponse{}, apperr.New(apperr.Invalid, "Invalid settings name %q", name)
	}
	s.prefs.Update(func(v *prefs.Values) {
		v.PinnedSettings = slices.DeleteFunc(v.PinnedSettings, func(p string) bool { return strings.EqualFold(p, name) })
		if pinned {
			v.PinnedSettings = append(v.PinnedSettings, name)
		}
	})
	return s.Pins(), nil
}
