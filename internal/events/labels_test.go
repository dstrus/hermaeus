package events

import "testing"

func TestLabel(t *testing.T) {
	for in, want := range map[string]string{
		"follower_recruited": "Follower recruited",
		"npc_combat_start":   "NPC combat start",
		"player_stt":         "Player STT",
		"dialogue":           "Dialogue",
		"":                   "",
	} {
		if got := Label(in); got != want {
			t.Errorf("Label(%q) = %q, want %q", in, got, want)
		}
	}
}
