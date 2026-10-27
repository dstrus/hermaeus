// Package calendar turns the game clock SkyrimNet records into dates as the
// game shows them, such as "7:54 AM, Sundas, 17th of Last Seed, 4E 201".
//
// The game clock (events' game_time, memories' game_time, diary entries'
// entry_date) counts seconds of game time. SkyrimNet also writes each event's
// date as the game shows it. The two don't always move together: a mod can
// set the game's date (Alternate Perspective, for one, adjusts it), and from
// then on the date the game shows is that many days off the clock. So the
// calendar follows the dates SkyrimNet wrote, in stretches: each stretch maps
// the clock to the date from where the game's date last jumped. Weekdays keep
// counting from the clock, as they do in the game.
package calendar

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"hermaeus/internal/store"
)

var months = []struct {
	name string
	days int
}{
	{"Morning Star", 31}, {"Sun's Dawn", 28}, {"First Seed", 31}, {"Rain's Hand", 30},
	{"Second Seed", 31}, {"Midyear", 30}, {"Sun's Height", 31}, {"Last Seed", 31},
	{"Hearthfire", 30}, {"Frostfall", 31}, {"Sun's Dusk", 30}, {"Evening Star", 31},
}

var weekdays = []string{"Sundas", "Morndas", "Tirdas", "Middas", "Turdas", "Fredas", "Loredas"}

const daySecs = 86400

// A date that differs from the current stretch's by more than this starts a
// new one. Dates are written to the minute, so they wobble by up to a minute.
const jump = 600

// maxClock is the largest reading taken as the game clock: over 31 years of
// game time. Larger values are real times (Unix seconds), which SkyrimNet
// stores in game_time for some world knowledge.
const maxClock = 1e9

// usualStart is the date at clock 0 in a new game, assumed when no event
// says otherwise.
const usualStart = "12:00 AM, Sundas, 17th of Last Seed, 4E 201"

// Calendar maps the game clock to a date: day n after 1 Morning Star of year
// 0 of the era starts at n*86400 seconds.
type Calendar struct {
	epoch       float64   // seconds from 1 Morning Star, year 0, to clock 0, at the start
	weekdayBase int       // weekday of day 0 by epoch, as an index into weekdays
	stretches   []stretch // by clock; the first also covers any earlier time
	era         string
	anchored    bool // false: no event to anchor on, so it assumes the usual start
}

// stretch: from clock from on, until the next stretch, the date is epoch + clock.
type stretch struct {
	from  float64
	epoch float64
}

// GameDate is a moment on the game clock, as the game shows it.
type GameDate struct {
	Clock float64 `json:"clock"` // seconds on the game clock
	Day   string  `json:"day"`   // "Sundas, 17th of Last Seed, 4E 201"
	Time  string  `json:"time"`  // "7:54 AM"
}

// Load reads the dates SkyrimNet wrote for events, in clock order, into
// stretches. Without any, it assumes the usual start: clock 0 is midnight on
// Sundas, 17th of Last Seed, 4E 201.
func Load(db store.DB) (*Calendar, error) {
	c := &Calendar{era: "4E"}
	if ok, err := store.HasTable(db, "events"); err != nil || !ok {
		c.add(0, usualStart)
		return c, err
	}
	rows, err := db.Query(`select game_time, game_time_str from events
		where game_time is not null and coalesce(game_time_str, '') <> '' order by game_time, id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var t float64
		var text string
		if err := rows.Scan(&t, &text); err != nil {
			return nil, err
		}
		c.add(t, text)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if c.anchored = len(c.stretches) > 0; !c.anchored {
		c.add(0, usualStart)
	}
	return c, nil
}

// add records that the game showed text at clock t, starting a new stretch
// if the date has jumped.
func (c *Calendar) add(t float64, text string) {
	secs, weekday, era, ok := parse(text)
	if !ok {
		return
	}
	// The game's day usually starts at a whole day on the clock, so snap to it
	// when the two agree to within a couple of minutes.
	epoch := secs - t
	if day := math.Round(epoch/daySecs) * daySecs; math.Abs(epoch-day) < 120 {
		epoch = day
	}
	if n := len(c.stretches); n == 0 {
		day := int(math.Floor(secs / daySecs))
		c.epoch, c.era = epoch, era
		c.weekdayBase = ((weekday-day)%7 + 7) % 7
	} else if math.Abs(epoch-c.stretches[n-1].epoch) <= jump {
		return
	}
	c.stretches = append(c.stretches, stretch{from: t, epoch: epoch})
}

// Date is the date the game showed at clock t. A clock of 0 or less, or one
// too large to be the game clock, means there's no in-game date.
func (c *Calendar) Date(t float64) (GameDate, bool) {
	if !(t > 0 && t <= maxClock) {
		return GameDate{}, false
	}
	st := c.stretches[0]
	for _, x := range c.stretches {
		if x.from > t {
			break
		}
		st = x
	}
	abs := st.epoch + t
	day := int(math.Floor(abs / daySecs))
	secs := int(abs) - day*daySecs
	year := day / 365
	d, m := day-year*365, 0
	for m < len(months)-1 && d >= months[m].days {
		d -= months[m].days
		m++
	}
	clockDay := int(math.Floor((c.epoch + t) / daySecs))
	weekday := weekdays[((c.weekdayBase+clockDay)%7+7)%7]
	hour, minute := secs/3600, secs%3600/60
	ampm := "AM"
	if hour >= 12 {
		ampm = "PM"
	}
	return GameDate{
		Clock: t,
		Day:   fmt.Sprintf("%s, %s of %s, %s %d", weekday, ordinal(d+1), months[m].name, c.era, year),
		Time:  fmt.Sprintf("%d:%02d %s", (hour+11)%12+1, minute, ampm),
	}, true
}

// DatePtr is Date for a response field that's absent when there's no date.
func (c *Calendar) DatePtr(t float64) *GameDate {
	if d, ok := c.Date(t); ok {
		return &d
	}
	return nil
}

func ordinal(n int) string {
	suffix := "th"
	if n%100 < 11 || n%100 > 13 {
		switch n % 10 {
		case 1:
			suffix = "st"
		case 2:
			suffix = "nd"
		case 3:
			suffix = "rd"
		}
	}
	return strconv.Itoa(n) + suffix
}

var dateRe = regexp.MustCompile(`^(\d{1,2}):(\d{2}) (AM|PM), (\w+), (\d{1,2})(?:st|nd|rd|th) of ([A-Za-z' ]+), (\d)E (\d+)$`)

// parse reads SkyrimNet's date text, such as "7:54 AM, Sundas, 17th of Last
// Seed, 4E 201", returning the seconds since 1 Morning Star of year 0 and the
// weekday's index.
func parse(s string) (secs float64, weekday int, era string, ok bool) {
	m := dateRe.FindStringSubmatch(strings.TrimSpace(s))
	if m == nil {
		return 0, 0, "", false
	}
	hour, _ := strconv.Atoi(m[1])
	minute, _ := strconv.Atoi(m[2])
	day, _ := strconv.Atoi(m[5])
	year, _ := strconv.Atoi(m[8])
	if hour < 1 || hour > 12 || minute > 59 {
		return 0, 0, "", false
	}
	hour %= 12
	if m[3] == "PM" {
		hour += 12
	}
	weekday = -1
	for i, w := range weekdays {
		if strings.EqualFold(w, m[4]) {
			weekday = i
		}
	}
	days := year * 365
	month := -1
	for i, mo := range months {
		if strings.EqualFold(mo.name, m[6]) {
			month = i
			break
		}
		days += mo.days
	}
	if weekday < 0 || month < 0 || day < 1 || day > months[month].days {
		return 0, 0, "", false
	}
	days += day - 1
	return float64(days)*daySecs + float64(hour*3600+minute*60), weekday, m[7] + "E", true
}
