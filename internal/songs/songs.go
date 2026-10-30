/*
Package songs reads and edits the playthrough's bard songs: songs a bard
composed and sang, generated as audio by SkyrimNet's music service. Each is
a row of bard_songs, with its lyrics and style, and an audio file under
data/bard_songs.
*/
package songs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"hermaeus/internal/apperr"
	"hermaeus/internal/store"
)

type Song struct {
	ID           int64   `json:"id"`
	Title        string  `json:"title"`
	Lyrics       string  `json:"lyrics"`
	Style        string  `json:"style"` // comma-separated tags the music was made from
	Voice        string  `json:"voice"` // male or female
	Composer     string  `json:"composer"`
	ComposerUUID string  `json:"composerUUID,omitempty"` // a string: UUIDs exceed JavaScript's safe integers
	Source       string  `json:"source"`                 // generated…
	Instrumental bool    `json:"instrumental"`
	Plays        int     `json:"plays"`
	Seconds      float64 `json:"seconds,omitempty"` // 0 when SkyrimNet didn't record it
	Created      string  `json:"created"`           // real time, as SQLite writes it, in UTC
	File         string  `json:"file"`              // the audio file's name
	HasAudio     bool    `json:"hasAudio"`
}

type List struct {
	Songs  []Song `json:"songs"`  // newest first
	Folder string `json:"folder"` // where the audio is, for people
}

// Columns SkyrimNet added after the table first appeared, with what an
// older database has instead.
var later = map[string]string{"composer_uuid": "0", "composer_name": "''", "source": "'generated'", "instrumental": "0"}

func columns(db store.DB) (map[string]string, error) {
	out := map[string]string{}
	for c, missing := range later {
		ok, err := store.HasColumn(db, "bard_songs", c)
		if err != nil {
			return nil, err
		}
		out[c] = c
		if !ok {
			out[c] = missing
		}
	}
	return out, nil
}

func folder(dir string) string { return filepath.Join(dir, "data", "bard_songs") }

// All lists the playthrough's songs. dir is the SkyrimNet folder.
func All(db store.DB, dir string) (List, error) {
	out := List{Songs: []Song{}, Folder: folder(dir)}
	if ok, err := store.HasTable(db, "bard_songs"); err != nil || !ok {
		return out, err
	}
	col, err := columns(db)
	if err != nil {
		return out, err
	}
	rows, err := db.Query(fmt.Sprintf(`select id, title, lyrics, style_tags, gender, coalesce(%s, ''), cast(coalesce(%s, 0) as text),
			coalesce(%s, ''), coalesce(%s, 0), coalesce(play_count, 0), coalesce(duration_seconds, 0), coalesce(created_at, ''), audio_filename
		from bard_songs order by id desc`, col["composer_name"], col["composer_uuid"], col["source"], col["instrumental"]))
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var s Song
		if err := rows.Scan(&s.ID, &s.Title, &s.Lyrics, &s.Style, &s.Voice, &s.Composer, &s.ComposerUUID,
			&s.Source, &s.Instrumental, &s.Plays, &s.Seconds, &s.Created, &s.File); err != nil {
			return out, err
		}
		if s.ComposerUUID == "0" {
			s.ComposerUUID = ""
		}
		s.HasAudio = AudioFile(dir, s.File) != ""
		out.Songs = append(out.Songs, s)
	}
	return out, rows.Err()
}

// Changes are the fields of a song that can be edited; nil leaves one as it
// is. Changing them doesn't change the recording.
type Changes struct {
	Title  *string `json:"title,omitempty"`
	Lyrics *string `json:"lyrics,omitempty"`
	Style  *string `json:"style,omitempty"`
}

// Update changes a song.
func Update(tx store.DB, id int64, c Changes) error {
	var sets []string
	var args []any
	if c.Title != nil {
		t := strings.TrimSpace(*c.Title)
		if t == "" {
			return apperr.New(apperr.Invalid, "A song needs a title")
		}
		sets, args = append(sets, "title = ?"), append(args, t)
	}
	if c.Lyrics != nil {
		l := strings.TrimSpace(strings.ReplaceAll(*c.Lyrics, "\r\n", "\n"))
		if l == "" {
			return apperr.New(apperr.Invalid, "A song’s lyrics can’t be empty")
		}
		sets, args = append(sets, "lyrics = ?"), append(args, l)
	}
	if c.Style != nil {
		sets, args = append(sets, "style_tags = ?"), append(args, strings.TrimSpace(*c.Style))
	}
	if len(sets) == 0 {
		return apperr.New(apperr.Invalid, "Nothing to change")
	}
	r, err := tx.Exec(`update bard_songs set `+strings.Join(sets, ", ")+` where id = ?`, append(args, id)...)
	if err != nil {
		return err
	}
	if n, _ := r.RowsAffected(); n == 0 {
		return apperr.New(apperr.NotFound, "Song %d not found", id)
	}
	return nil
}

// Delete removes a song from the playthrough, so bards no longer sing it.
// Its audio file stays.
func Delete(tx store.DB, id int64) error {
	r, err := tx.Exec(`delete from bard_songs where id = ?`, id)
	if err != nil {
		return err
	}
	if n, _ := r.RowsAffected(); n == 0 {
		return apperr.New(apperr.NotFound, "Song %d not found", id)
	}
	return nil
}

// AudioTypes are the audio files a song can be, by extension.
var AudioTypes = map[string]string{".mp3": "audio/mpeg", ".wav": "audio/wav", ".ogg": "audio/ogg", ".m4a": "audio/mp4", ".flac": "audio/flac"}

// AudioFile is a song's audio under data/bard_songs, refusing anything but a
// plain audio file name. "" when there's none.
func AudioFile(dir, name string) string {
	if dir == "" || name == "" || strings.ContainsAny(name, `\/:`) || strings.Contains(name, "..") {
		return ""
	}
	if AudioTypes[strings.ToLower(filepath.Ext(name))] == "" {
		return ""
	}
	f := filepath.Join(folder(dir), name)
	if info, err := os.Stat(f); err != nil || !info.Mode().IsRegular() {
		return ""
	}
	return f
}

// StoredFile is a song's audio_filename.
func StoredFile(db store.DB, id int64) (string, error) {
	var n string
	err := db.QueryRow(`select audio_filename from bard_songs where id = ?`, id).Scan(&n)
	if store.IsNoRows(err) {
		return "", apperr.New(apperr.NotFound, "Song %d not found", id)
	}
	return n, err
}
