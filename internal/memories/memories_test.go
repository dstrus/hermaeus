package memories

import (
	"fmt"
	"testing"

	"hermaeus/internal/apperr"
	"hermaeus/internal/store"
	"hermaeus/internal/testenv"
)

type harness struct {
	*testenv.Folder
	t     *testing.T
	store *store.Store
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	h := &harness{Folder: testenv.NewFolder(t), t: t}
	h.store = store.New("", func() bool { return false })
	if _, err := h.store.SetDir(h.Dir); err != nil {
		t.Fatal(err)
	}
	return h
}

func (h *harness) memories(uuid string) []Memory {
	h.t.Helper()
	r, err := store.Read(h.store, testenv.SaveID, func(db store.DB) (MemoriesResponse, error) { return ForCharacter(db, uuid) })
	if err != nil {
		h.t.Fatal(err)
	}
	return r.Memories
}

func (h *harness) diary(uuid string) []DiaryEntry {
	h.t.Helper()
	r, err := store.Read(h.store, testenv.SaveID, func(db store.DB) (DiaryResponse, error) { return Diary(db, uuid) })
	if err != nil {
		h.t.Fatal(err)
	}
	return r.Entries
}

func TestMemories(t *testing.T) {
	h := newHarness(t)
	h.Exec(`insert into memories (id, actor_uuid, content, memory_type, emotion, importance_score, tags, location, game_time, creation_time,
			is_active, always_inject, condition_expr, display_name) values
		(1, ` + testenv.SeranaUUID + `, 'Met the Dragonborn.', 'RELATIONSHIP', 'curious', 0.8, '["meeting"]', 'Dimhollow Crypt', 28440, 1791685096.5,
			1, 0, null, null),
		(2, ` + testenv.SeranaUUID + `, 'Diary Entry: Dear diary.', 'EXPERIENCE', null, null, '["diary_entry","diary_entry_7"]', null, 0, 2461335.5,
			0, 1, 'is_night()', 'A night'),
		(3, ` + testenv.SeranaUUID + `, 'From a pack.', 'KNOWLEDGE', null, 0.5, 'not json', null, null, null, 1, 0, null, null),
		(4, 42, 'Someone else.', 'EXPERIENCE', null, 0.5, '[]', null, null, null, 1, 0, null, null);
		update memories set pack_id = 1 where id = 3;
		insert into memory_embeddings (memory_id, embedding, dimension) values (1, x'00', 384);`)

	ms := h.memories(testenv.SeranaUUID)
	if len(ms) != 2 || ms[0].ID != 2 || ms[1].ID != 1 {
		t.Fatalf("memories = %+v; want 2 then 1, without the pack entry or someone else's", ms)
	}
	m := ms[1]
	if m.Type != "RELATIONSHIP" || m.Emotion != "curious" || m.Importance != 0.8 || fmt.Sprint(m.Tags) != "[meeting]" ||
		m.Location != "Dimhollow Crypt" || !m.Active || m.AlwaysInject || m.Condition != "" || !m.Embedded || m.DiaryEntry != nil || m.Retired != nil {
		t.Errorf("memory 1 = %+v", m)
	}
	if m.Date == nil || m.Date.Time != "7:54 AM" || m.Date.Day != "Sundas, 17th of Last Seed, 4E 201" {
		t.Errorf("memory 1 date = %+v", m.Date)
	}
	if m.Written == nil || *m.Written != 1791685096500 {
		t.Errorf("memory 1 written = %v", show(m.Written))
	}
	m = ms[0]
	if m.Active || !m.AlwaysInject || m.Condition != "is_night()" || m.Title != "A night" || m.Embedded || m.Emotion != "" || m.Importance != 0 {
		t.Errorf("memory 2 = %+v", m)
	}
	if m.DiaryEntry == nil || *m.DiaryEntry != 7 {
		t.Errorf("memory 2 copies diary entry %v, want 7", m.DiaryEntry)
	}
	if m.Date != nil {
		t.Errorf("a clock of 0 has a date: %+v", m.Date)
	}
	// The Julian day 2461335.5 is midnight UTC on 2026-10-22.
	if m.Written == nil || *m.Written != 1792627200000 {
		t.Errorf("memory 2 written = %v", show(m.Written))
	}
}

func TestDiary(t *testing.T) {
	h := newHarness(t)
	h.Exec(`insert into diary_entries (id, actor_uuid, content, emotion, location, importance_score, tags, entry_date, creation_time) values
			(7, ` + testenv.SeranaUUID + `, 'Dear diary.', 'wistful', 'Castle Volkihar', 0.7, '[]', 28440, 1791685096),
			(8, ` + testenv.SeranaUUID + `, 'No copy.', null, null, null, null, null, null);
		insert into memories (id, actor_uuid, content, tags) values (2, ` + testenv.SeranaUUID + `, 'Diary Entry: Dear diary.', '["diary_entry","diary_entry_7"]'),
			(3, ` + testenv.SeranaUUID + `, 'Not a copy of 7.', '["diary_entry_70"]');`)

	ds := h.diary(testenv.SeranaUUID)
	if len(ds) != 2 || ds[0].ID != 8 || ds[1].ID != 7 {
		t.Fatalf("diary = %+v", ds)
	}
	if d := ds[1]; d.Memory == nil || *d.Memory != 2 || d.Emotion != "wistful" || d.Date == nil || d.Date.Time != "7:54 AM" || d.Written == nil {
		t.Errorf("entry 7 = %+v", d)
	}
	if d := ds[0]; d.Memory != nil || d.Date != nil || d.Written != nil || len(d.Tags) != 0 {
		t.Errorf("entry 8 = %+v", d)
	}
}

func TestRetiredIdentityIsIncludedAndFlagged(t *testing.T) {
	h := newHarness(t)
	h.WithAlias()
	h.Exec(`insert into memories (id, actor_uuid, content, tags) values
			(1, ` + testenv.SeranaUUID + `, 'Now.', '[]'), (50, ` + testenv.MysteriousUUID + `, 'I awaken in Dimhollow Crypt.', '[]');
		insert into diary_entries (id, actor_uuid, content) values (9, ` + testenv.MysteriousUUID + `, 'Who am I?');`)
	ms := h.memories(testenv.SeranaUUID)
	if len(ms) != 2 || ms[1].Retired != nil || ms[0].Retired == nil || *ms[0].Retired != "Mysterious Woman" {
		t.Errorf("memories = %+v", ms)
	}
	if ds := h.diary(testenv.SeranaUUID); len(ds) != 1 || ds[0].Retired == nil || *ds[0].Retired != "Mysterious Woman" {
		t.Errorf("diary = %+v", ds)
	}
}

func TestInvalidCharacter(t *testing.T) {
	h := newHarness(t)
	_, err := store.Read(h.store, testenv.SaveID, func(db store.DB) (MemoriesResponse, error) { return ForCharacter(db, "serana") })
	if k, _ := apperr.KindOf(err); k != apperr.Invalid {
		t.Errorf("err = %v", err)
	}
}

// show prints what a pointer points to, or <nil>.
func show[T any](p *T) string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprint(*p)
}
