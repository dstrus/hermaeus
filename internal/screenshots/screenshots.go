/*
Package screenshots reads and edits OmniSight's screenshots: the pictures
SkyrimNet takes of places (and, in time, people and things), each with a
description a vision model wrote, which SkyrimNet gives the AI when the
player is there. They're kept in the playthrough's omnisight_screenshots
table, one per subject, with the image under screenshots/omnisight.
*/
package screenshots

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"hermaeus/internal/apperr"
	"hermaeus/internal/calendar"
	"hermaeus/internal/store"
)

// Screenshot is one subject's picture and its description.
type Screenshot struct {
	ID          int64              `json:"id"`
	SubjectType string             `json:"subjectType"` // location, actor, furniture…
	Subject     string             `json:"subject"`     // an editor ID for a place, a UUID for an actor
	Name        string             `json:"name"`
	Location    string             `json:"location"`
	Cell        string             `json:"cell"`
	Date        *calendar.GameDate `json:"date,omitempty"`
	GameTime    string             `json:"gameTime"` // as SkyrimNet wrote it, for when there's no Date
	Taken       string             `json:"taken"`    // real time, ISO 8601
	Description string             `json:"description"`
	Model       string             `json:"model"` // that wrote the description
	// Locked: OmniSight keeps the description as it is rather than writing
	// a new one.
	Locked       bool   `json:"locked"`
	NeedsRefresh bool   `json:"needsRefresh"` // to be taken again
	Manual       bool   `json:"manual"`       // taken by the player's hotkey
	Error        string `json:"error,omitempty"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	Weather      string `json:"weather"`
	TimeOfDay    string `json:"timeOfDay"`
	Indoor       bool   `json:"indoor"`
	HasImage     bool   `json:"hasImage"` // its image file was found
}

type List struct {
	Screenshots []Screenshot `json:"screenshots"` // newest first
	Folder      string       `json:"folder"`      // where the images are, for people
}

// Columns a playthrough's database may not have, with what to read in
// their place.
var later = map[string]string{"suggested_name": "''", "description_locked": "0", "manual_capture": "0"}

func columns(db store.DB) (map[string]string, error) {
	out := map[string]string{}
	for c, missing := range later {
		ok, err := store.HasColumn(db, "omnisight_screenshots", c)
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

// All lists the playthrough's screenshots. dir is the SkyrimNet folder, where
// their images are looked for.
func All(db store.DB, dir string) (List, error) {
	out := List{Screenshots: []Screenshot{}, Folder: filepath.Join(dir, "screenshots", "omnisight")}
	if ok, err := store.HasTable(db, "omnisight_screenshots"); err != nil || !ok {
		return out, err
	}
	col, err := columns(db)
	if err != nil {
		return out, err
	}
	cal, err := calendar.Load(db)
	if err != nil {
		return out, err
	}
	rows, err := db.Query(fmt.Sprintf(`select id, subject_type, subject_identifier,
			coalesce(nullif(subject_name, ''), nullif(%s, ''), subject_identifier), coalesce(location_name, ''), coalesce(location_cell, ''),
			coalesce(capture_timestamp_game, 0), coalesce(capture_time_game, ''), capture_time_local,
			coalesce(description, ''), coalesce(description_model, ''), %s, coalesce(needs_refresh, 0), %s,
			coalesce(generation_error, ''), coalesce(image_width, 0), coalesce(image_height, 0),
			coalesce(weather_type, ''), coalesce(time_of_day, ''), coalesce(is_indoor, 0), coalesce(file_path, '')
		from omnisight_screenshots order by capture_timestamp_local desc, id desc`, col["suggested_name"], col["description_locked"], col["manual_capture"]))
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var s Screenshot
		var clock float64
		var file string
		if err := rows.Scan(&s.ID, &s.SubjectType, &s.Subject, &s.Name, &s.Location, &s.Cell, &clock, &s.GameTime, &s.Taken,
			&s.Description, &s.Model, &s.Locked, &s.NeedsRefresh, &s.Manual, &s.Error, &s.Width, &s.Height,
			&s.Weather, &s.TimeOfDay, &s.Indoor, &file); err != nil {
			return out, err
		}
		s.Date = cal.DatePtr(clock)
		s.HasImage = ImageFile(dir, file) != ""
		out.Screenshots = append(out.Screenshots, s)
	}
	return out, rows.Err()
}

// Changes are the fields of a screenshot that can be edited; nil leaves one
// as it is.
type Changes struct {
	Name         *string `json:"name,omitempty"`
	Description  *string `json:"description,omitempty"`
	Locked       *bool   `json:"locked,omitempty"`
	NeedsRefresh *bool   `json:"needsRefresh,omitempty"`
}

// Update changes a screenshot.
func Update(tx store.DB, id int64, c Changes) error {
	var sets []string
	var args []any
	if c.Name != nil {
		sets, args = append(sets, "subject_name = ?"), append(args, strings.TrimSpace(*c.Name))
	}
	if c.Description != nil {
		sets, args = append(sets, "description = ?"), append(args, strings.TrimSpace(*c.Description))
	}
	if c.Locked != nil {
		ok, err := store.HasColumn(tx, "omnisight_screenshots", "description_locked")
		if err != nil {
			return err
		}
		if !ok {
			return apperr.New(apperr.Invalid, "This playthrough’s database has no way to keep a description from being rewritten")
		}
		sets, args = append(sets, "description_locked = ?"), append(args, *c.Locked)
	}
	if c.NeedsRefresh != nil {
		sets, args = append(sets, "needs_refresh = ?"), append(args, *c.NeedsRefresh)
	}
	if len(sets) == 0 {
		return apperr.New(apperr.Invalid, "Nothing to change")
	}
	r, err := tx.Exec(`update omnisight_screenshots set `+strings.Join(sets, ", ")+` where id = ?`, append(args, id)...)
	if err != nil {
		return err
	}
	if n, _ := r.RowsAffected(); n == 0 {
		return apperr.New(apperr.NotFound, "Screenshot %d not found", id)
	}
	return nil
}

// ImageTypes are the image files a screenshot can be, by extension.
var ImageTypes = map[string]string{".png": "image/png", ".jpg": "image/jpeg", ".jpeg": "image/jpeg", ".bmp": "image/bmp", ".webp": "image/webp"}

// ImageFile finds a screenshot's image from its stored file_path: a path or
// a JSON list of paths, as the game saw them, often on another machine or
// inside Mod Organizer's virtual Data folder, with \ and / mixed. Only
// images inside the SkyrimNet folder are found, or the stored path itself
// when Hermaeus runs where the game does. "" when there's none.
func ImageFile(dir, stored string) string {
	if stored == "" || dir == "" {
		return ""
	}
	var paths []string
	if err := json.Unmarshal([]byte(stored), &paths); err != nil {
		paths = []string{stored}
	}
	for _, p := range paths {
		if f := resolve(dir, p); f != "" && ImageTypes[strings.ToLower(filepath.Ext(f))] != "" {
			return f
		}
	}
	return ""
}

func resolve(dir, stored string) string {
	norm := strings.ReplaceAll(stored, `\`, "/")
	var candidates []string
	const marker = "/skse/plugins/skyrimnet/"
	if i := strings.Index(strings.ToLower(norm), marker); i >= 0 {
		candidates = append(candidates, filepath.Join(dir, filepath.FromSlash(norm[i+len(marker):])))
	}
	if base := norm[strings.LastIndex(norm, "/")+1:]; base != "" && base != ".." {
		candidates = append(candidates, filepath.Join(dir, "screenshots", "omnisight", base))
	}
	for _, c := range candidates {
		if within(dir, c) && isFile(c) {
			return c
		}
	}
	if p := filepath.FromSlash(norm); filepath.IsAbs(p) && filepath.Clean(p) == p && isFile(p) {
		return p
	}
	return ""
}

// within reports whether path stays inside dir once cleaned.
func within(dir, path string) bool {
	rel, err := filepath.Rel(dir, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}

func isFile(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.Mode().IsRegular()
}

// StoredFile is a screenshot's file_path, "" when it has none.
func StoredFile(db store.DB, id int64) (string, error) {
	var p *string
	err := db.QueryRow(`select file_path from omnisight_screenshots where id = ?`, id).Scan(&p)
	if store.IsNoRows(err) {
		return "", apperr.New(apperr.NotFound, "Screenshot %d not found", id)
	}
	if err != nil || p == nil {
		return "", err
	}
	return *p, nil
}
