package memories

import (
	"database/sql"

	"hermaeus/internal/calendar"
	"hermaeus/internal/store"
)

type DiaryEntry struct {
	ID         int64              `json:"id"`
	Content    string             `json:"content"`
	Location   string             `json:"location"`
	Emotion    string             `json:"emotion"`
	Importance float64            `json:"importance"` // 0 to 1
	Tags       []string           `json:"tags"`
	Memory     *int64             `json:"memory,omitempty"`  // its copy as a memory; missing if SkyrimNet didn't make one
	Retired    *string            `json:"retired,omitempty"` // stored under a retired identity with this name ("" if unnamed)
	Date       *calendar.GameDate `json:"date,omitempty"`
	Written    *int64             `json:"written,omitempty"` // when SkyrimNet wrote it, Unix ms
}

type DiaryResponse struct {
	Entries []DiaryEntry `json:"entries"` // newest first
}

// Diary lists a character's diary entries, including those stored under
// their retired identities.
func Diary(db store.DB, uuid string) (DiaryResponse, error) {
	out := DiaryResponse{Entries: []DiaryEntry{}}
	p, err := loadPerson(db, uuid)
	if err != nil {
		return out, err
	}
	in, args := p.in()
	rows, err := db.Query(`select d.id, cast(d.actor_uuid as text), d.content, coalesce(d.location, ''), coalesce(d.emotion, ''),
			coalesce(d.importance_score, 0), d.tags, d.entry_date, d.creation_time,
			(select m.id from memories m where instr(m.tags, '"`+diaryTagPrefix+`' || d.id || '"') > 0 limit 1)
		from diary_entries d
		where d.actor_uuid in (`+in+`)
		order by d.id desc`, args...)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	for rows.Next() {
		var d DiaryEntry
		var actor string
		var tags sql.NullString
		var date, written sql.NullFloat64
		var memory sql.NullInt64
		if err := rows.Scan(&d.ID, &actor, &d.Content, &d.Location, &d.Emotion, &d.Importance, &tags, &date, &written, &memory); err != nil {
			return out, err
		}
		d.Tags = parseTags(tags.String)
		if memory.Valid {
			d.Memory = &memory.Int64
		}
		d.Retired = p.retired(actor)
		d.Date = p.calendar.DatePtr(date.Float64)
		d.Written = unixMillis(written)
		out.Entries = append(out.Entries, d)
	}
	return out, rows.Err()
}
