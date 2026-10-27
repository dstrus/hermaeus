package characters

import (
	"fmt"
	"strings"
	"testing"

	"hermaeus/internal/store"
	"hermaeus/internal/testenv"
)

func (h *harness) list() []Character {
	h.t.Helper()
	l, err := store.Read(h.store, testenv.SaveID, List)
	if err != nil {
		h.t.Fatal(err)
	}
	return l.Characters
}

// summary lists characters as "group:name(memories/diaries)", in order.
func summary(cs []Character) string {
	var parts []string
	for _, c := range cs {
		name := c.Name
		if name == "" {
			name = "#" + c.UUID
		}
		parts = append(parts, fmt.Sprintf("%s:%s(%d/%d)", c.Group, name, c.Memories, c.Diaries))
	}
	return strings.Join(parts, " ")
}

// memories gives an actor n memories.
func (h *harness) memories(uuid string, n int) {
	h.t.Helper()
	for range n {
		h.Exec(`insert into memories (actor_uuid, content) values (` + uuid + `, 'x')`)
	}
}

func TestListGroupsAndOrder(t *testing.T) {
	h := newHarness(t)
	h.Exec(`insert into uuid_mappings (uuid, form_id, actor_name, bio_template_name) values
		(2, 20, 'Gram', 'player_special'),
		(3, 3, 'aela', 'aela'),
		(4, 4, 'Brynjolf', 'brynjolf'),
		(5, 5, 'Lydia', 'lydia'),
		(6, 6, 'Nazeem', 'nazeem')`)
	h.memories("2", 1)              // the player, with few memories
	h.memories("3", BusyMemories)   // busy
	h.memories("4", BusyMemories-1) // not quite
	h.memories(testenv.SeranaUUID, 2)
	h.Exec(`insert into diary_entries (actor_uuid, content) values (` + testenv.SeranaUUID + `, 'Dear diary.')`)
	h.memories("99", 1) // SkyrimNet never named them
	h.memories("0", 3)  // world knowledge isn't anyone's
	h.Exec(`insert into memories (actor_uuid, content, pack_id) values (4, 'from a pack', 1)`)
	if err := h.setPinned("5", true); err != nil { // pinned, with nothing stored
		t.Fatal(err)
	}

	want := "player:Gram(1/0) pinned:Lydia(0/0) busy:aela(10/0) other:Brynjolf(9/0) other:Serana(2/1) other:#99(1/0)"
	if got := summary(h.list()); got != want {
		t.Errorf("list =\n %s\nwant\n %s", got, want)
	}
}

func TestListFoldsRetiredIdentities(t *testing.T) {
	h := newHarness(t)
	h.WithAlias()
	h.memories(testenv.SeranaUUID, 2)
	h.memories(testenv.MysteriousUUID, 1)
	h.Exec(`insert into diary_entries (actor_uuid, content) values (` + testenv.MysteriousUUID + `, 'Dear diary.')`)

	l := h.list()
	if got := summary(l); got != "other:Serana(3/1)" {
		t.Fatalf("list = %s", got)
	}
	if c := l[0]; c.UUID != testenv.SeranaUUID || fmt.Sprint(c.FormerNames) != "[Mysterious Woman]" {
		t.Errorf("Serana = %+v", c)
	}
}

func TestListShowsCurrentIdentityWithNothingOfTheirOwn(t *testing.T) {
	h := newHarness(t)
	h.WithAlias()
	h.memories(testenv.MysteriousUUID, 1) // all she has is under her old identity
	l := h.list()
	if got := summary(l); got != "other:Serana(1/0)" || l[0].UUID != testenv.SeranaUUID {
		t.Errorf("list = %s, %+v", got, l)
	}
}

func TestListOfEmptyPlaythrough(t *testing.T) {
	h := newHarness(t)
	h.Exec(`drop table identity_aliases; drop table npc_group_members`) // as in older databases
	if l := h.list(); len(l) != 0 {
		t.Errorf("list = %s", summary(l))
	}
}
