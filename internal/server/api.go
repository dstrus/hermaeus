package server

import "hermaeus/internal/store"

// The server's own requests and replies; each domain declares its own.

// AppID identifies Hermaeus to /api/ping, so a second copy can find this one.
const AppID = "hermaeus"

// TokenHeader carries the per-session token on every API request.
const TokenHeader = "X-Hermaeus-Token"

type Ping struct {
	App string `json:"app"`
}

// Status is what the page needs to know before showing anything: where
// SkyrimNet's files are, and which playthroughs they hold.
type Status struct {
	SkyrimNetDir string              `json:"skyrimNetDir"` // "" until one is found or chosen
	BackupDir    string              `json:"backupDir"`
	GameRunning  bool                `json:"gameRunning"`
	Playthroughs []store.Playthrough `json:"playthroughs"` // most recently played first
	Version      string              `json:"version"`      // Hermaeus's own
	Theme        string              `json:"theme"`        // "system", "dark" or "light"
	// SkyrimNet's mod folder, when its defaults aren't in the SkyrimNet
	// folder (Mod Organizer 2); "" otherwise.
	ModDir         string `json:"modDir"`
	ModDirDetected bool   `json:"modDirDetected"` // found automatically rather than chosen
	DefaultBios    bool   `json:"defaultBios"`    // SkyrimNet's default bios were found
}

type FolderRequest struct {
	Path    string  `json:"path"`
	ModPath *string `json:"modPath,omitempty"` // SkyrimNet's mod folder under MO2; "" goes back to finding it
}

type ThemeRequest struct {
	Theme string `json:"theme"` // "system", "dark" or "light"
}

func (s *Server) status() Status {
	st := Status{
		SkyrimNetDir: s.Store.Dir(),
		BackupDir:    s.Store.BackupDir(),
		GameRunning:  s.Store.GameRunning(),
		Playthroughs: s.Store.Playthroughs(),
		Version:      s.Version,
		Theme:        s.Prefs.Get().ThemeOf(),
		DefaultBios:  s.Library.HasDefaults(),
	}
	st.ModDir, st.ModDirDetected = s.Library.ModDir()
	return st
}
