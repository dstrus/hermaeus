package memories

import (
	"fmt"
	"strings"

	"hermaeus/internal/apperr"
	"hermaeus/internal/identity"
	"hermaeus/internal/store"
)

// DiaryChanges are the fields to change in a diary entry. Fields left out
// stay as they are.
type DiaryChanges struct {
	Content    *string   `json:"content,omitempty"`
	Location   *string   `json:"location,omitempty"`
	Emotion    *string   `json:"emotion,omitempty"`
	Importance *float64  `json:"importance,omitempty"` // 0 to 1
	Tags       *[]string `json:"tags,omitempty"`
}

func (c DiaryChanges) columns() (sets []string, args []any, err error) {
	set := func(col string, v any) {
		sets = append(sets, col+" = ?")
		args = append(args, v)
	}
	if c.Content != nil {
		if strings.TrimSpace(*c.Content) == "" {
			return nil, nil, apperr.New(apperr.Invalid, "A diary entry’s text can’t be empty")
		}
		set("content", *c.Content)
	}
	if c.Location != nil {
		set("location", *c.Location)
	}
	if c.Emotion != nil {
		set("emotion", *c.Emotion)
	}
	if c.Importance != nil {
		if err := checkImportance(*c.Importance); err != nil {
			return nil, nil, err
		}
		set("importance_score", *c.Importance)
	}
	if c.Tags != nil {
		set("tags", tagsJSON(*c.Tags))
	}
	return sets, args, nil
}

// copyChanges are the changes to make to an entry's memory copy so it
// still matches. Tags stay the copy's own, since they mark it as a copy.
func (c DiaryChanges) copyChanges() MemoryChanges {
	m := MemoryChanges{Location: c.Location, Emotion: c.Emotion, Importance: c.Importance}
	if c.Content != nil {
		text := diaryPrefix + *c.Content
		m.Content = &text
	}
	return m
}

// diaryCopy finds the memory that copies a diary entry, if there is one.
func diaryCopy(tx store.DB, diaryID int64) (*int64, error) {
	var id int64
	err := tx.QueryRow(`select id from memories where instr(tags, ?) > 0 limit 1`, fmt.Sprintf(`"%s%d"`, diaryTagPrefix, diaryID)).Scan(&id)
	if store.IsNoRows(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &id, nil
}

// UpdateDiary changes a diary entry, and its memory copy to match.
func UpdateDiary(tx store.DB, id int64, c DiaryChanges) (SaveResult, error) {
	res := SaveResult{ID: id}
	sets, args, err := c.columns()
	if err != nil || len(sets) == 0 {
		return res, err
	}
	r, err := tx.Exec(`update diary_entries set `+strings.Join(sets, ", ")+`, updated_at = `+nowSQL+` where id = ?`, append(args, id)...)
	if err != nil {
		return res, err
	}
	if n, err := r.RowsAffected(); err != nil {
		return res, err
	} else if n == 0 {
		return res, apperr.New(apperr.NotFound, "Diary entry %d not found", id)
	}
	copyID, err := diaryCopy(tx, id)
	if err != nil || copyID == nil {
		return res, err
	}
	mc := c.copyChanges()
	if mc == (MemoryChanges{}) {
		return res, nil
	}
	r2, err := UpdateMemory(tx, *copyID, mc, false)
	if err != nil {
		return res, err
	}
	res.Memory, res.Reembed = copyID, r2.Reembed
	return res, nil
}

// AddDiaryEntry adds a diary entry to a character, with its memory copy, as
// SkyrimNet makes them, dated at the latest moment it recorded in the game.
func AddDiaryEntry(tx store.DB, uuid string, c DiaryChanges) (SaveResult, error) {
	actor, ok := identity.ParseUUID(uuid)
	if !ok {
		return SaveResult{}, apperr.New(apperr.Invalid, "Invalid character %q", uuid)
	}
	if actor == 0 {
		return SaveResult{}, apperr.New(apperr.Invalid, "World knowledge has no diary")
	}
	if c.Content == nil {
		return SaveResult{}, apperr.New(apperr.Invalid, "A diary entry’s text can’t be empty")
	}
	sets, args, err := c.columns()
	if err != nil {
		return SaveResult{}, err
	}
	t, err := latestGameTime(tx)
	if err != nil {
		return SaveResult{}, err
	}
	r, err := tx.Exec(`insert into diary_entries (actor_uuid, content, entry_date, creation_time, related_event_ids, related_memory_ids,
			importance_score, tags)
		values (?, '', ?, `+nowSQL+`, '[]', '[]', 0.5, '[]')`, actor, t)
	if err != nil {
		return SaveResult{}, err
	}
	id, err := r.LastInsertId()
	if err != nil {
		return SaveResult{}, err
	}
	if _, err := tx.Exec(`update diary_entries set `+strings.Join(sets, ", ")+` where id = ?`, append(args, id)...); err != nil {
		return SaveResult{}, err
	}

	var content, location, emotion string
	var importance float64
	if err := tx.QueryRow(`select content, coalesce(location, ''), coalesce(emotion, ''), importance_score from diary_entries where id = ?`, id).
		Scan(&content, &location, &emotion, &importance); err != nil {
		return SaveResult{}, err
	}
	r, err = tx.Exec(`insert into memories (actor_uuid, content, location, game_time, creation_time, related_event_ids, related_actors,
			emotion, importance_score, tags, memory_type, condition_expr, always_inject, is_active)
		values (?, ?, ?, ?, `+nowSQL+`, '[]', '[]', ?, ?, ?, 'EXPERIENCE', '', 0, 1)`,
		actor, diaryPrefix+content, location, t, emotion, importance, tagsJSON([]string{"diary_entry", fmt.Sprintf("%s%d", diaryTagPrefix, id)}))
	if err != nil {
		return SaveResult{}, err
	}
	copyID, err := r.LastInsertId()
	return SaveResult{ID: id, Memory: &copyID, Reembed: true}, err
}

type DeleteDiary struct {
	IDs        []int64 `json:"ids"`
	KeepMemory bool    `json:"keepMemory"` // leave each entry's memory copy, so the character still remembers it
}

type DeletedDiary struct {
	N        int64 `json:"n"`
	Memories int64 `json:"memories"` // memory copies deleted with them
}

// DeleteDiaryEntries deletes diary entries, and their memory copies unless
// asked to keep them.
func DeleteDiaryEntries(tx store.DB, d DeleteDiary) (DeletedDiary, error) {
	var out DeletedDiary
	for _, id := range d.IDs {
		if !d.KeepMemory {
			copyID, err := diaryCopy(tx, id)
			if err != nil {
				return out, err
			}
			if copyID != nil {
				n, err := deleteMemory(tx, *copyID)
				if err != nil {
					return out, err
				}
				out.Memories += n
			}
		}
		r, err := tx.Exec(`delete from diary_entries where id = ?`, id)
		if err != nil {
			return out, err
		}
		n, err := r.RowsAffected()
		if err != nil {
			return out, err
		}
		out.N += n
	}
	return out, nil
}
