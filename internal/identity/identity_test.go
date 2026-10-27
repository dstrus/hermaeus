package identity

import (
	"fmt"
	"sort"
	"testing"

	"hermaeus/internal/store"
	"hermaeus/internal/testenv"
)

func load(t *testing.T, f *testenv.Folder) *Aliases {
	t.Helper()
	st := store.New("", func() bool { return false })
	if _, err := st.SetDir(f.Dir); err != nil {
		t.Fatal(err)
	}
	a, err := store.Read(st, testenv.SaveID, Load)
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func TestChainsResolveToTheNewestIdentity(t *testing.T) {
	f := testenv.NewFolder(t)
	// 1 became 2, which became 3; 7 and 8 point at each other, so neither is newest.
	f.Exec(`insert into identity_aliases (old_uuid, new_uuid, old_name) values
		(1, 2, 'Stranger'), (2, 3, 'Mysterious Woman'), (7, 8, 'A'), (8, 7, 'B'), (9, 9, 'Self')`)
	a := load(t, f)
	for uuid, want := range map[string]string{"1": "3", "2": "3", "3": "3", "7": "7", "8": "8", "9": "9", "42": "42"} {
		if got := a.Resolve(uuid); got != want {
			t.Errorf("Resolve(%s) = %s, want %s", uuid, got, want)
		}
	}
	all := a.All("3")
	sort.Strings(all[1:])
	if fmt.Sprint(all) != "[3 1 2]" {
		t.Errorf("All(3) = %v", all)
	}
	if name, retired := a.Retired("2"); !retired || name != "Mysterious Woman" {
		t.Errorf("Retired(2) = %q, %v", name, retired)
	}
	if _, retired := a.Retired("3"); retired {
		t.Error("the newest identity counts as retired")
	}
}

func TestDatabaseWithoutAliases(t *testing.T) {
	f := testenv.NewFolder(t)
	f.Exec(`drop table identity_aliases`)
	if a := load(t, f); a.Resolve("1") != "1" {
		t.Error("resolved without any aliases")
	}
}
