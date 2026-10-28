package bios

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"hermaeus/internal/apperr"
	"hermaeus/internal/store"
	"hermaeus/internal/textfile"
)

// A save bio's earlier versions come from two places:
//
//   - SkyrimNet's dashboard and its bio updates copy a bio aside before
//     replacing it, as overlay/_history/prompts/characters/<file>.backup.<Unix
//     seconds>, the time it was replaced. The name doesn't say which
//     playthrough the bio belonged to, so each copy is matched to one: the
//     playthrough whose bio file was last written that same second, or else
//     the one being played at the time (from its events).
//   - Hermaeus backs up a bio before each save, as
//     <backups>/bios/<save id>/<file>.<local time>.bak.
//
// Restoring a version saves its text as the bio, after backing up the
// current one, so it can be undone the same way.

// Earlier is an earlier version of a bio.
type Earlier struct {
	ID       string `json:"id"`
	Source   string `json:"source"`   // "skyrimnet" or "hermaeus"
	Replaced int64  `json:"replaced"` // when this version stopped being the bio, Unix ms
	// Which playthrough it belonged to: "this", "other" or "unclear".
	Playthrough string `json:"playthrough"`
	Save        string `json:"save,omitempty"` // for "other": which one
	Raw         string `json:"raw"`
	Current     bool   `json:"current"` // the same text as the bio now
}

type HistoryResponse struct {
	Earlier []Earlier `json:"earlier"` // newest first
}

type Restore struct {
	ID       string `json:"id"`
	Modified int64  `json:"modified"` // the current bio's time when loaded, as in Edit
}

var (
	snHistoryRe = regexp.MustCompile(`\.backup\.(\d+)$`)
	hmBackupRe  = regexp.MustCompile(`\.(\d{8}-\d{6}\.\d{3})(?:-\d+)?\.bak$`)
)

// How close a history copy's time must be to a bio file's to count as the
// same write, and how far outside a playthrough's events it may fall.
const (
	sameWrite     = 2 * time.Second
	playthroughIn = time.Hour
)

// History lists a save bio's earlier versions, newest first.
func (b *Bios) History(playthrough, template, kind string) (HistoryResponse, error) {
	out := HistoryResponse{Earlier: []Earlier{}}
	relPath, err := file("", template, kind)
	if err != nil {
		return out, err
	}
	rel := filepath.ToSlash(relPath)
	dir, err := b.dir(playthrough)
	if err != nil {
		return out, err
	}
	current := ""
	if c, err := os.ReadFile(filepath.Join(dir, relPath)); err == nil {
		current = textfile.Normalize(string(c))
	}
	add := func(e Earlier, path string) {
		c, err := os.ReadFile(path)
		if err != nil {
			return
		}
		e.Raw = textfile.Normalize(string(c))
		e.Current = e.Raw == current
		out.Earlier = append(out.Earlier, e)
	}

	// SkyrimNet's copies.
	sky := b.store.Dir()
	histDir := filepath.Join(sky, "overlay", "_history", "prompts", "characters", filepath.Dir(relPath))
	prefix := strings.ToLower(filepath.Base(rel)) + ".backup."
	entries, _ := os.ReadDir(histDir)
	var spans []span
	for _, e := range entries {
		m := snHistoryRe.FindStringSubmatch(e.Name())
		if e.IsDir() || m == nil || !strings.HasPrefix(strings.ToLower(e.Name()), prefix) {
			continue
		}
		secs, _ := strconv.ParseInt(m[1], 10, 64)
		at := time.Unix(secs, 0)
		if spans == nil {
			spans = b.spans()
		}
		save, how := whose(sky, relPath, at, spans)
		entry := Earlier{ID: "skyrimnet/" + e.Name(), Source: "skyrimnet", Replaced: at.UnixMilli(), Playthrough: how}
		switch {
		case save == playthrough:
			entry.Playthrough = "this"
		case save != "":
			entry.Playthrough, entry.Save = "other", save
		}
		add(entry, filepath.Join(histDir, e.Name()))
	}

	// Hermaeus's backups.
	hmDir := filepath.Join(b.backupDir(playthrough), filepath.Dir(relPath))
	entries, _ = os.ReadDir(hmDir)
	for _, e := range entries {
		m := hmBackupRe.FindStringSubmatch(e.Name())
		if e.IsDir() || m == nil || !strings.EqualFold(strings.TrimSuffix(e.Name(), m[0]), filepath.Base(rel)) {
			continue
		}
		at, err := time.ParseInLocation(textfile.BackupStamp, m[1], time.Local)
		if err != nil {
			continue
		}
		add(Earlier{ID: "hermaeus/" + e.Name(), Source: "hermaeus", Replaced: at.UnixMilli(), Playthrough: "this"}, filepath.Join(hmDir, e.Name()))
	}
	sort.SliceStable(out.Earlier, func(i, j int) bool { return out.Earlier[i].Replaced > out.Earlier[j].Replaced })
	return out, nil
}

// span is when a playthrough was played, from its events.
type span struct {
	save        string
	first, last time.Time
}

func (b *Bios) spans() []span {
	spans := []span{}
	for _, p := range b.store.Playthroughs() {
		var first, last any
		_, err := store.Read(b.store, p.ID, func(db store.DB) (bool, error) {
			if ok, err := store.HasTable(db, "events"); err != nil || !ok {
				return false, err
			}
			return true, db.QueryRow(`select min(local_time), max(local_time) from events`).Scan(&first, &last)
		})
		f, l := store.Timestamp(first), store.Timestamp(last)
		if err != nil || f == nil || l == nil {
			continue
		}
		spans = append(spans, span{save: p.ID, first: time.UnixMilli(*f), last: time.UnixMilli(*l)})
	}
	return spans
}

// whose works out which playthrough a SkyrimNet history copy, replaced at
// the given time, came from. how is "this" when it's sure, else "unclear".
func whose(sky, relPath string, at time.Time, spans []span) (save, how string) {
	saves, _ := os.ReadDir(filepath.Join(sky, "saves"))
	for _, sv := range saves {
		info, err := os.Stat(filepath.Join(sky, "saves", sv.Name(), "prompts", "characters", relPath))
		if err == nil && info.ModTime().Sub(at).Abs() <= sameWrite {
			return sv.Name(), "this"
		}
	}
	found := ""
	for _, p := range spans {
		if !at.Before(p.first.Add(-playthroughIn)) && !at.After(p.last.Add(playthroughIn)) {
			if found != "" {
				return "", "unclear" // played more than one at the time
			}
			found = p.save
		}
	}
	if found == "" {
		return "", "unclear"
	}
	return found, "this"
}

// RestoreEarlier makes an earlier version the bio again.
func (b *Bios) RestoreEarlier(playthrough, template, kind string, r Restore) (Version, error) {
	history, err := b.History(playthrough, template, kind)
	if err != nil {
		return Version{}, err
	}
	for _, e := range history.Earlier {
		if e.ID == r.ID {
			return b.change(playthrough, template, kind, r.Modified, replaceRaw(e.Raw))
		}
	}
	return Version{}, apperr.New(apperr.NotFound, "That earlier version of the bio wasn't found")
}
