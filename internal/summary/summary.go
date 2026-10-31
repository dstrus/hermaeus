// Package summary is a playthrough at a glance: where the story stands, what
// was written and done last, how much SkyrimNet has stored, and who travels
// with the player.
package summary

import (
	"database/sql"
	"sort"

	"hermaeus/internal/calendar"
	"hermaeus/internal/characters"
	"hermaeus/internal/memories"
	"hermaeus/internal/store"
)

// How many of the newest diary entries and events a summary carries.
const (
	DiaryShown  = 3
	EventsShown = 12
)

type Summary struct {
	Now   *calendar.GameDate `json:"now,omitempty"`   // the newest event's date
	Began *calendar.GameDate `json:"began,omitempty"` // the oldest event's date
	Days  int                `json:"days"`            // in-game days with something recorded
	// When SkyrimNet last wrote to the playthrough, Unix ms; 0 if unknown.
	Written int64                 `json:"written"`
	Counts  Counts                `json:"counts"`
	Diary   []memories.DiaryEntry `json:"diary"`  // newest first
	Events  []characters.Event    `json:"events"` // newest first
	Company []Companion           `json:"company"`
}

// Counts is how much SkyrimNet has stored in each of the playthrough's
// sections.
type Counts struct {
	Characters  int `json:"characters"` // with something stored, as Characters lists them
	Memories    int `json:"memories"`   // characters' own, not world knowledge
	Diary       int `json:"diary"`
	Events      int `json:"events"`
	Knowledge   int `json:"knowledge"` // world-knowledge entries
	Screenshots int `json:"screenshots"`
	Songs       int `json:"songs"`
}

// Companion is a character traveling with the player, or one the player
// pinned.
type Companion struct {
	UUID      string `json:"uuid"`
	Name      string `json:"name"`
	Memories  int    `json:"memories"`
	Pinned    bool   `json:"pinned"`
	Traveling bool   `json:"traveling"` // joined the player and hasn't been dismissed since
}

// Of sums up a playthrough.
func Of(db store.DB) (Summary, error) {
	s := Summary{Diary: []memories.DiaryEntry{}, Events: []characters.Event{}, Company: []Companion{}}
	list, err := characters.List(db)
	if err != nil {
		return s, err
	}
	s.Counts.Characters = len(list.Characters)
	for table, n := range map[string]*int{
		"diary_entries":         &s.Counts.Diary,
		"events":                &s.Counts.Events,
		"omnisight_screenshots": &s.Counts.Screenshots,
		"bard_songs":            &s.Counts.Songs,
	} {
		if err := countRows(db, table, "1 = 1", n); err != nil {
			return s, err
		}
	}
	// The world (actor 0) holds knowledge; everyone else's are memories.
	if err := countRows(db, "memories", "actor_uuid <> 0", &s.Counts.Memories); err != nil {
		return s, err
	}
	if err := countRows(db, "memories", "actor_uuid = 0", &s.Counts.Knowledge); err != nil {
		return s, err
	}
	if err := s.dates(db); err != nil {
		return s, err
	}
	if s.Counts.Diary > 0 {
		d, err := memories.AllDiary(db)
		if err != nil {
			return s, err
		}
		s.Diary = d.Entries[:min(DiaryShown, len(d.Entries))]
	}
	if s.Counts.Events > 0 {
		if s.Events, err = characters.LatestEvents(db, EventsShown); err != nil {
			return s, err
		}
	}
	s.Company, err = company(db, list.Characters)
	return s, err
}

// countRows counts a table's rows that match where, or none when the
// database is too old to have the table.
func countRows(db store.DB, table, where string, n *int) error {
	if ok, err := store.HasTable(db, table); err != nil || !ok {
		return err
	}
	return db.QueryRow(`select count(*) from ` + table + ` where ` + where).Scan(n)
}

// dates finds the first and last in-game dates, how many days have
// something recorded, and when SkyrimNet last wrote an event.
func (s *Summary) dates(db store.DB) error {
	if s.Counts.Events == 0 {
		return nil
	}
	cal, err := calendar.Load(db)
	if err != nil {
		return err
	}
	rows, err := db.Query(`select game_time from events where game_time is not null order by game_time`)
	if err != nil {
		return err
	}
	defer rows.Close()
	days := map[string]bool{}
	var first, last float64
	for rows.Next() {
		var t float64
		if err := rows.Scan(&t); err != nil {
			return err
		}
		d, ok := cal.Date(t)
		if !ok {
			continue
		}
		if first == 0 {
			first = t
		}
		last = t
		days[d.Day] = true
	}
	if err := rows.Err(); err != nil {
		return err
	}
	s.Began, s.Now, s.Days = cal.DatePtr(first), cal.DatePtr(last), len(days)
	var written any
	if err := db.QueryRow(`select max(local_time) from events`).Scan(&written); err != nil {
		return err
	}
	if w := store.Timestamp(written); w != nil {
		s.Written = *w
	}
	return nil
}

// company is who travels with the player, by the last time each joined or
// left, and who the player pinned: those traveling first, then by memories.
func company(db store.DB, list []characters.Character) ([]Companion, error) {
	out := []Companion{}
	traveling := map[string]bool{}
	if ok, err := store.HasTable(db, "events"); err != nil {
		return out, err
	} else if ok {
		rows, err := db.Query(`select event_type, cast(originating_actor_UUID as text) from events
			where event_type in ('follower_recruited', 'follower_dismissed')
			order by game_time is null, game_time, id`)
		if err != nil {
			return out, err
		}
		defer rows.Close()
		for rows.Next() {
			var kind string
			var uuid sql.NullString
			if err := rows.Scan(&kind, &uuid); err != nil {
				return out, err
			}
			traveling[uuid.String] = kind == "follower_recruited"
		}
		if err := rows.Err(); err != nil {
			return out, err
		}
	}
	for _, c := range list {
		// The list knows characters by their current identity, which is
		// what a joining event names unless SkyrimNet re-identified them
		// since; a retired identity's joining isn't carried over.
		if c.Pinned || traveling[c.UUID] {
			out = append(out, Companion{UUID: c.UUID, Name: c.Name, Memories: c.Memories, Pinned: c.Pinned, Traveling: traveling[c.UUID]})
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Traveling != out[j].Traveling {
			return out[i].Traveling
		}
		return out[i].Memories > out[j].Memories
	})
	return out, nil
}
