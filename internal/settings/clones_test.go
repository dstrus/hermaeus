package settings

import (
	"encoding/json"
	"strings"
	"testing"

	"hermaeus/internal/apperr"
)

func TestClones(t *testing.T) {
	h := newHarness(t)
	h.config("Inworld.yaml", `api_key_base64: SECRET
workspace_id: WORKSPACE
clone_cache:
  max_voices: 50
  entries:
    - clone_key: malecommoner_en_us
      last_used_epoch: 1791652414.14
    - clone_key: laurina_cedalius_en_us
      last_used_epoch: 1792898542.5
    - clone_key: SOMETHING-ELSE
      last_used_epoch: 1
`)
	c, err := h.settings.Clones("Inworld")
	if err != nil {
		t.Fatal(err)
	}
	if !c.Found || c.MaxVoices != 50 || c.Unnamed != 1 || len(c.Voices) != 2 {
		t.Fatalf("cache = %+v", c)
	}
	if v := c.Voices[0]; v.VoiceType != "laurina_cedalius" || v.Language != "en-us" || v.LastUsed != 1792898542500 {
		t.Errorf("most recent = %+v", v)
	}
	sent, _ := json.Marshal(c)
	for _, s := range []string{"SECRET", "WORKSPACE", "SOMETHING"} {
		if strings.Contains(string(sent), s) {
			t.Errorf("%s was sent", s)
		}
	}
	if c, err := h.settings.Clones("ElevenLabs"); err != nil || c.Found || len(c.Voices) != 0 {
		t.Errorf("no file: %+v, %v", c, err)
	}
	_, err = h.settings.Clones("Piper")
	kindOf(t, err, apperr.NotFound)
}
