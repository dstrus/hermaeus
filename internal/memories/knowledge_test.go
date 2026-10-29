package memories

import (
	"testing"

	"hermaeus/internal/apperr"
	"hermaeus/internal/store"
	"hermaeus/internal/testenv"
)

func TestWorldKnowledge(t *testing.T) {
	h := newHarness(t)
	h.Exec(`insert into knowledge_packs (id, name, description, version, is_enabled, source_path) values
  (1, 'Winterhold', 'The city', '1.0.1', 1, 'knowledge/winterhold.sknpack'),
  (2, 'Alternate Perspective', '', '1.0.0', 0, null);
insert into memories (id, actor_uuid, display_name, content, memory_type, pack_id, importance_score, tags, is_active, always_inject, game_time, creation_time)
values (50, 0, 'Magnar', 'Magnar is the smith.', 'KNOWLEDGE', 1, 0.8, '["smith"]', 1, 0, 100, 2461000),
       (51, 0, 'Helgen', 'Helgen stands.', 'KNOWLEDGE', 2, 0.9, '[]', 1, 1, 100, 2461000),
       (52, 0, 'Gram and Aya', 'They cleared a mine.', 'KNOWLEDGE', null, 0.5, '[]', 1, 0, 100, 2461000);`)
	k, err := store.Read(h.store, testenv.SaveID, Knowledge)
	if err != nil {
		t.Fatal(err)
	}
	if len(k.Packs) != 2 || k.Packs[0].Name != "Alternate Perspective" || k.Packs[0].Enabled || k.Packs[1].Entries != 1 {
		t.Errorf("packs = %+v", k.Packs)
	}
	var titles []string
	for _, e := range k.Entries {
		titles = append(titles, e.Title)
	}
	if len(titles) != 3 || titles[0] != "Gram and Aya" || titles[1] != "Helgen" || k.Entries[1].Pack == nil || *k.Entries[1].Pack != 2 || k.Entries[0].Pack != nil {
		t.Errorf("entries = %v (%+v)", titles, k.Entries)
	}
	// A character's own memory of type KNOWLEDGE isn't world knowledge.
	for _, e := range k.Entries {
		if e.ActorUUID != "" && e.ActorUUID != "0" {
			t.Errorf("a character's memory listed: %+v", e)
		}
	}

	k = must(h, func(tx store.DB) (KnowledgeResponse, error) {
		if err := SetPackEnabled(tx, 2, true); err != nil {
			return KnowledgeResponse{}, err
		}
		return Knowledge(tx)
	})
	if !k.Packs[0].Enabled {
		t.Error("the pack wasn't switched on")
	}
	if _, err := write(h, func(tx store.DB) (KnowledgeResponse, error) { return KnowledgeResponse{}, SetPackEnabled(tx, 9, true) }); err == nil {
		t.Error("switched a pack that doesn't exist")
	} else if kind, _ := apperr.KindOf(err); kind != apperr.NotFound {
		t.Errorf("error = %v", err)
	}

	pack := int64(1)
	r := must(h, func(tx store.DB) (SaveResult, error) {
		return AddKnowledge(tx, NewEntry{Pack: &pack, Changes: MemoryChanges{Title: ptr("Frida"), Content: ptr("Frida lives here."), Tags: &[]string{"resident"}}})
	})
	k, _ = store.Read(h.store, testenv.SaveID, Knowledge)
	if len(k.Entries) != 4 || k.Packs[1].Entries != 2 || !r.Reembed {
		t.Errorf("after adding: %d entries, pack has %d, reembed %v", len(k.Entries), k.Packs[1].Entries, r.Reembed)
	}
	if _, err := write(h, func(tx store.DB) (SaveResult, error) {
		return AddKnowledge(tx, NewEntry{Changes: MemoryChanges{Content: ptr("  ")}})
	}); err == nil {
		t.Error("added an empty entry")
	}
	nine := int64(9)
	if _, err := write(h, func(tx store.DB) (SaveResult, error) {
		return AddKnowledge(tx, NewEntry{Pack: &nine, Changes: MemoryChanges{Content: ptr("x")}})
	}); err == nil {
		t.Error("added to a pack that doesn't exist")
	}
}
