package calendar

import (
	"fmt"
	"testing"

	"hermaeus/internal/store"
	"hermaeus/internal/testenv"
)

func load(t *testing.T, f *testenv.Folder) *Calendar {
	t.Helper()
	st := store.New("", func() bool { return false })
	if _, err := st.SetDir(f.Dir); err != nil {
		t.Fatal(err)
	}
	c, err := store.Read(st, testenv.SaveID, Load)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func date(c *Calendar, t float64) string {
	d, ok := c.Date(t)
	if !ok {
		return "none"
	}
	return d.Time + ", " + d.Day
}

func TestParse(t *testing.T) {
	secs, weekday, era, ok := parse("7:54 AM, Sundas, 17th of Last Seed, 4E 201")
	// 201 years, Morning Star through Sun's Height (212 days), 16 days, 7:54.
	if want := float64((201*365+212+16)*86400 + 7*3600 + 54*60); !ok || secs != want || weekday != 0 || era != "4E" {
		t.Errorf("parse = %v, %d, %q, %v; want %v, Sundas", secs, weekday, era, ok, want)
	}
	if secs, _, _, _ := parse("12:30 PM, Sundas, 1st of Morning Star, 4E 0"); secs != 12*3600+30*60 {
		t.Errorf("12:30 PM = %v seconds into the day", secs)
	}
	for _, bad := range []string{"", "13:00 AM, Sundas, 1st of Last Seed, 4E 201", "1:00 AM, Sundas, 32nd of Last Seed, 4E 201", "1:00 AM, Someday, 1st of Last Seed, 4E 201", "1:00 AM, Sundas, 1st of Lastseed, 4E 201"} {
		if _, _, _, ok := parse(bad); ok {
			t.Errorf("parsed %q", bad)
		}
	}
}

func TestUsualStartWithoutDatedEvents(t *testing.T) {
	f := testenv.NewFolder(t)
	f.Exec(`insert into events (event_type, event_data, game_time) values ('dialogue', '{}', 1234.5)`)
	c := load(t, f)
	if c.anchored {
		t.Error("anchored without any dated event")
	}
	for clock, want := range map[float64]string{
		1:          "12:00 AM, Sundas, 17th of Last Seed, 4E 201",
		28440:      "7:54 AM, Sundas, 17th of Last Seed, 4E 201",
		43200:      "12:00 PM, Sundas, 17th of Last Seed, 4E 201",
		86400 + 60: "12:01 AM, Morndas, 18th of Last Seed, 4E 201",
		// 15 days later is the 1st of Hearthfire, after Last Seed's 31 days.
		15*86400 + 3600: "1:00 AM, Morndas, 1st of Hearthfire, 4E 201",
		// No date: SkyrimNet didn't record one, or it's a real time.
		0: "none", -5: "none", 1791685675.11973: "none",
	} {
		if got := date(c, clock); got != want {
			t.Errorf("clock %v = %q, want %q", clock, got, want)
		}
	}
}

func TestDatesFollowTheGamesDate(t *testing.T) {
	f := testenv.NewFolder(t)
	// A mod set the game's date back a day on the morning of the 1st of
	// Hearthfire (clock day 15), as Alternate Perspective did in one
	// playthrough; the clock carried on.
	f.Exec(`insert into events (event_type, event_data, game_time, game_time_str) values
		('sleep_start', '{}', 28495.84, '7:54 AM, Sundas, 17th of Last Seed, 4E 201'),
		('sleep_stop', '{}', 1315339.84, '5:22 AM, Morndas, 1st of Hearthfire, 4E 201'),
		('sleep_start', '{}', 1325580, '8:13 AM, Morndas, 31st of Last Seed, 4E 201'),
		('sleep_stop', '{}', 1400000, '4:53 AM, Tirdas, 1st of Hearthfire, 4E 201')`)
	c := load(t, f)
	start := float64((201*365 + 212 + 16) * 86400) // midnight, 17th of Last Seed
	if !c.anchored || c.epoch != start || fmt.Sprint(c.stretches) != fmt.Sprintf("[{28495.84 %v} {1.32558e+06 %v}]", start, start-86400) {
		t.Fatalf("epoch %v, stretches %v; want the date set back a day from 8:13 AM on clock day 15", c.epoch, c.stretches)
	}
	for clock, want := range map[float64]string{
		28495.84: "7:54 AM, Sundas, 17th of Last Seed, 4E 201",
		// Before the first dated event, the first stretch applies.
		60:         "12:01 AM, Sundas, 17th of Last Seed, 4E 201",
		1315339.84: "5:22 AM, Morndas, 1st of Hearthfire, 4E 201",
		// After the jump the date is a day behind, but the weekday carries on.
		1325580: "8:13 AM, Morndas, 31st of Last Seed, 4E 201",
		1400000: "4:53 AM, Tirdas, 1st of Hearthfire, 4E 201",
	} {
		if got := date(c, clock); got != want {
			t.Errorf("clock %v = %q, want %q", clock, got, want)
		}
	}
}

func TestOrdinal(t *testing.T) {
	for n, want := range map[int]string{1: "1st", 2: "2nd", 3: "3rd", 4: "4th", 11: "11th", 12: "12th", 13: "13th", 21: "21st", 22: "22nd", 23: "23rd", 31: "31st"} {
		if got := ordinal(n); got != want {
			t.Errorf("ordinal(%d) = %q, want %q", n, got, want)
		}
	}
}
