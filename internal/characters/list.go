package characters

import (
	"sort"
	"strings"

	"hermaeus/internal/store"
)

// Group is where a character is listed. The player comes first, then pinned
// NPCs, then NPCs with many memories, then everyone else.
type Group string

const (
	GroupPlayer Group = "player"
	GroupPinned Group = "pinned"
	GroupBusy   Group = "busy" // BusyMemories or more memories
	GroupOther  Group = "other"
)

var groupOrder = map[Group]int{GroupPlayer: 0, GroupPinned: 1, GroupBusy: 2, GroupOther: 3}

// BusyMemories is how many memories put an unpinned NPC in GroupBusy.
const BusyMemories = 10

// playerFormID is the player's form ID in every Skyrim save (0x14).
const playerFormID = 20

type Character struct {
	UUID        string   `json:"uuid"`
	Name        string   `json:"name"`    // "" if SkyrimNet has no name for them
	Bio         string   `json:"bio"`     // their bio template's name, if any
	ModFile     string   `json:"modFile"` // the plugin that adds them, if known
	Memories    int      `json:"memories"`
	Diaries     int      `json:"diaries"`
	Pinned      bool     `json:"pinned"`
	Group       Group    `json:"group"`
	FormerNames []string `json:"formerNames"` // names of identities SkyrimNet has since retired
}

type CharactersResponse struct {
	Characters []Character `json:"characters"` // in Group order, then by name
}

// List returns the characters worth listing: everyone with memories or
// diary entries, and every pinned NPC. An NPC SkyrimNet re-identified is
// listed once, under their newest identity, with everything stored under
// their retired ones counted.
func List(db store.DB) (CharactersResponse, error) {
	out := CharactersResponse{Characters: []Character{}}
	a, err := loadAliases(db)
	if err != nil {
		return out, err
	}
	pins, err := Pins(db)
	if err != nil {
		return out, err
	}

	byUUID := map[string]*Character{}
	var order []string
	get := func(uuid string) (*Character, error) {
		if c := byUUID[uuid]; c != nil {
			return c, nil
		}
		c, err := lookup(db, uuid)
		if err != nil {
			return nil, err
		}
		byUUID[uuid] = c
		order = append(order, uuid)
		return c, nil
	}

	counts, err := countStored(db)
	if err != nil {
		return out, err
	}
	for _, n := range counts {
		c, err := get(a.resolve(n.uuid))
		if err != nil {
			return out, err
		}
		c.Memories += n.memories
		c.Diaries += n.diaries
		if a.resolve(n.uuid) != n.uuid {
			c.FormerNames = append(c.FormerNames, a.oldName[n.uuid])
		}
	}
	for _, p := range pins.Pins {
		c, err := get(p.UUID)
		if err != nil {
			return out, err
		}
		c.Pinned = true
	}

	for _, uuid := range order {
		c := byUUID[uuid]
		if c.Group != GroupPlayer {
			switch {
			case c.Pinned:
				c.Group = GroupPinned
			case c.Memories >= BusyMemories:
				c.Group = GroupBusy
			default:
				c.Group = GroupOther
			}
		}
		sort.Strings(c.FormerNames)
		out.Characters = append(out.Characters, *c)
	}
	sort.SliceStable(out.Characters, func(i, j int) bool {
		ci, cj := out.Characters[i], out.Characters[j]
		if gi, gj := groupOrder[ci.Group], groupOrder[cj.Group]; gi != gj {
			return gi < gj
		}
		return sortName(ci) < sortName(cj)
	})
	return out, nil
}

// sortName puts unnamed characters after named ones.
func sortName(c Character) string {
	if c.Name == "" {
		return "￿" + c.UUID
	}
	return strings.ToLower(c.Name)
}

type stored struct {
	uuid              string
	memories, diaries int
}

// countStored counts each actor's memories and diary entries. Knowledge
// (memories in a pack, or under actor 0) isn't anyone's.
func countStored(db store.DB) ([]stored, error) {
	rows, err := db.Query(`
		with counts as (
			select actor_uuid as u, count(*) as memories, 0 as diaries from memories where pack_id is null group by actor_uuid
			union all
			select actor_uuid, 0, count(*) from diary_entries group by actor_uuid
		)
		select cast(u as text), sum(memories), sum(diaries) from counts where u <> 0 group by u`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []stored
	for rows.Next() {
		var s stored
		if err := rows.Scan(&s.uuid, &s.memories, &s.diaries); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// lookup returns what SkyrimNet knows about a character, or just their UUID
// if it knows nothing.
func lookup(db store.DB, uuid string) (*Character, error) {
	c := &Character{UUID: uuid, FormerNames: []string{}}
	n, _ := parseUUID(uuid)
	var name, bio, mod *string
	var formID int64
	err := db.QueryRow(`select actor_name, bio_template_name, mod_file, form_id from uuid_mappings where uuid = ?`, n).Scan(&name, &bio, &mod, &formID)
	if store.IsNoRows(err) {
		return c, nil
	}
	if err != nil {
		return nil, err
	}
	c.Name, c.Bio, c.ModFile = deref(name), deref(bio), deref(mod)
	if formID == playerFormID {
		c.Group = GroupPlayer
	}
	return c, nil
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
