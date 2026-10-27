package characters

import (
	"hermaeus/internal/apperr"
	"hermaeus/internal/identity"
	"hermaeus/internal/store"
)

// Pinned NPCs are SkyrimNet's own: its dashboard keeps them as members of an
// npc_groups row named "Pinned", so a pin made in either place shows in both.
// Pins on a retired identity count for the NPC's current one.

const (
	pinnedGroup     = "Pinned"
	pinnedGroupDesc = "Actors pinned for quick access on the dashboard"
)

type PinnedNPC struct {
	UUID string `json:"uuid"`
	Name string `json:"name"` // "" if SkyrimNet has no name for them
}

type PinsResponse struct {
	Pins []PinnedNPC `json:"pins"`
}

// Pins lists the pinned NPCs under their current UUIDs, in the order they
// were pinned.
func Pins(db store.DB) (PinsResponse, error) {
	out := PinsResponse{Pins: []PinnedNPC{}}
	if ok, err := store.HasTable(db, "npc_group_members"); err != nil || !ok {
		return out, err // older databases don't have groups
	}
	a, err := identity.Load(db)
	if err != nil {
		return out, err
	}
	rows, err := db.Query(`select cast(m.actor_uuid as text) from npc_group_members m
		join npc_groups g on g.id = m.group_id where g.group_name = ? order by m.rowid`, pinnedGroup)
	if err != nil {
		return out, err
	}
	var uuids []string
	seen := map[string]bool{}
	for rows.Next() {
		var u string
		if err := rows.Scan(&u); err != nil {
			rows.Close()
			return out, err
		}
		if u = a.Resolve(u); !seen[u] {
			seen[u] = true
			uuids = append(uuids, u)
		}
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return out, err
	}
	for _, u := range uuids {
		c, err := lookup(db, u)
		if err != nil {
			return out, err
		}
		out.Pins = append(out.Pins, PinnedNPC{UUID: u, Name: c.Name})
	}
	return out, nil
}

// SetPinned pins or unpins an NPC. Unpinning also removes pins left on
// their retired identities.
func SetPinned(tx store.DB, uuid string, pinned bool) error {
	actor, ok := identity.ParseUUID(uuid)
	if !ok {
		return apperr.New(apperr.Invalid, "Invalid NPC %q", uuid)
	}
	if actor == 0 {
		return apperr.New(apperr.Invalid, "Only NPCs can be pinned")
	}
	if ok, err := store.HasTable(tx, "npc_group_members"); err != nil || !ok {
		if err == nil {
			err = apperr.New(apperr.Conflict, "This playthrough's database doesn't support pinned NPCs. A newer SkyrimNet adds them.")
		}
		return err
	}
	var group int64
	err := tx.QueryRow(`select id from npc_groups where group_name = ?`, pinnedGroup).Scan(&group)
	switch {
	case store.IsNoRows(err) && !pinned:
		return nil
	case store.IsNoRows(err):
		res, err := tx.Exec(`insert into npc_groups (group_name, description) values (?, ?)`, pinnedGroup, pinnedGroupDesc)
		if err != nil {
			return err
		}
		if group, err = res.LastInsertId(); err != nil {
			return err
		}
	case err != nil:
		return err
	}
	if pinned {
		_, err = tx.Exec(`insert or ignore into npc_group_members (group_id, actor_uuid) values (?, ?)`, group, actor)
		return err
	}
	a, err := identity.Load(tx)
	if err != nil {
		return err
	}
	for _, u := range a.All(uuid) {
		n, _ := identity.ParseUUID(u)
		if _, err := tx.Exec(`delete from npc_group_members where group_id = ? and actor_uuid = ?`, group, n); err != nil {
			return err
		}
	}
	return nil
}
