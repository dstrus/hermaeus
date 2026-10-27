// Package memories is about what SkyrimNet remembers for each character:
// their memories, which it draws on in conversation, and their diary entries.
// Every diary entry also has a copy as a memory ("Diary Entry: …"), tagged
// with the entry's ID.
package memories

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"hermaeus/internal/apperr"
	"hermaeus/internal/calendar"
	"hermaeus/internal/identity"
	"hermaeus/internal/store"
)

type Memory struct {
	ID           int64              `json:"id"`
	Type         string             `json:"type"`  // EXPERIENCE, RELATIONSHIP, KNOWLEDGE…
	Title        string             `json:"title"` // a display name, which few memories have
	Content      string             `json:"content"`
	Location     string             `json:"location"`
	Emotion      string             `json:"emotion"`
	Importance   float64            `json:"importance"` // 0 to 1
	Tags         []string           `json:"tags"`
	Active       bool               `json:"active"`
	AlwaysInject bool               `json:"alwaysInject"`         // included in every prompt, not only when relevant
	Condition    string             `json:"condition"`            // only used while this is true, if set
	DiaryEntry   *int64             `json:"diaryEntry,omitempty"` // the diary entry this memory is a copy of
	Embedded     bool               `json:"embedded"`             // has a search vector; SkyrimNet rebuilds missing ones when it starts
	Retired      *string            `json:"retired,omitempty"`    // stored under a retired identity with this name ("" if unnamed)
	Date         *calendar.GameDate `json:"date,omitempty"`
	Written      *int64             `json:"written,omitempty"` // when SkyrimNet wrote it, Unix ms
}

type MemoriesResponse struct {
	Memories []Memory `json:"memories"` // newest first
}

// person is everything needed to read one character's records: their UUIDs
// (current first, then retired ones), and how to name and date them.
type person struct {
	uuids    []string
	aliases  *identity.Aliases
	calendar *calendar.Calendar
}

func loadPerson(db store.DB, uuid string) (*person, error) {
	if _, ok := identity.ParseUUID(uuid); !ok {
		return nil, apperr.New(apperr.Invalid, "Invalid character %q", uuid)
	}
	a, err := identity.Load(db)
	if err != nil {
		return nil, err
	}
	cal, err := calendar.Load(db)
	if err != nil {
		return nil, err
	}
	return &person{uuids: a.All(uuid), aliases: a, calendar: cal}, nil
}

// in is the SQL list and arguments for "actor_uuid in (…)".
func (p *person) in() (string, []any) {
	args := make([]any, len(p.uuids))
	for i, u := range p.uuids {
		args[i], _ = identity.ParseUUID(u)
	}
	return strings.TrimSuffix(strings.Repeat("?,", len(args)), ","), args
}

// retired is the name of the retired identity a record is stored under, or
// nil if it's under the character's current one.
func (p *person) retired(uuid string) *string {
	if name, ok := p.aliases.Retired(uuid); ok {
		return &name
	}
	return nil
}

// ForCharacter lists a character's memories, including those stored under
// their retired identities. Knowledge-pack entries aren't anyone's own.
func ForCharacter(db store.DB, uuid string) (MemoriesResponse, error) {
	out := MemoriesResponse{Memories: []Memory{}}
	p, err := loadPerson(db, uuid)
	if err != nil {
		return out, err
	}
	in, args := p.in()
	rows, err := db.Query(`select m.id, cast(m.actor_uuid as text), coalesce(m.memory_type, ''), coalesce(m.display_name, ''),
			m.content, coalesce(m.location, ''), coalesce(m.emotion, ''), coalesce(m.importance_score, 0), m.tags,
			coalesce(m.is_active, 1), coalesce(m.always_inject, 0), coalesce(m.condition_expr, ''),
			m.game_time, m.creation_time,
			exists (select 1 from memory_embeddings e where e.memory_id = m.id)
		from memories m
		where m.actor_uuid in (`+in+`) and m.pack_id is null
		order by m.id desc`, args...)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var m Memory
		var actor string
		var tags sql.NullString
		var gameTime, written sql.NullFloat64
		if err := rows.Scan(&m.ID, &actor, &m.Type, &m.Title, &m.Content, &m.Location, &m.Emotion, &m.Importance, &tags,
			&m.Active, &m.AlwaysInject, &m.Condition, &gameTime, &written, &m.Embedded); err != nil {
			return out, err
		}
		m.Tags = parseTags(tags.String)
		m.DiaryEntry = diaryEntryOf(m.Tags)
		m.Retired = p.retired(actor)
		m.Date = p.calendar.DatePtr(gameTime.Float64)
		m.Written = unixMillis(written)
		out.Memories = append(out.Memories, m)
	}
	return out, rows.Err()
}

// parseTags reads a JSON list of tags. Anything else counts as none.
func parseTags(s string) []string {
	var raw []any
	if json.Unmarshal([]byte(s), &raw) != nil {
		return []string{}
	}
	tags := make([]string, 0, len(raw))
	for _, t := range raw {
		tags = append(tags, fmt.Sprint(t))
	}
	return tags
}

// diaryTagPrefix tags a memory that copies a diary entry: diary_entry_<id>.
const diaryTagPrefix = "diary_entry_"

func diaryEntryOf(tags []string) *int64 {
	for _, t := range tags {
		if rest, ok := strings.CutPrefix(t, diaryTagPrefix); ok {
			if id, err := strconv.ParseInt(rest, 10, 64); err == nil {
				return &id
			}
		}
	}
	return nil
}

// unixMillis reads when SkyrimNet wrote something. It stores Unix seconds,
// though the column's default would be a Julian day, so both are read.
func unixMillis(v sql.NullFloat64) *int64 {
	var ms int64
	switch t := v.Float64; {
	case !v.Valid || t <= 0:
		return nil
	case t > 1e8: // Unix seconds
		ms = int64(t * 1000)
	default: // a Julian day
		ms = int64((t - 2440587.5) * 86400000)
	}
	return &ms
}
