// Package events is about what SkyrimNet records happening in the game:
// dialogue, deaths, books read, trades… Each event has a type, its details
// (JSON, or plain text for some types), a place, the game clock, and the
// actors involved.
package events

import (
	"regexp"
	"strings"
)

var acronyms = regexp.MustCompile(`(?i)\b(npc|stt|tton)\b`)

// Label turns an event type such as "follower_recruited" into words for
// people: "Follower recruited". Acronyms SkyrimNet uses stay capitalized.
func Label(eventType string) string {
	s := strings.ReplaceAll(eventType, "_", " ")
	s = acronyms.ReplaceAllStringFunc(s, strings.ToUpper)
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}
