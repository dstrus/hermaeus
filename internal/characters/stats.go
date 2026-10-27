package characters

import (
	"database/sql"
	"sort"
	"strings"

	"hermaeus/internal/events"
	"hermaeus/internal/store"
)

// Stats is how much SkyrimNet has recorded about a character, and who and
// what they're most involved with. It counts their retired identities as
// theirs.
//
// Factions aren't here: SkyrimNet asks the game for them when it builds a
// prompt and doesn't store them.
type Stats struct {
	Events          EventStats  `json:"events"`
	Company         []Count     `json:"company"`         // others in the most events with them
	TalksTo         []Count     `json:"talksTo"`         // who they speak to most
	Mentions        []Count     `json:"mentions"`        // others their memories mention most
	UnknownMentions int         `json:"unknownMentions"` // mentions of actors SkyrimNet has no name for
	Memories        MemoryStats `json:"memories"`
	Diaries         int         `json:"diaries"`
}

type EventStats struct {
	Total    int     `json:"total"`
	Started  int     `json:"started"`  // events they were the one acting in
	Targeted int     `json:"targeted"` // events they were acted on in
	Kills    int     `json:"kills"`
	Types    []Count `json:"types"` // Name is the type in words; Key is the type
}

type MemoryStats struct {
	Total            int         `json:"total"`
	Inactive         int         `json:"inactive"`
	AlwaysInject     int         `json:"alwaysInject"`
	PendingEmbedding int         `json:"pendingEmbedding"` // no search vector yet; SkyrimNet builds them when it starts
	Unembeddable     int         `json:"unembeddable"`     // SkyrimNet couldn't build a search vector
	Types            []TypeCount `json:"types"`
	Emotions         []Count     `json:"emotions"` // Name "" counts memories without an emotion
	Places           []Count     `json:"places"`   // Name "" counts memories without a location
}

// Count is how often something comes up. For people, UUID links to them
// when SkyrimNet knows who they are.
type Count struct {
	Key  string `json:"key,omitempty"`
	UUID string `json:"uuid,omitempty"`
	Name string `json:"name"`
	N    int    `json:"n"`
}

type TypeCount struct {
	Type       string  `json:"type"`
	N          int     `json:"n"`
	Importance float64 `json:"importance"` // the average, 0 to 1
}

// statsTop is how many of the most frequent of each kind are listed.
const statsTop = 8

func StatsOf(db store.DB, uuid string) (Stats, error) {
	s := Stats{Events: EventStats{Types: []Count{}}, Company: []Count{}, TalksTo: []Count{}, Mentions: []Count{},
		Memories: MemoryStats{Types: []TypeCount{}, Emotions: []Count{}, Places: []Count{}}}
	p, err := loadPerson(db, uuid)
	if err != nil {
		return s, err
	}
	names, err := allNames(db)
	if err != nil {
		return s, err
	}
	in, ids := p.in()

	if ok, err := hasEvents(db); err != nil {
		return s, err
	} else if ok {
		if err := p.eventStats(db, names, &s); err != nil {
			return s, err
		}
	}

	var mentioned map[string]int
	if err := eachRow(db, `select cast(j.value as text), count(*) from memories m, json_each(m.related_actors) j
		where m.actor_uuid in (`+in+`) and m.pack_id is null and json_valid(m.related_actors) group by j.value`, ids, func(r *sql.Rows) error {
		var u string
		var n int
		if err := r.Scan(&u, &n); err != nil {
			return err
		}
		mentioned = p.add(mentioned, u, n)
		return nil
	}); err != nil {
		return s, err
	}
	s.Mentions, s.UnknownMentions = named(mentioned, names)

	mem := &s.Memories
	own := ` from memories m where m.actor_uuid in (` + in + `) and m.pack_id is null`
	if err := db.QueryRow(`select count(*), coalesce(sum(is_active = 0), 0), coalesce(sum(always_inject <> 0), 0),
			coalesce(sum(not exists (select 1 from memory_embeddings e where e.memory_id = m.id)), 0),
			coalesce(sum(exists (select 1 from memory_embeddings e where e.memory_id = m.id and e.dimension = 0)), 0)`+own, ids...).
		Scan(&mem.Total, &mem.Inactive, &mem.AlwaysInject, &mem.PendingEmbedding, &mem.Unembeddable); err != nil {
		return s, err
	}
	if err := eachRow(db, `select coalesce(memory_type, ''), count(*), coalesce(avg(importance_score), 0)`+own+` group by 1 order by 2 desc, 1`, ids, func(r *sql.Rows) error {
		var t TypeCount
		if err := r.Scan(&t.Type, &t.N, &t.Importance); err != nil {
			return err
		}
		mem.Types = append(mem.Types, t)
		return nil
	}); err != nil {
		return s, err
	}
	if mem.Emotions, err = topCounts(db, `select coalesce(nullif(trim(emotion), ''), ''), count(*)`+own+` group by 1 order by 2 desc, 1`, ids); err != nil {
		return s, err
	}
	if mem.Places, err = topCounts(db, `select coalesce(nullif(trim(location), ''), ''), count(*)`+own+` group by 1 order by 2 desc, 1`, ids); err != nil {
		return s, err
	}
	err = db.QueryRow(`select count(*) from diary_entries where actor_uuid in (`+in+`)`, ids...).Scan(&s.Diaries)
	return s, err
}

func (p *person) eventStats(db store.DB, names map[string]string, s *Stats) error {
	theirs, args := p.theirEvents()
	in, ids := p.in()
	ev := &s.Events
	if err := db.QueryRow(theirs+`select count(*), coalesce(sum(originating_actor_UUID in (`+in+`)), 0),
			coalesce(sum(target_actor_UUID in (`+in+`)), 0) from theirs`, append(args, repeat(ids, 2)...)...).
		Scan(&ev.Total, &ev.Started, &ev.Targeted); err != nil {
		return err
	}
	if err := eachRow(db, theirs+`select event_type, count(*) from theirs group by event_type order by 2 desc, 1`, args, func(r *sql.Rows) error {
		var c Count
		if err := r.Scan(&c.Key, &c.N); err != nil {
			return err
		}
		c.Name = events.Label(c.Key)
		ev.Types = append(ev.Types, c)
		return nil
	}); err != nil {
		return err
	}

	var together map[string]int
	if err := eachRow(db, theirs+`select cast(j.value as text), count(*) from theirs, json_each(theirs.actor_UUIDs) j
		where json_valid(theirs.actor_UUIDs) group by j.value`, args, func(r *sql.Rows) error {
		var u string
		var n int
		if err := r.Scan(&u, &n); err != nil {
			return err
		}
		together = p.add(together, u, n)
		return nil
	}); err != nil {
		return err
	}
	s.Company, _ = named(together, names)

	// Kills and dialogue name actors by name only.
	var myNames []any
	for _, u := range p.uuids {
		if n := names[u]; n != "" {
			myNames = append(myNames, n)
		}
	}
	if len(myNames) == 0 {
		return nil
	}
	nameIn := strings.TrimSuffix(strings.Repeat("?,", len(myNames)), ",")
	if err := db.QueryRow(theirs+`select count(*) from theirs where event_type = 'death' and json_valid(event_data)
		and json_extract(event_data, '$.killer') in (`+nameIn+`)`, append(args, myNames...)...).Scan(&ev.Kills); err != nil {
		return err
	}
	listeners, err := topCounts(db, theirs+`select json_extract(event_data, '$.listener'), count(*) from theirs
		where event_type = 'dialogue' and json_valid(event_data) and json_extract(event_data, '$.speaker') in (`+nameIn+`)
		  and coalesce(json_extract(event_data, '$.listener'), '') <> ''
		group by 1 order by 2 desc, 1`, append(args, myNames...))
	if err != nil {
		return err
	}
	// A name can belong to several NPCs (the player can have more than one
	// record, and creatures share names), so a listener links to the one the
	// character spent the most time with, or to the only one there is.
	byName := map[string][]string{}
	for u, n := range names {
		if p.aliases.Resolve(u) == u {
			byName[n] = append(byName[n], u)
		}
	}
	for _, l := range listeners {
		if c := byName[l.Name]; len(c) == 1 {
			l.UUID = c[0]
		} else {
			for _, u := range c {
				if together[u] > together[l.UUID] {
					l.UUID = u
				}
			}
		}
		s.TalksTo = append(s.TalksTo, l)
	}
	return nil
}

// add counts n more for whoever uuid belongs to, unless it's the person
// themself or the world (actor 0).
func (p *person) add(counts map[string]int, uuid string, n int) map[string]int {
	if counts == nil {
		counts = map[string]int{}
	}
	if u := p.aliases.Resolve(uuid); u != "0" && !p.is(u) {
		counts[u] += n
	}
	return counts
}

func (p *person) is(uuid string) bool {
	for _, u := range p.uuids {
		if u == uuid {
			return true
		}
	}
	return false
}

// named turns counts by UUID into the most frequent named people, and the
// total for those SkyrimNet has no name for.
func named(counts map[string]int, names map[string]string) (top []Count, unknown int) {
	top = []Count{}
	for u, n := range counts {
		if names[u] == "" {
			unknown += n
			continue
		}
		top = append(top, Count{UUID: u, Name: names[u], N: n})
	}
	sort.Slice(top, func(i, j int) bool {
		if top[i].N != top[j].N {
			return top[i].N > top[j].N
		}
		return strings.ToLower(top[i].Name) < strings.ToLower(top[j].Name)
	})
	if len(top) > statsTop {
		top = top[:statsTop]
	}
	return top, unknown
}

// topCounts reads the first statsTop (name, count) rows of a query.
func topCounts(db store.DB, query string, args []any) ([]Count, error) {
	out := []Count{}
	err := eachRow(db, query, args, func(r *sql.Rows) error {
		var name sql.NullString
		var n int
		if err := r.Scan(&name, &n); err != nil {
			return err
		}
		if len(out) < statsTop {
			out = append(out, Count{Name: name.String, N: n})
		}
		return nil
	})
	return out, err
}

// allNames maps every actor UUID SkyrimNet knows to their name.
func allNames(db store.DB) (map[string]string, error) {
	names := map[string]string{}
	err := eachRow(db, `select cast(uuid as text), coalesce(actor_name, '') from uuid_mappings`, nil, func(r *sql.Rows) error {
		var u, n string
		if err := r.Scan(&u, &n); err != nil {
			return err
		}
		names[u] = n
		return nil
	})
	return names, err
}

func eachRow(db store.DB, query string, args []any, fn func(*sql.Rows) error) error {
	rows, err := db.Query(query, args...)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		if err := fn(rows); err != nil {
			return err
		}
	}
	return rows.Err()
}
