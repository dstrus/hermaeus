// Package identity resolves actor UUIDs to the people they belong to.
//
// When SkyrimNet re-identifies an NPC, for example when "Mysterious Woman"
// turns out to be Serana, it gives them a new actor UUID and records the old
// one in identity_aliases. Data stored under the old UUID stays there, and
// SkyrimNet doesn't appear to recall it any more. So Hermaeus treats each
// such NPC as one person, under their newest UUID, and flags what's still
// stored under a retired one.
package identity

import (
	"strconv"

	"hermaeus/internal/store"
)

type Aliases struct {
	to      map[string]string // retired UUID -> newest UUID, following chains
	oldName map[string]string // retired UUID -> the name it had
}

func Load(db store.DB) (*Aliases, error) {
	a := &Aliases{to: map[string]string{}, oldName: map[string]string{}}
	if ok, err := store.HasTable(db, "identity_aliases"); err != nil || !ok {
		return a, err // older databases don't have the table
	}
	rows, err := db.Query(`select cast(old_uuid as text), cast(new_uuid as text), coalesce(old_name, '') from identity_aliases order by id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	next := map[string]string{}
	for rows.Next() {
		var old, nu, name string
		if err := rows.Scan(&old, &nu, &name); err != nil {
			return nil, err
		}
		if old != nu {
			next[old] = nu
			a.oldName[old] = name
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for old := range next {
		u, seen := old, map[string]bool{old: true}
		for next[u] != "" && !seen[next[u]] {
			u = next[u]
			seen[u] = true
		}
		if next[u] == "" { // a loop has no newest UUID, so leave it alone
			a.to[old] = u
		}
	}
	return a, nil
}

// Resolve returns the newest UUID of the person uuid belongs to.
func (a *Aliases) Resolve(uuid string) string {
	if to, ok := a.to[uuid]; ok {
		return to
	}
	return uuid
}

// Retired reports whether uuid is a retired identity, and the name it had
// ("" if SkyrimNet didn't record one).
func (a *Aliases) Retired(uuid string) (name string, retired bool) {
	if _, ok := a.to[uuid]; !ok {
		return "", false
	}
	return a.oldName[uuid], true
}

// All returns uuid and every retired UUID that now resolves to it, uuid first.
func (a *Aliases) All(uuid string) []string {
	out := []string{uuid}
	for old, to := range a.to {
		if to == uuid {
			out = append(out, old)
		}
	}
	return out
}

// ParseUUID reads an actor UUID. They're sent as strings, because they're
// larger than JavaScript can hold exactly as numbers.
func ParseUUID(s string) (int64, bool) {
	n, err := strconv.ParseInt(s, 10, 64)
	return n, err == nil
}
