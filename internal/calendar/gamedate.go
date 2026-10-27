package calendar

// GameDate is a moment on the game clock, as the game shows it.
type GameDate struct {
	Clock float64 `json:"clock"` // seconds on the game clock
	Day   string  `json:"day"`   // "Sundas, 17th of Last Seed, 4E 201"
	Time  string  `json:"time"`  // "7:54 AM"
}
