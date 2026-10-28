package bios

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"hermaeus/internal/apperr"
	"hermaeus/internal/install"
	"hermaeus/internal/layers"
)

// The catalog is every character SkyrimNet or a plugin ships a bio for,
// met or not: the prompts/characters folder of every layer, the highest
// layer's copy counting. Save bios aren't in it; they belong to a
// playthrough.

type CatalogEntry struct {
	Template string `json:"template"`
	Layer    int    `json:"layer"`    // index into Catalog.Layers: the bio in use
	Modified int64  `json:"modified"` // Unix ms
}

type Catalog struct {
	Layers []layers.Layer `json:"layers"`
	Bios   []CatalogEntry `json:"bios"`
}

// CatalogOf lists every template with a default bio.
func (b *Bios) CatalogOf() (Catalog, error) {
	lay := b.library.Layers()
	if len(lay) == 0 {
		return Catalog{}, apperr.New(apperr.Unavailable, "No SkyrimNet folder chosen yet")
	}
	out := Catalog{Layers: lay, Bios: []CatalogEntry{}}
	seen := map[string]bool{}
	for i, l := range lay {
		dir := filepath.Join(l.Root, "prompts", "characters")
		for _, f := range install.FilesWithSuffix(dir, ".prompt") {
			t := strings.TrimSuffix(f, f[len(f)-len(".prompt"):])
			key := strings.ToLower(t)
			if seen[key] || !templateRe.MatchString(t) {
				continue
			}
			info, err := os.Stat(filepath.Join(dir, f))
			if err != nil {
				continue
			}
			seen[key] = true
			out.Bios = append(out.Bios, CatalogEntry{Template: t, Layer: i, Modified: info.ModTime().UnixMilli()})
		}
	}
	sort.Slice(out.Bios, func(i, j int) bool {
		return strings.ToLower(out.Bios[i].Template) < strings.ToLower(out.Bios[j].Template)
	})
	return out, nil
}
