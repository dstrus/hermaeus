package prompts

import (
	"fmt"
	"os"
	"path/filepath"

	"hermaeus/internal/apperr"
	"hermaeus/internal/layers"
	"hermaeus/internal/templates"
	"hermaeus/internal/textfile"
)

// Edit is new text for the player's copy of a prompt. Modified is that
// copy's time when the page loaded it, or 0 when there wasn't one yet, so a
// change made since (by SkyrimNet's dashboard, say) isn't overwritten.
type Edit struct {
	Modified int64  `json:"modified"`
	Text     string `json:"text"`
}

// Revert removes the player's copy. Modified is as in Edit.
type Revert struct {
	Modified int64 `json:"modified"`
}

// Saved is the prompt after a save, with anything the player should check.
type Saved struct {
	Document Document `json:"document"`
	Warnings []string `json:"warnings,omitempty"`
}

// Save writes the player's copy of a prompt into overlay/. A new copy takes
// the line endings of the version it replaces; broken template markup is
// refused.
func (p *Prompts) Save(path string, e Edit) (Saved, error) {
	path, err := check(path)
	if err != nil {
		return Saved{}, err
	}
	text := textfile.Normalize(e.Text)
	if err := templates.Check(text, templates.Rules{}); err != nil {
		return Saved{}, err
	}
	err = p.store.ChangeFiles("Saved your copy of prompt "+path, func() error {
		overlay, lay, err := p.overlay()
		if err != nil {
			return err
		}
		dest := layers.Lookup(overlay, path)
		var old, from []byte
		if dest != "" {
			info, err := os.Stat(dest)
			if err != nil {
				return err
			}
			if info.ModTime().UnixMilli() != e.Modified {
				return apperr.New(apperr.Conflict, "Your copy of this prompt changed after you opened it. Reload to see the latest, then make your changes again.")
			}
			if old, err = os.ReadFile(dest); err != nil {
				return err
			}
			from = old
		} else {
			if e.Modified != 0 {
				return apperr.New(apperr.Conflict, "Your copy of this prompt was removed after you opened it. Reload to see the latest.")
			}
			dest = filepath.Join(overlay, filepath.FromSlash(path))
			for _, l := range lay[1:] {
				if found := layers.Lookup(filepath.Join(l.Root, dir), path); found != "" {
					from, _ = os.ReadFile(found)
					break
				}
			}
		}
		if old != nil {
			if err := p.backup(overlay, dest, old); err != nil {
				return fmt.Errorf("could not back up your copy, so nothing was changed: %w", err)
			}
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return err
		}
		return textfile.WriteAtomic(dest, []byte(textfile.MatchLineEndings(text, string(from))), 0o644)
	})
	if err != nil {
		return Saved{}, err
	}
	doc, err := p.Read(path)
	return Saved{Document: doc, Warnings: templates.Warnings(text, p.exists)}, err
}

// RemoveCopy removes the player's copy of a prompt, after backing it up, so
// the version below applies again. The document that comes back may have no
// versions, when only the player had the prompt.
func (p *Prompts) RemoveCopy(path string, r Revert) (Document, error) {
	path, err := check(path)
	if err != nil {
		return Document{}, err
	}
	err = p.store.ChangeFiles("Removed your copy of prompt "+path, func() error {
		overlay, _, err := p.overlay()
		if err != nil {
			return err
		}
		dest := layers.Lookup(overlay, path)
		if dest == "" {
			return apperr.New(apperr.NotFound, "You don't have your own copy of %s", path)
		}
		info, err := os.Stat(dest)
		if err != nil {
			return err
		}
		if info.ModTime().UnixMilli() != r.Modified {
			return apperr.New(apperr.Conflict, "Your copy of this prompt changed after you opened it. Reload to see the latest.")
		}
		old, err := os.ReadFile(dest)
		if err != nil {
			return err
		}
		if err := p.backup(overlay, dest, old); err != nil {
			return fmt.Errorf("could not back up your copy, so nothing was changed: %w", err)
		}
		return textfile.Remove(dest, overlay)
	})
	if err != nil {
		return Document{}, err
	}
	doc, err := p.Read(path)
	if kind, ok := apperr.KindOf(err); ok && kind == apperr.NotFound {
		lay, _ := p.layers()
		return Document{Layers: lay, Versions: []Version{}}, nil
	}
	return doc, err
}

// overlay is the player's own layer's prompts folder, with every layer.
func (p *Prompts) overlay() (string, []layers.Layer, error) {
	lay, err := p.layers()
	if err != nil {
		return "", nil, err
	}
	if lay[0].Kind != "overlay" {
		return "", nil, apperr.New(apperr.Unavailable, "No SkyrimNet folder chosen yet")
	}
	return filepath.Join(lay[0].Root, dir), lay, nil
}

// backup keeps the player's copy before it changes or goes, as
// <backups>/overlay/prompts/<path>.<time>.bak.
func (p *Prompts) backup(overlay, path string, old []byte) error {
	rel, err := filepath.Rel(overlay, path)
	if err != nil {
		return err
	}
	_, err = textfile.Backup(filepath.Join(p.store.BackupDir(), "overlay", dir, rel), old, 0o644)
	return err
}
