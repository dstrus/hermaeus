package memories

import (
	"strings"
	"testing"

	"hermaeus/internal/apperr"
	"hermaeus/internal/store"
	"hermaeus/internal/testenv"
)

// withEntries gives Serana memory 1 (with a search vector), diary entry 7,
// and memory 2, its copy (also with a search vector).
func (h *harness) withEntries() {
	h.t.Helper()
	s := testenv.SeranaUUID
	h.Exec(`
insert into events (event_type, event_data, game_time) values ('dialogue', '{}', 1234.5);
insert into memories (id, actor_uuid, content, tags, emotion, embedding_checksum) values
  (1, ` + s + `, 'Met the Dragonborn.', '["meeting"]', 'curious', 'abc'),
  (2, ` + s + `, 'Diary Entry: Dear diary.', '["diary_entry","diary_entry_7"]', null, 'def');
insert into memory_embeddings (memory_id, embedding, dimension) values (1, x'00', 384), (2, x'00', 384);
insert into diary_entries (id, actor_uuid, content) values (7, ` + s + `, 'Dear diary.');`)
}

// write runs fn as a save, the way the routes do.
func write[T any](h *harness, fn func(tx store.DB) (T, error)) (T, error) {
	h.t.Helper()
	return store.Write(h.store, testenv.SaveID, "test", fn)
}

func must[T any](h *harness, fn func(tx store.DB) (T, error)) T {
	h.t.Helper()
	v, err := write(h, fn)
	if err != nil {
		h.t.Fatal(err)
	}
	return v
}

func ptr[T any](v T) *T { return &v }

func TestEditingMemoryText(t *testing.T) {
	h := newHarness(t)
	h.withEntries()

	r := must(h, func(tx store.DB) (SaveResult, error) {
		return UpdateMemory(tx, 1, MemoryChanges{Emotion: ptr("wary")}, true)
	})
	if r.Reembed || h.Query(`select count(*) from memory_embeddings where memory_id = 1`) != int64(1) {
		t.Errorf("changing only the emotion should keep the search vector: %+v", r)
	}

	r = must(h, func(tx store.DB) (SaveResult, error) {
		return UpdateMemory(tx, 1, MemoryChanges{Content: ptr("Met the Dragonborn at Dimhollow."), Tags: &[]string{"meeting", "crypt"}}, true)
	})
	if !r.Reembed {
		t.Errorf("reembed = false after changing the text")
	}
	if n := h.Query(`select count(*) from memory_embeddings where memory_id = 1`); n != int64(0) {
		t.Errorf("search vectors = %v, want 0", n)
	}
	if c := h.Query(`select embedding_checksum from memories where id = 1`); c != nil {
		t.Errorf("embedding_checksum = %v, want NULL", c)
	}
	if u := h.Query(`select updated_at from memories where id = 1`).(float64); u < 1e9 {
		t.Errorf("updated_at = %v, want Unix seconds", u)
	}
	if tags := h.Query(`select tags from memories where id = 1`); tags != `["meeting","crypt"]` {
		t.Errorf("tags = %v", tags)
	}

	must(h, func(tx store.DB) (SaveResult, error) { return UpdateMemory(tx, 1, MemoryChanges{Title: ptr("")}, true) })
	if d := h.Query(`select display_name from memories where id = 1`); d != nil {
		t.Errorf("an empty title is stored as %v, want NULL", d)
	}
}

func TestEditingDiaryCopyUpdatesDiary(t *testing.T) {
	h := newHarness(t)
	h.withEntries()

	r := must(h, func(tx store.DB) (SaveResult, error) {
		return UpdateMemory(tx, 2, MemoryChanges{Content: ptr("Diary Entry: Rewritten.")}, true)
	})
	if r.Diary == nil || *r.Diary != 7 || h.Query(`select content from diary_entries where id = 7`) != "Rewritten." {
		t.Errorf("diary not kept in step: %+v", r)
	}

	r = must(h, func(tx store.DB) (SaveResult, error) {
		return UpdateMemory(tx, 2, MemoryChanges{Content: ptr("No prefix.")}, true)
	})
	if r.Warning == "" || h.Query(`select content from diary_entries where id = 7`) != "Rewritten." {
		t.Errorf("the diary should be left alone, with a warning, when the prefix is gone: %+v", r)
	}
}

func TestEditingDiaryUpdatesItsCopy(t *testing.T) {
	h := newHarness(t)
	h.withEntries()
	r := must(h, func(tx store.DB) (SaveResult, error) {
		return UpdateDiary(tx, 7, DiaryChanges{Content: ptr("New words."), Emotion: ptr("sad"), Tags: &[]string{"mine"}})
	})
	if r.Memory == nil || *r.Memory != 2 || !r.Reembed {
		t.Errorf("result = %+v", r)
	}
	if c := h.Query(`select content from memories where id = 2`); c != "Diary Entry: New words." {
		t.Errorf("copy's text = %v", c)
	}
	if e := h.Query(`select emotion from memories where id = 2`); e != "sad" {
		t.Errorf("copy's emotion = %v", e)
	}
	if tags := h.Query(`select tags from memories where id = 2`); tags != `["diary_entry","diary_entry_7"]` {
		t.Errorf("copy's tags = %v; they mark it as a copy, so they stay", tags)
	}

	// Changing only the tags leaves the copy alone.
	r = must(h, func(tx store.DB) (SaveResult, error) {
		return UpdateDiary(tx, 7, DiaryChanges{Tags: &[]string{"other"}})
	})
	if r.Memory != nil || r.Reembed {
		t.Errorf("tags only: %+v", r)
	}
}

func TestNewDiaryEntryGetsItsCopy(t *testing.T) {
	h := newHarness(t)
	h.withEntries()
	r := must(h, func(tx store.DB) (SaveResult, error) {
		return AddDiaryEntry(tx, testenv.SeranaUUID, DiaryChanges{Content: ptr("A new day."), Importance: ptr(0.8), Emotion: ptr("calm")})
	})
	if a := h.Query(`select cast(actor_uuid as text) from diary_entries where id = ?`, r.ID); a != testenv.SeranaUUID {
		t.Errorf("diary actor = %v", a)
	}
	if g := h.Query(`select entry_date from diary_entries where id = ?`, r.ID); g != 1234.5 {
		t.Errorf("entry_date = %v, want the latest event's game time", g)
	}
	if r.Memory == nil {
		t.Fatal("no memory copy")
	}
	if c := h.Query(`select content || '|' || emotion || '|' || importance_score from memories where id = ?`, *r.Memory); c != "Diary Entry: A new day.|calm|0.8" {
		t.Errorf("copy = %v", c)
	}
	if tags := h.Query(`select tags from memories where id = ?`, *r.Memory).(string); !strings.Contains(tags, `"diary_entry_`) {
		t.Errorf("copy's tags = %v", tags)
	}
}

func TestNewMemory(t *testing.T) {
	h := newHarness(t)
	h.withEntries()
	r := must(h, func(tx store.DB) (SaveResult, error) {
		return AddMemory(tx, testenv.SeranaUUID, MemoryChanges{Content: ptr("Remembered."), Type: ptr("JOY")})
	})
	if c := h.Query(`select content || '|' || memory_type || '|' || game_time || '|' || is_active from memories where id = ?`, r.ID); c != "Remembered.|JOY|1234.5|1" {
		t.Errorf("new memory = %v", c)
	}
	for name, m := range map[string]MemoryChanges{
		"no text":    {},
		"blank text": {Content: ptr(" ")},
	} {
		if _, err := write(h, func(tx store.DB) (SaveResult, error) { return AddMemory(tx, testenv.SeranaUUID, m) }); err == nil {
			t.Errorf("%s: added", name)
		}
	}
	if _, err := write(h, func(tx store.DB) (SaveResult, error) {
		return AddMemory(tx, "0", MemoryChanges{Content: ptr("x")})
	}); err == nil {
		t.Error("added a memory to world knowledge")
	}
}

func TestDeletingDiaryEntry(t *testing.T) {
	h := newHarness(t)
	h.withEntries()
	h.Exec(`insert into diary_entries (id, actor_uuid, content) values (8, ` + testenv.SeranaUUID + `, 'Kept.');
		insert into memories (id, actor_uuid, content, tags) values (3, ` + testenv.SeranaUUID + `, 'Diary Entry: Kept.', '["diary_entry_8"]');`)

	r := must(h, func(tx store.DB) (DeletedDiary, error) { return DeleteDiaryEntries(tx, DeleteDiary{IDs: []int64{7}}) })
	if r.N != 1 || r.Memories != 1 {
		t.Errorf("result = %+v", r)
	}
	if n := h.Query(`select count(*) from memories where id = 2`); n != int64(0) {
		t.Error("its copy is still there")
	}
	if n := h.Query(`select count(*) from memory_embeddings where memory_id = 2`); n != int64(0) {
		t.Error("its copy's search vector is still there")
	}

	r = must(h, func(tx store.DB) (DeletedDiary, error) {
		return DeleteDiaryEntries(tx, DeleteDiary{IDs: []int64{8}, KeepMemory: true})
	})
	if r.N != 1 || r.Memories != 0 || h.Query(`select count(*) from memories where id = 3`) != int64(1) {
		t.Errorf("keeping the copy: %+v", r)
	}
}

func TestInvalidChangesAreRejected(t *testing.T) {
	h := newHarness(t)
	h.withEntries()
	for name, c := range map[string]MemoryChanges{
		"importance too high": {Importance: ptr(2.0)},
		"importance negative": {Importance: ptr(-0.1)},
		"empty text":          {Content: ptr("  ")},
		"empty type":          {Type: ptr("")},
	} {
		_, err := write(h, func(tx store.DB) (SaveResult, error) { return UpdateMemory(tx, 1, c, true) })
		if k, _ := apperr.KindOf(err); k != apperr.Invalid {
			t.Errorf("%s: err = %v", name, err)
		}
	}
	if _, err := write(h, func(tx store.DB) (SaveResult, error) {
		return UpdateMemory(tx, 999, MemoryChanges{Emotion: ptr("x")}, true)
	}); err == nil {
		t.Error("changed a memory that doesn't exist")
	}
	if _, err := write(h, func(tx store.DB) (SaveResult, error) { return UpdateDiary(tx, 999, DiaryChanges{Emotion: ptr("x")}) }); err == nil {
		t.Error("changed a diary entry that doesn't exist")
	}

	// A bulk change that fails part-way changes nothing.
	_, err := write(h, func(tx store.DB) (Count, error) {
		return UpdateMemories(tx, BulkChanges{IDs: []int64{1, 999}, Changes: MemoryChanges{Active: ptr(false)}})
	})
	if k, _ := apperr.KindOf(err); k != apperr.NotFound {
		t.Errorf("bulk change: err = %v", err)
	}
	if a := h.Query(`select is_active from memories where id = 1`); a != int64(1) {
		t.Errorf("is_active = %v after a failed bulk change, want 1", a)
	}
}

func TestMoveToCurrentIdentity(t *testing.T) {
	h := newHarness(t)
	h.withEntries()
	h.WithAlias()
	h.Exec(`insert into memories (id, actor_uuid, content, tags) values (50, ` + testenv.MysteriousUUID + `, 'I awaken.', '[]')`)
	r := must(h, func(tx store.DB) (Count, error) { return MoveToCurrentIdentity(tx, []int64{50, 1}) })
	if r.N != 1 {
		t.Errorf("moved %d, want only the old identity's memory", r.N)
	}
	if u := h.Query(`select cast(actor_uuid as text) from memories where id = 50`); u != testenv.SeranaUUID {
		t.Errorf("memory 50 belongs to %v", u)
	}
	if _, err := write(h, func(tx store.DB) (Count, error) { return MoveToCurrentIdentity(tx, []int64{999}) }); err == nil {
		t.Error("moved a memory that doesn't exist")
	}
}
