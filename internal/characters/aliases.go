// Package characters is about the player and the NPCs SkyrimNet knows:
// who they are, and which of them the player has pinned.
package characters

import (
	"strconv"

	"hermaeus/internal/store"
)

// When SkyrimNet re-identifies an NPC, for example when "Mysterious Woman"
// turns out to be Serana, it gives them a new actor UUID and records the old
// one in identity_aliases. Data stored under the old UUID stays there, so
// Hermaeus treats each such NPC as one person, under their newest UUID.

type aliases struct {
	to      map[string]string // retired UUID -> newest UUID, following chains
	oldName map[string]string // retired UUID -> the name it had
}

func loadAliases(db store.DB) (*aliases, error) {
	a := &aliases{to: map[string]string{}, oldName: map[string]string{}}
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

func (a *aliases) resolve(uuid string) string {
	if to, ok := a.to[uuid]; ok {
		return to
	}
	return uuid
}

// all returns uuid and every retired UUID that now resolves to it.
func (a *aliases) all(uuid string) []string {
	out := []string{uuid}
	for old, to := range a.to {
		if to == uuid {
			out = append(out, old)
		}
	}
	return out
}

// parseUUID reads an actor UUID. They're sent as strings, because they're
// larger than JavaScript can hold exactly as numbers.
func parseUUID(s string) (int64, bool) {
	n, err := strconv.ParseInt(s, 10, 64)
	return n, err == nil
}
