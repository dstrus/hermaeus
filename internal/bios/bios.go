// Package bios is about a character's bio: the text SkyrimNet gives the
// model to say who they are.
//
// SkyrimNet keeps save-specific bios under saves/<save id>/prompts/characters,
// named after the character's bio template (uuid_mappings.bio_template_name):
//
//	<template>.prompt                  the save's fixed bio, which SkyrimNet uses
//	                                   as the baseline when it rewrites the dynamic one
//	dynamic/<template>.dynamic.prompt  the evolving bio NPCs actually use
//
// A character without a save bio uses the highest-priority bio in SkyrimNet's
// content library (see the layers package). Each file is a series of
// {% block <section> %}...{% endblock %} sections.
package bios

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"hermaeus/internal/apperr"
	"hermaeus/internal/install"
	"hermaeus/internal/layers"
	"hermaeus/internal/store"
	"hermaeus/internal/textfile"
)

const (
	KindDynamic  = "dynamic"
	KindBaseline = "baseline"
	KindDefault  = "default" // from the content library; read-only
)

var (
	templateRe = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)
	blockRe    = regexp.MustCompile(`(?s)\{%-?\s*block\s+(\w+)\s*-?%\}(.*?)\{%-?\s*endblock\s*-?%\}`)
)

// Bios reads and changes bios for the playthroughs in a store, using a
// library for the defaults.
type Bios struct {
	store   *store.Store
	library *layers.Library
}

func New(st *store.Store, lib *layers.Library) *Bios {
	return &Bios{store: st, library: lib}
}

type Section struct {
	Name string `json:"name"`
	Text string `json:"text"`
}

// Version is one of a character's bios.
type Version struct {
	Kind     string    `json:"kind"`
	Modified int64     `json:"modified"` // the file's time, Unix ms; sent back on save so an outside change isn't overwritten
	Sections []Section `json:"sections"`
	Raw      string    `json:"raw"`               // the whole file, with \n line endings
	Warning  string    `json:"warning,omitempty"` // after a save: something the user should know
	Source   string    `json:"source,omitempty"`  // a default: which layer it came from
	Shared   string    `json:"shared,omitempty"`  // a default: the shared bio it is, such as bandit_generic
	path     string
}

type BioResponse struct {
	// The bios the character has, the one NPCs use first: the dynamic bio,
	// the save's baseline, then the default the baseline would replace.
	Versions []Version `json:"versions"`
}

// SaveBios says which kinds of save bio each template has in a playthrough.
type SaveBios struct {
	Bios map[string][]string `json:"bios"`
}

func newVersion(kind string, modified int64, content string) Version {
	return Version{Kind: kind, Modified: modified, Sections: Parse(content), Raw: textfile.Normalize(content)}
}

// dir is a playthrough's characters folder.
func (b *Bios) dir(playthrough string) (string, error) {
	save, err := b.store.SaveDir(playthrough)
	if err != nil {
		return "", err
	}
	return filepath.Join(save, "prompts", "characters"), nil
}

// file is a bio's path under a characters folder.
func file(dir, template, kind string) (string, error) {
	if !templateRe.MatchString(template) || strings.Contains(template, "..") {
		return "", apperr.New(apperr.Invalid, "Invalid bio template %q", template)
	}
	switch kind {
	case KindDynamic:
		return filepath.Join(dir, "dynamic", template+".dynamic.prompt"), nil
	case KindBaseline:
		return filepath.Join(dir, template+".prompt"), nil
	}
	return "", apperr.New(apperr.Invalid, "Unknown bio kind %q", kind)
}

// List says which kinds of save bio each template has in a playthrough.
func (b *Bios) List(playthrough string) (SaveBios, error) {
	out := SaveBios{Bios: map[string][]string{}}
	dir, err := b.dir(playthrough)
	if err != nil {
		return out, err
	}
	add := func(dir, suffix, kind string) {
		for _, f := range install.FilesWithSuffix(dir, suffix) {
			if t := f[:len(f)-len(suffix)]; templateRe.MatchString(t) {
				out.Bios[t] = append(out.Bios[t], kind)
			}
		}
	}
	add(filepath.Join(dir, "dynamic"), ".dynamic.prompt", KindDynamic)
	add(dir, ".prompt", KindBaseline)
	return out, nil
}

// Read returns a template's bios in a playthrough.
func (b *Bios) Read(playthrough, template string) (BioResponse, error) {
	out := BioResponse{Versions: []Version{}}
	dir, err := b.dir(playthrough)
	if err != nil {
		return out, err
	}
	for _, kind := range []string{KindDynamic, KindBaseline} {
		path, err := file(dir, template, kind)
		if err != nil {
			return out, err
		}
		content, info, err := read(path)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return out, err
		}
		out.Versions = append(out.Versions, newVersion(kind, info.ModTime().UnixMilli(), content))
	}
	def, err := b.Default(template)
	if err != nil {
		return out, err
	}
	if def != nil {
		out.Versions = append(out.Versions, *def)
	}
	return out, nil
}

func read(path string) (string, os.FileInfo, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", nil, err
	}
	info, err := os.Stat(path)
	return string(content), info, err
}

// Default is the bio the template's character uses when its save has none,
// or nil if no layer has one.
//
// NPCs without a bio of their own share one named after the template with
// its form-ID suffix replaced by "generic": every bandit_XXX uses
// bandit_generic. SkyrimNet's logs show the generic text being sent for
// such NPCs.
func (b *Bios) Default(template string) (*Version, error) {
	names := []string{template}
	if g := Generic(template); g != "" {
		names = append(names, g)
	}
	lay := b.library.Layers()
	for _, name := range names {
		for _, l := range lay {
			path, err := file(filepath.Join(l.Root, "prompts", "characters"), name, KindBaseline)
			if err != nil {
				return nil, err
			}
			content, info, err := read(path)
			if os.IsNotExist(err) {
				continue
			}
			if err != nil {
				return nil, err
			}
			v := newVersion(KindDefault, info.ModTime().UnixMilli(), content)
			v.Source, v.path = l.Source, path
			if l.Kind == "base" {
				v.Source = "SkyrimNet’s default bios"
			}
			if name != template {
				v.Shared = name
			}
			return &v, nil
		}
	}
	return nil, nil
}

// Generic is the shared bio name for a template like bandit_7AF, or "".
func Generic(template string) string {
	i := strings.LastIndex(template, "_")
	if i <= 0 || strings.HasSuffix(template, "_generic") {
		return ""
	}
	return template[:i] + "_generic"
}

// Parse splits a bio into its sections.
func Parse(content string) []Section {
	sections := []Section{}
	seen := map[string]bool{}
	for _, m := range blockRe.FindAllStringSubmatch(content, -1) {
		if seen[m[1]] {
			continue // SkyrimNet uses the first; so does Hermaeus
		}
		seen[m[1]] = true
		sections = append(sections, Section{Name: m[1], Text: strings.TrimSpace(textfile.Normalize(m[2]))})
	}
	return sections
}

// replaceSections rewrites the text inside the named sections and leaves
// every other byte of the file alone, including the whitespace around each
// section's text and the file's line endings.
func replaceSections(content string, changes map[string]string) (string, error) {
	done := map[string]bool{}
	var out strings.Builder
	last := 0
	for _, m := range blockRe.FindAllStringSubmatchIndex(content, -1) {
		name := content[m[2]:m[3]]
		text, ok := changes[name]
		if !ok || done[name] {
			continue
		}
		done[name] = true
		inner := content[m[4]:m[5]]
		lead := inner[:len(inner)-len(strings.TrimLeft(inner, " \t\r\n"))]
		trail := inner[len(strings.TrimRight(inner, " \t\r\n")):]
		text = textfile.MatchLineEndings(strings.TrimSpace(textfile.Normalize(text)), content)
		out.WriteString(content[last:m[4]])
		out.WriteString(lead + text + trail)
		last = m[5]
	}
	out.WriteString(content[last:])
	var missing []string
	for name := range changes {
		if !done[name] {
			missing = append(missing, name)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return "", apperr.New(apperr.Invalid, "This bio has no %s section", strings.Join(missing, ", "))
	}
	return out.String(), nil
}
