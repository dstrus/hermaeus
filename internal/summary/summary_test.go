package summary

import (
	"fmt"
	"testing"

	"hermaeus/internal/store"
	"hermaeus/internal/testenv"
)

func summarize(t *testing.T, f *testenv.Folder) Summary {
	t.Helper()
	st := store.New("", func() bool { return false })
	if _, err := st.SetDir(f.Dir); err != nil {
		t.Fatal(err)
	}
	s, err := store.Read(st, testenv.SaveID, Of)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestANewPlaythroughHasNothingYet(t *testing.T) {
	s := summarize(t, testenv.NewFolder(t))
	if s.Now != nil || s.Days != 0 || s.Counts.Events != 0 || len(s.Diary) != 0 || len(s.Events) != 0 || len(s.Company) != 0 {
		t.Errorf("summary = %+v", s)
	}
}

func TestSummary(t *testing.T) {
	f := testenv.NewFolder(t)
	serana, vilja, aela := testenv.SeranaUUID, "3122424291542467906", "77"
	f.Exec(`
insert into uuid_mappings (uuid, form_id, actor_name) values (` + vilja + `, 2, 'Vilja'), (` + aela + `, 3, 'Aela');
insert into npc_groups (id, group_name, description) values (1, 'Pinned', '');
insert into npc_group_members (group_id, actor_uuid) values (1, ` + aela + `);
insert into events (event_type, event_data, game_time, game_time_str, originating_actor_UUID) values
  ('follower_recruited', 'Vilja is now traveling with Gram as a companion.', 3600, '1:00 AM, Sundas, 17th of Last Seed, 4E 201', ` + vilja + `),
  ('follower_recruited', 'Serana is now traveling with Gram as a companion.', 90000, '1:00 AM, Morndas, 18th of Last Seed, 4E 201', ` + serana + `),
  ('follower_dismissed', 'Vilja has stopped traveling with Gram.', 180000, '2:00 AM, Tirdas, 19th of Last Seed, 4E 201', ` + vilja + `);
insert into memories (actor_uuid, content) values (` + serana + `, 'a'), (` + serana + `, 'b'), (` + aela + `, 'c'), (0, 'The world is round.');`)
	for i := 1; i <= 5; i++ {
		f.Exec(fmt.Sprintf(`insert into diary_entries (id, actor_uuid, content) values (%d, %s, 'Entry %d.')`, i, serana, i))
	}

	s := summarize(t, f)
	if s.Began == nil || s.Now == nil || s.Began.Day != "Sundas, 17th of Last Seed, 4E 201" || s.Now.Day != "Tirdas, 19th of Last Seed, 4E 201" || s.Days != 3 {
		t.Errorf("dates: began %+v, now %+v, %d days", s.Began, s.Now, s.Days)
	}
	if c := s.Counts; c.Events != 3 || c.Memories != 3 || c.Knowledge != 1 || c.Diary != 5 || c.Characters != 2 {
		t.Errorf("counts = %+v", c)
	}
	if len(s.Diary) != DiaryShown || s.Diary[0].Content != "Entry 5." {
		t.Errorf("diary = %+v", s.Diary)
	}
	if len(s.Events) != 3 || s.Events[0].Type != "follower_dismissed" {
		t.Errorf("events = %+v", s.Events)
	}
	// Serana joined and stayed; Vilja left; Aela is pinned.
	var got string
	for _, c := range s.Company {
		got += fmt.Sprintf("%s:%v:%v ", c.Name, c.Traveling, c.Pinned)
	}
	if got != "Serana:true:false Aela:false:true " {
		t.Errorf("company = %s", got)
	}
}
