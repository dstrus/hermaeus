package layers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	os.MkdirAll(filepath.Dir(path), 0o755)
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// library builds a SkyrimNet folder with dashboard edits and two installed
// plugins (one disabled, one stray), and an MO2-style mod folder holding
// SkyrimNet's defaults.
func library(t *testing.T) (sky, mod string) {
	t.Helper()
	sky = filepath.Join(t.TempDir(), "[NoDelete] SkyrimNet")
	mod = filepath.Join(t.TempDir(), "[NoDelete] SkyrimNet Beta27")
	writeFile(t, filepath.Join(mod, "SKSE", "Plugins", "SkyrimNet", "library", BasePlugin, "prompts", "characters", "lydia_A2C.prompt"), "base")
	writeFile(t, filepath.Join(mod, "SKSE", "Plugins", "SkyrimNet", "library", BasePlugin, "manifest.json"), `{"version": "0.25.3"}`)
	writeFile(t, filepath.Join(sky, "library", "bios.pack", "prompts", "characters", "lydia_A2C.prompt"), "pack")
	writeFile(t, filepath.Join(sky, "library", "bios.pack", "manifest.json"), `{"title": "Lydia Pack"}`)
	writeFile(t, filepath.Join(sky, "library", "bios.off", "prompts", "characters", "ralof_2F6.prompt"), "off")
	writeFile(t, filepath.Join(sky, "library", "bios.stray", "prompts", "characters", "aela_697.prompt"), "stray")
	writeFile(t, filepath.Join(sky, "overlay", "prompts", "characters", "aela_697.prompt"), "overlay")
	writeFile(t, filepath.Join(sky, registryFile), `{
		"plugins": {"bios.pack": {"enabled": true}, "bios.off": {"enabled": false}},
		"priority": ["bios.off", "bios.pack"]}`)
	return sky, mod
}

func TestLayersInPriorityOrder(t *testing.T) {
	sky, mod := library(t)
	l := New(func() string { return sky })
	if _, err := l.SetModDir(mod); err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, layer := range l.Layers() {
		got = append(got, layer.Kind+":"+layer.Source)
	}
	want := "[overlay:your dashboard edits pack:Lydia Pack base:SkyrimNet’s defaults]"
	if s := "[" + join(got) + "]"; s != want {
		t.Errorf("layers =\n %s\nwant\n %s (the disabled and stray plugins left out)", s, want)
	}
	if !l.HasDefaults() {
		t.Error("defaults not found through the mod folder")
	}
}

func TestSetModDirAcceptsTheModOrItsSkyrimNetFolder(t *testing.T) {
	_, mod := library(t)
	l := New(func() string { return "" })
	inner := filepath.Join(mod, "SKSE", "Plugins", "SkyrimNet")
	for _, dir := range []string{mod, inner} {
		if got, err := l.SetModDir(dir); err != nil || got != inner {
			t.Errorf("SetModDir(%s) = %q, %v; want %q", dir, got, err, inner)
		}
	}
	if _, err := l.SetModDir(t.TempDir()); err == nil {
		t.Error("accepted a folder without SkyrimNet's defaults")
	}
	if got, err := l.SetModDir(""); err != nil || got != "" {
		t.Errorf("SetModDir(\"\") = %q, %v", got, err)
	}
}

func TestModFolderFoundNextToOverwrite(t *testing.T) {
	root := t.TempDir()
	overwrite := filepath.Join(root, "overwrite", "SKSE", "Plugins", "SkyrimNet")
	os.MkdirAll(overwrite, 0o755)
	for name, version := range map[string]string{"SkyrimNet Beta26": "0.24.9", "[NoDelete] SkyrimNet Beta27": "0.25.3", "Other": ""} {
		lib := filepath.Join(root, "mods", name, "SKSE", "Plugins", "SkyrimNet", "library", BasePlugin)
		if name == "Other" {
			lib = filepath.Join(root, "mods", name, "SKSE")
		}
		writeFile(t, filepath.Join(lib, "manifest.json"), `{"version": "`+version+`"}`)
	}
	want := filepath.Join(root, "mods", "[NoDelete] SkyrimNet Beta27", "SKSE", "Plugins", "SkyrimNet")
	l := New(func() string { return overwrite })
	if got, detected := l.ModDir(); got != want || !detected {
		t.Errorf("ModDir = %q, %v; want %q, detected", got, detected, want)
	}
	if got := detectModDir(filepath.Join(root, "somewhere", "else")); got != "" {
		t.Errorf("found a mod folder outside MO2: %q", got)
	}
}

func join(s []string) string { return strings.Join(s, " ") }

// Under MO2 a plugin can come in its own mod, as SKSE/Plugins/SkyrimNet/
// external/<id>; SkyrimNet sees every mod's folder merged, so Hermaeus
// looks in each, or only in those the selected profile switches on.
func TestPluginsInOtherMods(t *testing.T) {
	root := t.TempDir()
	sky := filepath.Join(root, "overwrite", "SKSE", "Plugins", "SkyrimNet")
	writeFile(t, filepath.Join(sky, "content-registry.json"), `{"plugins": {"a.sever": {"enabled": true}, "a.off": {"enabled": true}}, "priority": ["a.sever", "a.off"]}`)
	sn := filepath.Join(root, "mods", "[NoDelete] SkyrimNet Beta28", "SKSE", "Plugins", "SkyrimNet")
	writeFile(t, filepath.Join(sn, "library", BasePlugin, "manifest.json"), `{"version": "0.28"}`)
	writeFile(t, filepath.Join(sn, "library", BasePlugin, "prompts", "x.prompt"), "x")
	writeFile(t, filepath.Join(root, "mods", "SeverActions 4.5.0", "SKSE", "Plugins", "SkyrimNet", "external", "a.sever", "manifest.json"), `{"title": "SeverActions", "version": "4.5.0"}`)
	writeFile(t, filepath.Join(root, "mods", " Switched Off", "SKSE", "Plugins", "SkyrimNet", "external", "a.off", "manifest.json"), `{"title": "Off"}`)
	writeFile(t, filepath.Join(root, "mods", "Textures", "textures", "a.dds"), "")

	l := New(func() string { return sky })
	if got := filepath.Base(l.PluginDir("a.sever")); got != "a.sever" {
		t.Fatalf("a.sever's folder = %q", l.PluginDir("a.sever"))
	}
	var ids []string
	for _, lay := range l.Layers() {
		ids = append(ids, lay.ID+"="+lay.Source)
	}
	if got := join(ids); got != "overlay=your dashboard edits a.sever=SeverActions a.off=Off skyrimnet.base=SkyrimNet’s defaults" {
		t.Errorf("layers = %s", got)
	}

	// A profile that switches a mod off hides its plugin.
	writeFile(t, filepath.Join(root, "ModOrganizer.ini"), "[General]\r\nselected_profile=@ByteArray(Default)\r\n")
	writeFile(t, filepath.Join(root, "profiles", "Default", "modlist.txt"), "# This file was automatically generated by Mod Organizer.\r\n+SeverActions 4.5.0\r\n- Switched Off\r\n+[NoDelete] SkyrimNet Beta28\r\n")
	if l.PluginDir("a.off") != "" || l.PluginDir("a.sever") == "" {
		t.Errorf("with the profile: off = %q, sever = %q", l.PluginDir("a.off"), l.PluginDir("a.sever"))
	}
	if mod, detected := l.ModDir(); !detected || filepath.Base(filepath.Dir(filepath.Dir(filepath.Dir(mod)))) != "[NoDelete] SkyrimNet Beta28" {
		t.Errorf("mod dir = %q, %v", mod, detected)
	}
	if got := l.SwitchedOffMod("a.off"); got != " Switched Off" {
		t.Errorf("a.off's switched-off mod = %q", got)
	}
	if got := l.SwitchedOffMod("a.sever"); got != "" {
		t.Errorf("a.sever's mod is on, but got %q", got)
	}

	// A profile from another layout, naming none of these mods, is ignored.
	writeFile(t, filepath.Join(root, "profiles", "Default", "modlist.txt"), "+Something Else\r\n-Another\r\n")
	if l.PluginDir("a.off") == "" || l.PluginDir("a.sever") == "" || !l.HasDefaults() {
		t.Error("a profile naming none of the mods hid them")
	}
}
