package characters

import (
	"fmt"
	"strings"
	"testing"

	"hermaeus/internal/store"
	"hermaeus/internal/testenv"
)

// withStory gives Serana (and her retired identity, Mysterious Woman) a few
// events, memories, and a diary entry. The player has two records named Gram
// (only 8 was there); Aya is 9; 12345 has no record at all.
func (h *harness) withStory() {
	h.t.Helper()
	h.WithAlias()
	s, m := testenv.SeranaUUID, testenv.MysteriousUUID
	h.Exec(`
insert into uuid_mappings (uuid, form_id, actor_name, bio_template_name, mod_file, local_form_id, created_at, updated_at) values
  (7, 20, 'Gram', 'player_special', null, null, null, null), (8, 20, 'Gram', 'player_special', null, null, null, null),
  (9, 4268943386, 'Aya', 'aya_81A', 'aya_follower.esp', 2074, 2461324.5, 2461334.5);
update uuid_mappings set mod_file = 'dawnguard.esm', local_form_id = 11124, form_id = 67119988, created_at = 2461328.5 where uuid = ` + s + `;
update identity_aliases set created_at = '2026-10-15 04:54:35';
insert into events (id, event_type, event_data, game_time, game_time_str, actor_UUIDs, originating_actor_UUID, target_actor_UUID) values
  (1, 'dialogue', '{"speaker":"Serana","listener":"Gram","dialogue":"Hello."}', 3600, '', '[` + s + `,8]', ` + s + `, 8),
  (2, 'dialogue', '{"speaker":"Serana","listener":"Gram","dialogue":"Again."}', 7200, '', '[` + s + `,8]', ` + s + `, 8),
  (3, 'death', '{"killer":"Serana","victim":"Wolf"}', 1800, '', '[` + m + `,9]', 0, 0),
  (4, 'follower_recruited', 'Serana is now traveling with Gram as a companion.', 86400, '', '[]', 0, ` + s + `),
  (5, 'trade_complete', '{"items_bought":"10x Iron Ingot","items_sold":"","merchant":"Adrianne Avenicci","net_gold":-1046}', 90000, '', '[` + s + `]', ` + s + `, 0),
  (6, 'dialogue', '{"speaker":"Aya"}', 100000, '', 'not json', 9, 0);
insert into memories (id, actor_uuid, content, memory_type, importance_score, emotion, location, related_actors, is_active, always_inject) values
  (1, ` + s + `, 'Met the Dragonborn.', 'JOY', 0.8, 'curious', 'Dimhollow Crypt', '[9, 12345, ` + s + `]', 1, 0),
  (2, ` + s + `, 'A quiet night.', 'JOY', 0.4, '', null, '[]', 0, 1),
  (50, ` + m + `, 'I awaken in Dimhollow Crypt.', 'EXPERIENCE', 0.5, 'curious', 'Dimhollow Crypt', '[]', 1, 0),
  (3, 9, 'Not hers.', 'JOY', 0.5, null, null, '[]', 1, 0);
insert into memory_embeddings (memory_id, embedding, dimension) values (1, x'00', 384), (2, x'', 0);
insert into diary_entries (actor_uuid, content) values (` + s + `, 'Dear diary.');`)
}

func (h *harness) overview(uuid string) Overview {
	h.t.Helper()
	o, err := store.Read(h.store, testenv.SaveID, func(db store.DB) (Overview, error) { return OverviewOf(db, uuid) })
	if err != nil {
		h.t.Fatal(err)
	}
	return o
}

func (h *harness) stats(uuid string) Stats {
	h.t.Helper()
	s, err := store.Read(h.store, testenv.SaveID, func(db store.DB) (Stats, error) { return StatsOf(db, uuid) })
	if err != nil {
		h.t.Fatal(err)
	}
	return s
}

func TestOverview(t *testing.T) {
	h := newHarness(t)
	h.withStory()
	o := h.overview(testenv.SeranaUUID)
	if o.Name != "Serana" || o.Bio != "serana" || o.FormID != "04002B74" || o.ModFile != "dawnguard.esm" || o.LocalFormID != "2B74" {
		t.Errorf("identity = %+v", o)
	}
	if o.FirstSeen == nil || *o.FirstSeen != 1792022400000 { // 2026-10-15, midnight UTC
		t.Errorf("first seen = %v", show(o.FirstSeen))
	}
	if len(o.Retired) != 1 || o.Retired[0].Name != "Mysterious Woman" || show(o.Retired[0].Since) != "1792040075000" {
		t.Errorf("retired = %+v", o.Retired)
	}
	// Her first event was under her retired identity.
	if o.FirstEvent == nil || o.FirstEvent.Time != "12:30 AM" || o.LastEvent == nil || o.LastEvent.Day != "Morndas, 18th of Last Seed, 4E 201" {
		t.Errorf("first, last = %+v, %+v", o.FirstEvent, o.LastEvent)
	}
	want := "[follower_recruited: Serana is now traveling with Gram as a companion. | trade_complete: Traded with Adrianne Avenicci: bought 10x Iron Ingot (-1,046 gold)]"
	var got []string
	for _, m := range o.Milestones {
		got = append(got, m.Type+": "+m.Summary)
	}
	if s := "[" + strings.Join(got, " | ") + "]"; s != want {
		t.Errorf("milestones =\n %s\nwant\n %s", s, want)
	}
}

func TestStats(t *testing.T) {
	h := newHarness(t)
	h.withStory()
	s := h.stats(testenv.SeranaUUID)
	ev := s.Events
	if ev.Total != 5 || ev.Started != 3 || ev.Targeted != 1 || ev.Kills != 1 {
		t.Errorf("events = %d, started %d, targeted %d, kills %d; want 5, 3, 1, 1 (counting her earlier identity)", ev.Total, ev.Started, ev.Targeted, ev.Kills)
	}
	if fmt.Sprint(ev.Types[0]) != "{dialogue  Dialogue 2}" {
		t.Errorf("event types = %v", ev.Types)
	}
	if len(s.TalksTo) != 1 || s.TalksTo[0] != (Count{UUID: "8", Name: "Gram", N: 2}) {
		t.Errorf("talks to = %+v, want Gram (the record she was with) twice", s.TalksTo)
	}
	if fmt.Sprint(s.Company) != "[{ 8 Gram 2} { 9 Aya 1}]" {
		t.Errorf("company = %v", s.Company)
	}
	if fmt.Sprint(s.Mentions) != "[{ 9 Aya 1}]" || s.UnknownMentions != 1 {
		t.Errorf("mentions = %v, unknown %d", s.Mentions, s.UnknownMentions)
	}
	m := s.Memories
	if m.Total != 3 || m.Inactive != 1 || m.AlwaysInject != 1 || m.PendingEmbedding != 1 || m.Unembeddable != 1 {
		t.Errorf("memories = %+v", m)
	}
	if fmt.Sprint(m.Types) != "[{JOY 2 0.6000000000000001} {EXPERIENCE 1 0.5}]" {
		t.Errorf("memory types = %v", m.Types)
	}
	if fmt.Sprint(m.Emotions) != "[{  curious 2} {   1}]" || fmt.Sprint(m.Places) != "[{  Dimhollow Crypt 2} {   1}]" {
		t.Errorf("emotions %v, places %v", m.Emotions, m.Places)
	}
	if s.Diaries != 1 {
		t.Errorf("diaries = %d", s.Diaries)
	}
}

func TestNothingStored(t *testing.T) {
	h := newHarness(t)
	h.Exec(`drop table events`)
	if o := h.overview("7"); o.Name != "" || o.FirstEvent != nil || len(o.Milestones) != 0 {
		t.Errorf("overview = %+v", o)
	}
	if s := h.stats("7"); s.Events.Total != 0 || s.Memories.Total != 0 || len(s.Company) != 0 {
		t.Errorf("stats = %+v", s)
	}
}

func TestMilestoneSummary(t *testing.T) {
	for _, c := range []struct{ typ, data, want string }{
		{"follower_recruited", "Vilja is now traveling with Gram as a companion.", "Vilja is now traveling with Gram as a companion."},
		{"brawl_ended", "", "Brawl ended"},
		{"item_given", `{"item":"Torch"}`, "Item given"},
		{"trade_complete", `{"items_sold":"Wolf Pelt","net_gold":1234567}`, "Traded with a merchant: sold Wolf Pelt (+1,234,567 gold)"},
		{"trade_complete", `{}`, "Traded with a merchant: nothing listed"},
	} {
		if got := milestoneSummary(c.typ, c.data); got != c.want {
			t.Errorf("milestoneSummary(%s, %s) = %q, want %q", c.typ, c.data, got, c.want)
		}
	}
}

func show[T any](p *T) string {
	if p == nil {
		return "<nil>"
	}
	return fmt.Sprint(*p)
}
