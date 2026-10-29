package characters

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"

	"hermaeus/internal/calendar"
	"hermaeus/internal/events"
	"hermaeus/internal/store"
)

// A character's events are everything SkyrimNet recorded them taking part
// in: what they said and heard, who they killed, what they read, traded,
// ate… Each is summed up in a sentence for the list, with its details for
// the pane.

type Event struct {
	ID      int64  `json:"id"`
	Type    string `json:"type"`
	Label   string `json:"label"`   // the type, in words
	Summary string `json:"summary"` // what happened, in a sentence
	// For something said: who said it, to whom, and the words, so the page
	// can set the names apart from the words.
	Speaker      string             `json:"speaker,omitempty"`
	Listener     string             `json:"listener,omitempty"`
	Line         string             `json:"line,omitempty"`
	SpeakerUUID  string             `json:"speakerUuid,omitempty"`  // the speaker's current identity, when the event names one
	ListenerUUID string             `json:"listenerUuid,omitempty"` // likewise the listener's
	Location     string             `json:"location,omitempty"`
	Date         *calendar.GameDate `json:"date,omitempty"`
	Written      *int64             `json:"written,omitempty"` // when SkyrimNet recorded it, Unix ms
	// The character's part in it: "acting", "acted on", or "present".
	Role    string         `json:"role"`
	Present int            `json:"present"`        // how many characters were there
	Data    map[string]any `json:"data,omitempty"` // the details SkyrimNet stored, when they're fields
	Text    string         `json:"text,omitempty"` // the details, when they're a sentence instead
}

type EventsResponse struct {
	Events []Event          `json:"events"` // newest first
	Types  []EventTypeCount `json:"types"`  // how many of each type, most first
}

type EventTypeCount struct {
	Type  string `json:"type"`
	Label string `json:"label"`
	Count int    `json:"count"`
}

// Past this many characters, a detail (a book's whole text, say) is cut
// short for the page.
const detailText = 2000

// EventsOf lists a character's events.
func EventsOf(db store.DB, uuid string) (EventsResponse, error) {
	out := EventsResponse{Events: []Event{}, Types: []EventTypeCount{}}
	p, err := loadPerson(db, uuid)
	if err != nil {
		return out, err
	}
	if ok, err := hasEvents(db); err != nil || !ok {
		return out, err
	}
	cal, err := calendar.Load(db)
	if err != nil {
		return out, err
	}
	mine := map[string]bool{}
	for _, u := range p.uuids {
		mine[u] = true
	}
	theirs, args := p.theirEvents()
	rows, err := db.Query(theirs+`select id, event_type, coalesce(event_data, ''), coalesce(location, ''), game_time, local_time,
		coalesce(actor_UUIDs, ''), cast(originating_actor_UUID as text), cast(target_actor_UUID as text)
		from theirs order by game_time is null, game_time desc, id desc`, args...)
	if err != nil {
		return out, err
	}
	defer rows.Close()
	counts := map[string]int{}
	for rows.Next() {
		var e Event
		var data, actors, origin, target string
		var t sql.NullFloat64
		var written any
		if err := rows.Scan(&e.ID, &e.Type, &data, &e.Location, &t, &written, &actors, &origin, &target); err != nil {
			return out, err
		}
		e.Label = events.Label(e.Type)
		e.Date = cal.DatePtr(t.Float64)
		e.Written = store.Timestamp(written)
		e.Present = len(uuidStrings(actors))
		switch {
		case mine[origin]:
			e.Role = "acting"
		case mine[target]:
			e.Role = "acted on"
		default:
			e.Role = "present"
		}
		e.Data, e.Text = details(data)
		e.Summary = summarize(e.Type, e.Data, e.Text)
		e.Speaker, e.Listener, e.Line = said(e.Type, e.Data)
		if e.Speaker != "" && origin != "0" {
			e.SpeakerUUID = p.aliases.Resolve(origin)
		}
		if e.Listener != "" && target != "0" {
			e.ListenerUUID = p.aliases.Resolve(target)
		}
		counts[e.Type]++
		out.Events = append(out.Events, e)
	}
	if err := rows.Err(); err != nil {
		return out, err
	}
	for t, n := range counts {
		out.Types = append(out.Types, EventTypeCount{Type: t, Label: events.Label(t), Count: n})
	}
	sort.Slice(out.Types, func(i, j int) bool {
		if out.Types[i].Count != out.Types[j].Count {
			return out.Types[i].Count > out.Types[j].Count
		}
		return out.Types[i].Type < out.Types[j].Type
	})
	return out, nil
}

// details reads an event's stored details: fields when they're JSON, or
// the sentence SkyrimNet wrote instead. Long text is cut short.
func details(data string) (map[string]any, string) {
	var m map[string]any
	if json.Unmarshal([]byte(data), &m) != nil || m == nil {
		return nil, shorten(data)
	}
	for k, v := range m {
		if s, ok := v.(string); ok {
			m[k] = shorten(s)
		}
	}
	return m, ""
}

func shorten(s string) string {
	if len(s) <= detailText {
		return s
	}
	n := detailText
	for n > 0 && !utf8.RuneStart(s[n]) {
		n--
	}
	return s[:n] + "…"
}

// uuidStrings reads a JSON list of actor UUIDs, which are beyond
// JavaScript's number precision, as strings.
func uuidStrings(s string) []string {
	dec := json.NewDecoder(strings.NewReader(s))
	dec.UseNumber()
	var list []any
	if dec.Decode(&list) != nil {
		return nil
	}
	out := make([]string, 0, len(list))
	for _, v := range list {
		switch x := v.(type) {
		case json.Number:
			out = append(out, x.String())
		case string:
			out = append(out, x)
		}
	}
	return out
}

// summarize says what an event was, in a sentence, from what SkyrimNet
// stores for its type.
func summarize(eventType string, d map[string]any, text string) string {
	if d == nil {
		if text != "" {
			return text
		}
		return events.Label(eventType)
	}
	str := func(keys ...string) string {
		for _, k := range keys {
			if s, _ := d[k].(string); s != "" {
				return s
			}
		}
		return ""
	}
	switch eventType {
	case "dialogue", "dialogue_npc", "dialogue_background", "dialogue_player", "dialogue_player_stt", "gamemaster_dialogue":
		who, to, line := said(eventType, d)
		if who != "" && to != "" {
			who += " to " + to
		}
		if who != "" && line != "" {
			return fmt.Sprintf("%s: “%s”", who, line)
		}
		if line != "" {
			return line
		}
	case "death":
		killer, victim := str("killer"), str("victim")
		switch {
		case killer != "" && victim != "":
			return killer + " killed " + victim
		case victim != "":
			return victim + " died"
		}
	case "book_read":
		if b := str("book_name"); b != "" {
			return "Read " + b
		}
	case "diary_entry_created":
		if who := str("actor"); who != "" {
			if entry := str("entry_contents"); entry != "" {
				return who + " wrote in their diary: " + entry
			}
			return who + " wrote in their diary"
		}
	case "sleep_start":
		if who := str("actor"); who != "" {
			return who + " went to sleep"
		}
	case "sleep_stop":
		if who := str("actor"); who != "" {
			return who + " woke up"
		}
	case "trade_start":
		if m := str("merchant"); m != "" {
			return "Started trading with " + m
		}
	case "trade_complete":
		return milestoneSummary(eventType, mustJSON(d))
	case "tton_event":
		who := str("mainActorNames")
		if with := str("secondaryActorNames"); with != "" {
			who += " and " + with
		}
		if msg := str("msg"); msg != "" && who != "" {
			return who + ": " + msg
		}
	}
	if s := str("dialogue", "narration", "line", "text", "msg", "summary", "description", "message", "content"); s != "" {
		return s
	}
	return events.Label(eventType)
}

// said is who spoke, to whom, and what, for the dialogue event types; all
// empty for anything else.
func said(eventType string, d map[string]any) (who, to, line string) {
	switch eventType {
	case "dialogue", "dialogue_npc", "dialogue_background", "dialogue_player", "dialogue_player_stt", "gamemaster_dialogue":
	default:
		return "", "", ""
	}
	str := func(keys ...string) string {
		for _, k := range keys {
			if s, _ := d[k].(string); s != "" {
				return s
			}
		}
		return ""
	}
	return str("speaker", "player_name"), str("listener", "target"), str("dialogue")
}

func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}
