package characters

import (
	"hermaeus/internal/apperr"
	"hermaeus/internal/identity"
	"hermaeus/internal/store"
)

// person is one character for the overview and stats: their UUIDs (current
// first, then retired ones) and how to resolve everyone else's.
type person struct {
	uuid    string
	uuids   []string
	aliases *identity.Aliases
}

func loadPerson(db store.DB, uuid string) (*person, error) {
	if _, ok := identity.ParseUUID(uuid); !ok {
		return nil, apperr.New(apperr.Invalid, "Invalid character %q", uuid)
	}
	a, err := identity.Load(db)
	if err != nil {
		return nil, err
	}
	return &person{uuid: uuid, uuids: a.All(uuid), aliases: a}, nil
}

// in is the SQL list and arguments for "actor_uuid in (…)".
func (p *person) in() (string, []any) { return identity.SQLIn(p.uuids) }

// theirEvents starts a query with a CTE, "theirs", of the events the person
// took part in: those that list them as present, or name them as the one
// acting or acted on. Append the rest of the query and its arguments.
func (p *person) theirEvents() (string, []any) {
	in, ids := p.in()
	cte := `with theirs as (
		select e.* from events e
		where e.originating_actor_UUID in (` + in + `) or e.target_actor_UUID in (` + in + `)
		   or (json_valid(e.actor_UUIDs) and exists (select 1 from json_each(e.actor_UUIDs) j where j.value in (` + in + `))))
	`
	return cte, repeat(ids, 3)
}

// repeat returns args times times over, then extra, for a query that uses
// the same list several times.
func repeat(args []any, times int, extra ...any) []any {
	var out []any
	for range times {
		out = append(out, args...)
	}
	return append(out, extra...)
}

func hasEvents(db store.DB) (bool, error) { return store.HasTable(db, "events") }
