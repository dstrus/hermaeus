package bios

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"hermaeus/internal/apperr"
	"hermaeus/internal/templates"
	"hermaeus/internal/textfile"
)

// Edit is a change to a save bio. Either Sections or Raw is set. Modified
// is the file's time when the bio was loaded, so a change made since (by
// SkyrimNet, say) isn't overwritten.
type Edit struct {
	Modified int64             `json:"modified"`
	Sections map[string]string `json:"sections,omitempty"` // new text for each named section
	Raw      *string           `json:"raw,omitempty"`      // the whole file instead
}

// edit turns a bio file's current contents into its new contents, with an
// optional warning for the user.
type edit func(old string) (updated, warning string, err error)

func (e Edit) apply() (edit, error) {
	switch {
	case e.Raw != nil && len(e.Sections) > 0:
		return nil, apperr.New(apperr.Invalid, "Change the bio's sections or its whole text, not both")
	case e.Raw != nil:
		return replaceRaw(*e.Raw), nil
	case len(e.Sections) > 0:
		return func(old string) (string, string, error) {
			updated, err := replaceSections(old, e.Sections)
			return updated, "", err
		}, nil
	}
	return nil, apperr.New(apperr.Invalid, "Nothing to change")
}

// replaceRaw replaces the whole file, keeping its line-ending style. Broken
// template markup is refused; dropping sections the file had only warns,
// since a later SkyrimNet version might use different ones.
func replaceRaw(raw string) edit {
	return func(old string) (string, string, error) {
		raw = textfile.Normalize(raw)
		if err := templates.Check(raw, templates.Rules{FlatBlocks: true}); err != nil {
			return "", "", err
		}
		kept := map[string]bool{}
		for _, sec := range Parse(raw) {
			kept[sec.Name] = true
		}
		var dropped []string
		for _, sec := range Parse(old) {
			if !kept[sec.Name] {
				dropped = append(dropped, sec.Name)
			}
		}
		warning := ""
		if len(dropped) > 0 {
			warning = "The bio no longer has these sections: " + strings.Join(dropped, ", ") + "."
		}
		return textfile.MatchLineEndings(raw, old), warning, nil
	}
}

// Save applies an edit to one of a template's save bios.
func (b *Bios) Save(playthrough, template, kind string, e Edit) (Version, error) {
	fn, err := e.apply()
	if err != nil {
		return Version{}, err
	}
	return b.change(playthrough, template, kind, e.Modified, fn)
}

func (b *Bios) change(playthrough, template, kind string, modified int64, fn edit) (Version, error) {
	var v Version
	err := b.store.ChangeFiles(fmt.Sprintf("Saved %s bio %s", kind, template), func() error {
		dir, err := b.dir(playthrough)
		if err != nil {
			return err
		}
		path, err := file(dir, template, kind)
		if err != nil {
			return err
		}
		info, err := os.Stat(path)
		if os.IsNotExist(err) {
			return apperr.New(apperr.NotFound, "No %s bio for %s in this playthrough", kind, template)
		}
		if err != nil {
			return err
		}
		if info.ModTime().UnixMilli() != modified {
			return apperr.New(apperr.Conflict, "This bio changed after you opened it. Reload it to see the latest version, then make your changes again.")
		}
		old, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		updated, warning, err := fn(string(old))
		if err != nil {
			return err
		}
		if _, err := b.backup(playthrough, dir, path, old); err != nil {
			return fmt.Errorf("could not back up the bio, so nothing was changed: %w", err)
		}
		if err := textfile.WriteAtomic(path, []byte(updated), info.Mode().Perm()); err != nil {
			return err
		}
		if info, err = os.Stat(path); err != nil {
			return err
		}
		v = newVersion(kind, info.ModTime().UnixMilli(), updated)
		v.Warning = warning
		return nil
	})
	return v, err
}

// backup copies a bio's previous contents to
// <backups>/bios/<save id>/<path under characters>.<time>.bak.
func (b *Bios) backup(playthrough, dir, path string, old []byte) (string, error) {
	rel, err := filepath.Rel(dir, path)
	if err != nil {
		return "", err
	}
	return textfile.Backup(filepath.Join(b.backupDir(playthrough), rel), old, 0o644)
}

func (b *Bios) backupDir(playthrough string) string {
	return filepath.Join(b.store.BackupDir(), "bios", playthrough)
}

// Create gives the template's character a baseline bio in the playthrough,
// copied byte for byte from the bio it uses now.
func (b *Bios) Create(playthrough, template string) (Version, error) {
	var v Version
	err := b.store.ChangeFiles("Created baseline bio "+template, func() error {
		dir, err := b.dir(playthrough)
		if err != nil {
			return err
		}
		path, err := file(dir, template, KindBaseline)
		if err != nil {
			return err
		}
		def, err := b.Default(template)
		if err != nil {
			return err
		}
		if def == nil {
			return apperr.New(apperr.NotFound, "No bio for %s to start from", template)
		}
		content, err := os.ReadFile(def.path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
		f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
		if os.IsExist(err) {
			return apperr.New(apperr.Conflict, "This playthrough already has a baseline bio for %s", template)
		}
		if err != nil {
			return err
		}
		if _, err := f.Write(content); err != nil {
			f.Close()
			os.Remove(path)
			return err
		}
		if err := f.Close(); err != nil {
			os.Remove(path)
			return err
		}
		info, err := os.Stat(path)
		if err != nil {
			return err
		}
		v = newVersion(KindBaseline, info.ModTime().UnixMilli(), string(content))
		return nil
	})
	return v, err
}
