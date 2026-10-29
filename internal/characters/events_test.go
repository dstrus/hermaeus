package characters

import (
	"strings"
	"testing"

	"hermaeus/internal/store"
	"hermaeus/internal/testenv"
)

func TestEventsOfACharacter(t *testing.T) {
	f := testenv.NewFolder(t)
	f.WithAlias()
	s, m := testenv.SeranaUUID, testenv.MysteriousUUID
	f.Exec(`insert into events (id, event_type, event_data, game_time, game_time_str, actor_UUIDs, originating_actor_UUID, target_actor_UUID) values
  (1, 'dialogue', '{"speaker":"Serana","listener":"Gram","dialogue":"Hello."}', 3600, '', '[` + s + `,8]', ` + s + `, 8),
  (2, 'death', '{"killer":"Serana","victim":"Wolf"}', 1800, '', '[` + m + `,9]', 0, 0),
  (3, 'follower_recruited', 'Serana is now traveling with Gram as a companion.', 86400, '', '[]', 0, ` + s + `),
  (4, 'book_read', '{"book_name":"The Book of Life and Service","book_text":"` + strings.Repeat("x", 3000) + `"}', 90000, '', '[` + s + `]', ` + s + `, 0),
  (5, 'dialogue', '{"speaker":"Aya","dialogue":"Not hers."}', 100000, '', '[9]', 9, 0);`)
	st := store.New("", func() bool { return false })
	if _, err := st.SetDir(f.Dir); err != nil {
		t.Fatal(err)
	}
	r, err := store.Read(st, testenv.SaveID, func(db store.DB) (EventsResponse, error) { return EventsOf(db, s) })
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, e := range r.Events {
		got = append(got, e.Role+": "+e.Summary)
	}
	want := []string{
		"acting: Read The Book of Life and Service",
		"acted on: Serana is now traveling with Gram as a companion.",
		"acting: Serana to Gram: “Hello.”",
		"present: Serana killed Wolf", // under her earlier identity
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("events =\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
	if r.Events[2].Speaker != "Serana" || r.Events[2].Listener != "Gram" || r.Events[2].Line != "Hello." || r.Events[0].Speaker != "" || r.Events[2].SpeakerUUID != s || r.Events[2].ListenerUUID != "8" {
		t.Errorf("who said what = %+v", r.Events[2])
	}
	if r.Events[0].Present != 1 || r.Events[0].Date == nil || len(r.Events[0].Data["book_text"].(string)) > detailText+4 {
		t.Errorf("book event = %+v", r.Events[0])
	}
	if len(r.Types) != 4 || r.Types[0].Count != 1 || r.Types[0].Type != "book_read" {
		t.Errorf("types = %+v", r.Types)
	}
	if r.Events[2].From != "Serana" || r.Events[2].FromUUID != s || r.Events[1].To != "Serana" {
		t.Errorf("from and to = %+v / %+v", r.Events[2], r.Events[1])
	}

	// Everyone's events: Aya's too, with no part for anyone.
	all, err := store.Read(st, testenv.SaveID, AllEvents)
	if err != nil {
		t.Fatal(err)
	}
	if len(all.Events) != 5 || all.Events[0].Summary != "Aya: “Not hers.”" || all.Events[0].Role != "" {
		t.Errorf("everyone's events = %+v", all.Events)
	}
}
