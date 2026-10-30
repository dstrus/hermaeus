package memories

import (
	"encoding/json"
	"fmt"
	"strings"

	"hermaeus/internal/apperr"
	"hermaeus/internal/calendar"
	"hermaeus/internal/identity"
	"hermaeus/internal/store"
)

// SkyrimNet times its own changes in Unix seconds.
const nowSQL = "(julianday('now') - 2440587.5) * 86400.0"

// diaryPrefix starts the text of every memory that copies a diary entry.
const diaryPrefix = "Diary Entry: "

// MemoryChanges are the fields to change in a memory. Fields left out stay
// as they are.
type MemoryChanges struct {
	Type         *string   `json:"type,omitempty"`
	Title        *string   `json:"title,omitempty"` // "" removes it
	Content      *string   `json:"content,omitempty"`
	Location     *string   `json:"location,omitempty"`
	Emotion      *string   `json:"emotion,omitempty"`
	Importance   *float64  `json:"importance,omitempty"` // 0 to 1
	Tags         *[]string `json:"tags,omitempty"`
	Active       *bool     `json:"active,omitempty"`
	AlwaysInject *bool     `json:"alwaysInject,omitempty"`
	Condition    *string   `json:"condition,omitempty"`
}

// SaveResult says what a save did besides the change asked for.
type SaveResult struct {
	ID      int64  `json:"id"`
	Reembed bool   `json:"reembed"`           // the text changed, so SkyrimNet makes a new search vector
	Diary   *int64 `json:"diary,omitempty"`   // the diary entry also updated to match a memory
	Memory  *int64 `json:"memory,omitempty"`  // the memory also updated to match a diary entry
	Warning string `json:"warning,omitempty"` // something the user should know about what wasn't done
}

// columns turns changes into SQL assignments, checking each value.
func (c MemoryChanges) columns() (sets []string, args []any, err error) {
	set := func(col string, v any) {
		sets = append(sets, col+" = ?")
		args = append(args, v)
	}
	if c.Type != nil {
		if strings.TrimSpace(*c.Type) == "" {
			return nil, nil, apperr.New(apperr.Invalid, "A memory needs a type")
		}
		set("memory_type", *c.Type)
	}
	if c.Title != nil {
		if *c.Title == "" {
			set("display_name", nil)
		} else {
			set("display_name", *c.Title)
		}
	}
	if c.Content != nil {
		if strings.TrimSpace(*c.Content) == "" {
			return nil, nil, apperr.New(apperr.Invalid, "A memory’s text can’t be empty")
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
	if c.Active != nil {
		set("is_active", *c.Active)
	}
	if c.AlwaysInject != nil {
		set("always_inject", *c.AlwaysInject)
	}
	if c.Condition != nil {
		set("condition_expr", *c.Condition)
	}
	return sets, args, nil
}

func checkImportance(v float64) error {
	if !(v >= 0 && v <= 1) {
		return apperr.New(apperr.Invalid, "Importance must be between 0 and 1")
	}
	return nil
}

func tagsJSON(tags []string) string {
	if tags == nil {
		tags = []string{}
	}
	b, _ := json.Marshal(tags)
	return string(b)
}

// UpdateMemory changes a memory. Changing its text or title clears its search
// vector, which SkyrimNet makes again in the background. If the memory copies
// a diary entry, syncDiary keeps the entry's text in step.
func UpdateMemory(tx store.DB, id int64, c MemoryChanges, syncDiary bool) (SaveResult, error) {
	res := SaveResult{ID: id}
	var tags *string
	if err := tx.QueryRow(`select tags from memories where id = ?`, id).Scan(&tags); err != nil {
		if store.IsNoRows(err) {
			return res, apperr.New(apperr.NotFound, "Memory %d not found", id)
		}
		return res, err
	}
	sets, args, err := c.columns()
	if err != nil || len(sets) == 0 {
		return res, err
	}
	if _, err := tx.Exec(`update memories set `+strings.Join(sets, ", ")+`, updated_at = `+nowSQL+` where id = ?`, append(args, id)...); err != nil {
		return res, err
	}
	// Knowledge entries are embedded from title and text, so either change makes the vector stale.
	if c.Content != nil || c.Title != nil {
		if err := clearEmbedding(tx, id); err != nil {
			return res, err
		}
		res.Reembed = true
	}
	if c.Content == nil || !syncDiary || tags == nil {
		return res, nil
	}
	diaryID := diaryEntryOf(parseTags(*tags))
	if diaryID == nil {
		return res, nil
	}
	rest, ok := strings.CutPrefix(*c.Content, diaryPrefix)
	if !ok {
		res.Warning = fmt.Sprintf("This memory copies diary entry %d, but its text no longer starts with “%s”, so the diary entry wasn’t changed.", *diaryID, diaryPrefix)
		return res, nil
	}
	if _, err := tx.Exec(`update diary_entries set content = ?, updated_at = `+nowSQL+` where id = ?`, rest, *diaryID); err != nil {
		return res, err
	}
	res.Diary = diaryID
	return res, nil
}

// clearEmbedding drops a memory's search vector. SkyrimNet makes missing
// vectors in the background while a save is loaded.
func clearEmbedding(tx store.DB, memoryID int64) error {
	if _, err := tx.Exec(`delete from memory_embeddings where memory_id = ?`, memoryID); err != nil {
		return err
	}
	_, err := tx.Exec(`update memories set embedding_checksum = null where id = ?`, memoryID)
	return err
}

// AddMemory adds a memory to a character, dated at the latest moment
// SkyrimNet recorded in the game. Fields left out of m get SkyrimNet's usual
// defaults; it needs at least its text.
func AddMemory(tx store.DB, uuid string, m MemoryChanges) (SaveResult, error) {
	actor, ok := identity.ParseUUID(uuid)
	if !ok {
		return SaveResult{}, apperr.New(apperr.Invalid, "Invalid character %q", uuid)
	}
	if actor == 0 {
		return SaveResult{}, apperr.New(apperr.Invalid, "World knowledge is kept in knowledge packs, not as a character’s memories")
	}
	if m.Content == nil {
		return SaveResult{}, apperr.New(apperr.Invalid, "A memory’s text can’t be empty")
	}
	sets, args, err := m.columns()
	if err != nil {
		return SaveResult{}, err
	}
	t, err := latestGameTime(tx)
	if err != nil {
		return SaveResult{}, err
	}
	r, err := tx.Exec(`insert into memories (actor_uuid, content, game_time, creation_time, related_event_ids, related_actors,
			memory_type, importance_score, tags, always_inject, is_active)
		values (?, '', ?, `+nowSQL+`, '[]', '[]', 'EXPERIENCE', 0.5, '[]', 0, 1)`, actor, t)
	if err != nil {
		return SaveResult{}, err
	}
	id, err := r.LastInsertId()
	if err != nil {
		return SaveResult{}, err
	}
	if _, err := tx.Exec(`update memories set `+strings.Join(sets, ", ")+` where id = ?`, append(args, id)...); err != nil {
		return SaveResult{}, err
	}
	return SaveResult{ID: id, Reembed: true}, nil
}

// latestGameTime is the latest moment on the game clock SkyrimNet recorded,
// for dating new memories and diary entries.
func latestGameTime(tx store.DB) (float64, error) {
	var t *float64
	if ok, err := store.HasTable(tx, "events"); err != nil {
		return 0, err
	} else if ok {
		err := tx.QueryRow(`select game_time from events where game_time > 0 and game_time <= ? order by id desc limit 1`, calendar.MaxClock).Scan(&t)
		if err != nil && !store.IsNoRows(err) {
			return 0, err
		}
	}
	if t == nil {
		if err := tx.QueryRow(`select max(game_time) from memories where game_time <= ?`, calendar.MaxClock).Scan(&t); err != nil {
			return 0, err
		}
	}
	if t == nil {
		return 0, nil
	}
	return *t, nil
}

// IDs lists the memories or diary entries an action applies to.
type IDs struct {
	IDs []int64 `json:"ids"`
}

type BulkChanges struct {
	IDs     []int64       `json:"ids"`
	Changes MemoryChanges `json:"changes"`
}

// Count says how many things an action changed.
type Count struct {
	N int64 `json:"n"`
}

// UpdateMemories makes the same change to several memories. Diary entries
// aren't kept in step, since the change is the same for every memory.
func UpdateMemories(tx store.DB, b BulkChanges) (Count, error) {
	for _, id := range b.IDs {
		if _, err := UpdateMemory(tx, id, b.Changes, false); err != nil {
			return Count{}, err
		}
	}
	return Count{int64(len(b.IDs))}, nil
}

// DeleteMemories deletes memories and their search vectors.
func DeleteMemories(tx store.DB, ids []int64) (Count, error) {
	var n int64
	for _, id := range ids {
		c, err := deleteMemory(tx, id)
		if err != nil {
			return Count{}, err
		}
		n += c
	}
	return Count{n}, nil
}

func deleteMemory(tx store.DB, id int64) (int64, error) {
	if _, err := tx.Exec(`delete from memory_embeddings where memory_id = ?`, id); err != nil {
		return 0, err
	}
	r, err := tx.Exec(`delete from memories where id = ?`, id)
	if err != nil {
		return 0, err
	}
	return r.RowsAffected()
}

// MoveToCurrentIdentity moves memories stored under a retired identity to
// the newest one SkyrimNet recorded for that NPC, so SkyrimNet recalls them
// again. Memories already under a current identity are left alone.
func MoveToCurrentIdentity(tx store.DB, ids []int64) (Count, error) {
	a, err := identity.Load(tx)
	if err != nil {
		return Count{}, err
	}
	var moved int64
	for _, id := range ids {
		var uuid string
		err := tx.QueryRow(`select cast(actor_uuid as text) from memories where id = ?`, id).Scan(&uuid)
		if store.IsNoRows(err) {
			return Count{}, apperr.New(apperr.NotFound, "Memory %d not found", id)
		}
		if err != nil {
			return Count{}, err
		}
		to := a.Resolve(uuid)
		if to == uuid {
			continue
		}
		n, _ := identity.ParseUUID(to)
		if _, err := tx.Exec(`update memories set actor_uuid = ?, updated_at = `+nowSQL+` where id = ?`, n, id); err != nil {
			return Count{}, err
		}
		moved++
	}
	return Count{moved}, nil
}
