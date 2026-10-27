package characters

import (
	"fmt"
	"testing"

	"hermaeus/internal/apperr"
	"hermaeus/internal/store"
	"hermaeus/internal/testenv"
)

type harness struct {
	*testenv.Folder
	t       *testing.T
	running bool
	store   *store.Store
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	h := &harness{Folder: testenv.NewFolder(t), t: t}
	h.store = store.New("", func() bool { return h.running })
	if _, err := h.store.SetDir(h.Dir); err != nil {
		t.Fatal(err)
	}
	return h
}

func (h *harness) pins() string {
	h.t.Helper()
	p, err := store.Read(h.store, testenv.SaveID, Pins)
	if err != nil {
		h.t.Fatal(err)
	}
	return fmt.Sprint(p.Pins)
}

func (h *harness) setPinned(uuid string, pinned bool) error {
	_, err := store.Write(h.store, testenv.SaveID, "pin", func(tx store.DB) (bool, error) {
		return true, SetPinned(tx, uuid, pinned)
	})
	return err
}

func TestPins(t *testing.T) {
	h := newHarness(t)
	h.WithAlias()
	if p := h.pins(); p != "[]" {
		t.Fatalf("pins = %s, want none", p)
	}

	// The database has no Pinned group yet, so pinning creates SkyrimNet's.
	for range 2 { // pinning twice is the same as once
		if err := h.setPinned(testenv.SeranaUUID, true); err != nil {
			t.Fatal(err)
		}
	}
	if d := h.Query(`select description from npc_groups where group_name = 'Pinned'`); d != pinnedGroupDesc {
		t.Errorf("group description = %v", d)
	}
	want := "[{" + testenv.SeranaUUID + " Serana}]"
	if p := h.pins(); p != want {
		t.Errorf("pins = %s, want %s", p, want)
	}

	// A pin the dashboard left on her retired identity counts as hers, and
	// unpinning her removes it too.
	h.Exec(`insert into npc_group_members (group_id, actor_uuid) select id, ` + testenv.MysteriousUUID + ` from npc_groups where group_name = 'Pinned'`)
	if p := h.pins(); p != want {
		t.Errorf("pins = %s, want Serana once", p)
	}
	if err := h.setPinned(testenv.SeranaUUID, false); err != nil {
		t.Fatal(err)
	}
	if n := h.Query(`select count(*) from npc_group_members`); n != int64(0) {
		t.Errorf("%v pins left after unpinning", n)
	}
}

func TestPinRefusals(t *testing.T) {
	h := newHarness(t)
	for uuid, want := range map[string]apperr.Kind{"0": apperr.Invalid, "serana": apperr.Invalid, "": apperr.Invalid} {
		if k, _ := apperr.KindOf(h.setPinned(uuid, true)); k != want {
			t.Errorf("pinning %q: kind %v, want %v", uuid, k, want)
		}
	}
	h.running = true
	if k, _ := apperr.KindOf(h.setPinned(testenv.SeranaUUID, true)); k != apperr.Conflict {
		t.Errorf("pinning while Skyrim runs: kind %v", k)
	}
}

func TestUnpinningWithoutGroupChangesNothing(t *testing.T) {
	h := newHarness(t)
	if err := h.setPinned(testenv.SeranaUUID, false); err != nil {
		t.Fatal(err)
	}
	if n := h.Query(`select count(*) from npc_groups`); n != int64(0) {
		t.Errorf("unpinning created %v groups", n)
	}
}

func TestDatabaseWithoutGroups(t *testing.T) {
	h := newHarness(t)
	h.Exec(`drop table npc_group_members`)
	if p := h.pins(); p != "[]" {
		t.Errorf("pins = %s", p)
	}
	if k, _ := apperr.KindOf(h.setPinned(testenv.SeranaUUID, true)); k != apperr.Conflict {
		t.Errorf("pinning without groups: kind %v", k)
	}
}
