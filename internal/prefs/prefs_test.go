package prefs

import (
	"os"
	"path/filepath"
	"strings"
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
	if got := Load(path).Get(); got.SkyrimNetDir != "" || got.Theme != "" || got.ModDir != "" || len(got.PinnedSettings) != 0 {
		t.Errorf("values = %+v, want none", got)
	}
}

func TestThemeOf(t *testing.T) {
	for saved, want := range map[string]string{"": "system", "dark": "dark", "light": "light", "mgo": "system"} {
		if got := (Values{Theme: saved}).ThemeOf(); got != want {
			t.Errorf("ThemeOf(%q) = %q, want %q", saved, got, want)
		}
	}
}

func TestUnknownKeysAreKept(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	os.WriteFile(path, []byte(`{"skyrimNetDir": "/old", "skyrimNetModDir": "/mods", "showAllBios": true}`), 0o644)
	Load(path).Update(func(v *Values) { v.Theme = "dark" })
	b, _ := os.ReadFile(path)
	for _, want := range []string{`"skyrimNetModDir": "/mods"`, `"showAllBios": true`, `"theme": "dark"`, `"skyrimNetDir": "/old"`} {
		if !strings.Contains(string(b), want) {
			t.Errorf("settings.json lacks %s: %s", want, b)
		}
	}
}
