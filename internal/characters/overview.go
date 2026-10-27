package characters

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"hermaeus/internal/calendar"
	"hermaeus/internal/events"
	"hermaeus/internal/identity"
	"hermaeus/internal/store"
)

// Overview is who a character is in the game's files and in SkyrimNet, and
// the turns their story has taken. It counts their retired identities as
// theirs.
type Overview struct {
	UUID        string             `json:"uuid"`
	Name        string             `json:"name"`
	Bio         string             `json:"bio"`                 // their bio template's name, if any
	FormID      string             `json:"formId"`              // in hex, as the game's console shows it; "" if unknown
	ModFile     string             `json:"modFile"`             // the plugin that adds them, if known
	LocalFormID string             `json:"localFormId"`         // their ID within ModFile, in hex
	FirstSeen   *int64             `json:"firstSeen,omitempty"` // when SkyrimNet first recorded them, Unix ms
	Updated     *int64             `json:"updated,omitempty"`   // when SkyrimNet last updated their record, Unix ms
	Retired     []RetiredIdentity  `json:"retired"`
	FirstEvent  *calendar.GameDate `json:"firstEvent,omitempty"`
	LastEvent   *calendar.GameDate `json:"lastEvent,omitempty"`
	Milestones  []Milestone        `json:"milestones"` // in game order
}

// RetiredIdentity is an identity SkyrimNet replaced with the character's
// current one, such as "Mysterious Woman" before Serana was named.
type RetiredIdentity struct {
	UUID  string `json:"uuid"`
	Name  string `json:"name"`
	Since *int64 `json:"since,omitempty"` // when SkyrimNet re-identified them, Unix ms
}

// Milestone is an event that marks a turn in a character's story, such as
// joining the player.
type Milestone struct {
	EventID int64              `json:"eventId"`
	Type    string             `json:"type"`
	Label   string             `json:"label"`   // the type, in words
	Summary string             `json:"summary"` // what happened, in a sentence
	Date    *calendar.GameDate `json:"date,omitempty"`
}

// Event types that are milestones.
var milestoneTypes = []any{"follower_recruited", "follower_dismissed", "item_given", "trade_complete", "brawl_ended"}

func OverviewOf(db store.DB, uuid string) (Overview, error) {
	o := Overview{UUID: uuid, Retired: []RetiredIdentity{}, Milestones: []Milestone{}}
	p, err := loadPerson(db, uuid)
	if err != nil {
		return o, err
	}
	n, _ := strconv.ParseInt(uuid, 10, 64)
	var name, bio, mod sql.NullString
	var formID, localFormID sql.NullInt64
	var created, updated any
	err = db.QueryRow(`select actor_name, bio_template_name, mod_file, form_id, local_form_id, created_at, updated_at
		from uuid_mappings where uuid = ?`, n).Scan(&name, &bio, &mod, &formID, &localFormID, &created, &updated)
	if err != nil && !store.IsNoRows(err) {
		return o, err
	}
	o.Name, o.Bio, o.ModFile = name.String, bio.String, mod.String
	if formID.Valid {
		o.FormID = fmt.Sprintf("%08X", formID.Int64)
	}
	if localFormID.Valid {
		o.LocalFormID = fmt.Sprintf("%03X", localFormID.Int64)
	}
	o.FirstSeen, o.Updated = store.Timestamp(created), store.Timestamp(updated)

	if len(p.uuids) > 1 {
		in, ids := identity.SQLIn(p.uuids[1:])
		rows, err := db.Query(`select cast(old_uuid as text), coalesce(old_name, ''), created_at from identity_aliases
			where old_uuid in (`+in+`) order by id`, ids...)
		if err != nil {
			return o, err
		}
		for rows.Next() {
			var r RetiredIdentity
			var since any
			if err := rows.Scan(&r.UUID, &r.Name, &since); err != nil {
				rows.Close()
				return o, err
			}
			r.Since = store.Timestamp(since)
			o.Retired = append(o.Retired, r)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return o, err
		}
	}

	if ok, err := hasEvents(db); err != nil || !ok {
		return o, err
	}
	cal, err := calendar.Load(db)
	if err != nil {
		return o, err
	}
	theirs, args := p.theirEvents()
	var first, last sql.NullFloat64
	if err := db.QueryRow(theirs+`select min(game_time), max(game_time) from theirs where game_time > 0`, args...).Scan(&first, &last); err != nil {
		return o, err
	}
	o.FirstEvent, o.LastEvent = cal.DatePtr(first.Float64), cal.DatePtr(last.Float64)

	types := strings.TrimSuffix(strings.Repeat("?,", len(milestoneTypes)), ",")
	rows, err := db.Query(theirs+`select id, event_type, game_time, coalesce(event_data, '') from theirs
		where event_type in (`+types+`) order by game_time, id`, append(args, milestoneTypes...)...)
	if err != nil {
		return o, err
	}
	defer rows.Close()
	for rows.Next() {
		var m Milestone
		var t sql.NullFloat64
		var data string
		if err := rows.Scan(&m.EventID, &m.Type, &t, &data); err != nil {
			return o, err
		}
		m.Label = events.Label(m.Type)
		m.Summary = milestoneSummary(m.Type, data)
		m.Date = cal.DatePtr(t.Float64)
		o.Milestones = append(o.Milestones, m)
	}
	return o, rows.Err()
}

// milestoneSummary says what a milestone event was. Most types store a
// sentence; trades store JSON.
func milestoneSummary(eventType, data string) string {
	var d map[string]any
	if json.Unmarshal([]byte(data), &d) != nil {
		if data == "" {
			return events.Label(eventType)
		}
		return data
	}
	if eventType != "trade_complete" {
		return events.Label(eventType)
	}
	var parts []string
	if s, _ := d["items_bought"].(string); s != "" {
		parts = append(parts, "bought "+s)
	}
	if s, _ := d["items_sold"].(string); s != "" {
		parts = append(parts, "sold "+s)
	}
	if len(parts) == 0 {
		parts = append(parts, "nothing listed")
	}
	merchant, _ := d["merchant"].(string)
	if merchant == "" {
		merchant = "a merchant"
	}
	s := fmt.Sprintf("Traded with %s: %s", merchant, strings.Join(parts, "; "))
	if net, _ := d["net_gold"].(float64); net != 0 {
		sign := ""
		if net > 0 {
			sign = "+"
		}
		s += fmt.Sprintf(" (%s%s gold)", sign, thousands(int64(net)))
	}
	return s
}

// thousands writes n with commas between groups of three digits.
func thousands(n int64) string {
	s := strconv.FormatInt(n, 10)
	sign := ""
	if n < 0 {
		sign, s = "-", s[1:]
	}
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return sign + s
}
