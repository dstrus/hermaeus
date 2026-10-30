package settings

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"hermaeus/internal/apperr"

	"go.yaml.in/yaml/v3"
)

// The voices a cloud voice engine (Inworld, ElevenLabs) has cloned into the
// player's account, as its file's clone_cache lists them: which voice type
// each is, in what language, and when it was last used. Each entry's
// clone_key is the voice type and language ("malecommoner_en_us"), not a
// credential, but its field's name makes the settings pages treat it as
// one; this reads just those parts, so nothing else in the cache is sent.

// ClonedVoice is one voice in an engine's clone cache.
type ClonedVoice struct {
	VoiceType string `json:"voiceType"`
	Language  string `json:"language,omitempty"` // such as en-us
	LastUsed  int64  `json:"lastUsed,omitempty"` // Unix ms
}

type CloneCache struct {
	Engine    string        `json:"engine"` // the settings file
	Found     bool          `json:"found"`
	MaxVoices int           `json:"maxVoices"` // 0 for no limit
	Voices    []ClonedVoice `json:"voices"`    // most recently used first
	Unnamed   int           `json:"unnamed"`   // entries whose voice can't be read
}

// cloneEngines are the files with a clone cache.
var cloneEngines = map[string]bool{"Inworld": true, "ElevenLabs": true}

var cloneKeyRe = regexp.MustCompile(`^([a-z0-9_]+?)_([a-z]{2})_([a-z]{2})$`)

// Clones lists an engine's cloned voices.
func (s *Settings) Clones(engine string) (CloneCache, error) {
	if !cloneEngines[engine] {
		return CloneCache{}, apperr.New(apperr.NotFound, "%s keeps no clone cache", engine)
	}
	out := CloneCache{Engine: engine, Voices: []ClonedVoice{}}
	dir, err := s.configDir()
	if err != nil {
		return out, err
	}
	b, err := os.ReadFile(filepath.Join(dir, engine+".yaml"))
	if os.IsNotExist(err) {
		return out, nil
	}
	if err != nil {
		return out, err
	}
	out.Found = true
	var doc struct {
		CloneCache struct {
			MaxVoices int              `yaml:"max_voices"`
			Entries   []map[string]any `yaml:"entries"`
		} `yaml:"clone_cache"`
	}
	if err := yaml.Unmarshal(b, &doc); err != nil {
		return out, apperr.New(apperr.Invalid, "%s.yaml isn’t valid YAML", engine)
	}
	out.MaxVoices = doc.CloneCache.MaxVoices
	for _, e := range doc.CloneCache.Entries {
		key, _ := e["clone_key"].(string)
		m := cloneKeyRe.FindStringSubmatch(strings.ToLower(strings.TrimSpace(key)))
		if m == nil {
			out.Unnamed++
			continue
		}
		v := ClonedVoice{VoiceType: m[1], Language: m[2] + "-" + m[3]}
		if t, ok := toNumber(e["last_used_epoch"]); ok && t > 0 {
			v.LastUsed = int64(t * 1000)
		}
		out.Voices = append(out.Voices, v)
	}
	sort.SliceStable(out.Voices, func(i, j int) bool { return out.Voices[i].LastUsed > out.Voices[j].LastUsed })
	return out, nil
}
