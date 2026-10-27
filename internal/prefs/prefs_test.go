package prefs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestChangesAreRemembered(t *testing.T) {
	path := filepath.Join(t.TempDir(), "Hermaeus", "settings.json")
	Load(path).Update(func(v *Values) { v.SkyrimNetDir = "/games/SkyrimNet" })
	if got := Load(path).Get().SkyrimNetDir; got != "/games/SkyrimNet" {
		t.Errorf("remembered folder = %q", got)
	}
}

func TestUnreadableFileRemembersNothing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	os.WriteFile(path, []byte("not json"), 0o644)
	if got := Load(path).Get(); got != (Values{}) {
		t.Errorf("values = %+v, want none", got)
	}
}
