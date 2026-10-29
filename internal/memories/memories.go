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
	// Whose it is, in a list across characters: their current identity.
	Actor     string `json:"actor,omitempty"`
	ActorUUID string `json:"actorUuid,omitempty"`
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
	names    map[string]string // by UUID, for a list across characters
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
func (p *person) in() (string, []any) { return identity.SQLIn(p.uuids) }

// retired is the name of the retired identity a record is stored under, or
// nil if it's under the character's current one.
func (p *person) retired(uuid string) *string {
	if name, ok := p.aliases.Retired(uuid); ok {
		return &name
	}
	return nil
}

// everyone is the whole playthrough for a list across characters: no
// UUIDs of its own, but every name.
func everyone(db store.DB) (*person, error) {
	a, err := identity.Load(db)
	if err != nil {
		return nil, err
	}
	cal, err := calendar.Load(db)
	if err != nil {
		return nil, err
	}
	names := map[string]string{}
	rows, err := db.Query(`select cast(uuid as text), coalesce(actor_name, '') from uuid_mappings`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var u, n string
		if err := rows.Scan(&u, &n); err != nil {
			return nil, err
		}
		names[u] = n
	}
	return &person{aliases: a, calendar: cal, names: names}, rows.Err()
}

// owner names a record's character by their current identity, for a list
// across characters.
func (p *person) owner(actor string) (name, uuid string) {
	if p.names == nil {
		return "", ""
	}
	uuid = p.aliases.Resolve(actor)
	return p.names[uuid], uuid
}

// ForCharacter lists a character's memories, including those stored under
// their retired identities. Knowledge-pack entries aren't anyone's own.
func ForCharacter(db store.DB, uuid string) (MemoriesResponse, error) {
	p, err := loadPerson(db, uuid)
	if err != nil {
		return MemoriesResponse{Memories: []Memory{}}, err
	}
	in, args := p.in()
	return listMemories(db, p, "m.actor_uuid in ("+in+") and m.pack_id is null", args)
}

// All lists every character's memories in the playthrough.
func All(db store.DB) (MemoriesResponse, error) {
	p, err := everyone(db)
	if err != nil {
		return MemoriesResponse{Memories: []Memory{}}, err
	}
	return listMemories(db, p, "m.pack_id is null", nil)
}

func listMemories(db store.DB, p *person, where string, args []any) (MemoriesResponse, error) {
	out := MemoriesResponse{Memories: []Memory{}}
	rows, err := db.Query(`select m.id, cast(m.actor_uuid as text), coalesce(m.memory_type, ''), coalesce(m.display_name, ''),
			m.content, coalesce(m.location, ''), coalesce(m.emotion, ''), coalesce(m.importance_score, 0), m.tags,
			coalesce(m.is_active, 1), coalesce(m.always_inject, 0), coalesce(m.condition_expr, ''),
			m.game_time, m.creation_time,
			exists (select 1 from memory_embeddings e where e.memory_id = m.id)
		from memories m
		where `+where+`
		order by m.id desc`, args...)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var m Memory
		var actor string
		var tags sql.NullString
		var gameTime sql.NullFloat64
		var written any
		if err := rows.Scan(&m.ID, &actor, &m.Type, &m.Title, &m.Content, &m.Location, &m.Emotion, &m.Importance, &tags,
			&m.Active, &m.AlwaysInject, &m.Condition, &gameTime, &written, &m.Embedded); err != nil {
			return out, err
		}
		m.Tags = parseTags(tags.String)
		m.DiaryEntry = diaryEntryOf(m.Tags)
		m.Retired = p.retired(actor)
		m.Actor, m.ActorUUID = p.owner(actor)
		m.Date = p.calendar.DatePtr(gameTime.Float64)
		m.Written = store.Timestamp(written)
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
