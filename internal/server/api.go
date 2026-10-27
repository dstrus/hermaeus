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
}

type FolderRequest struct {
	Path string `json:"path"`
}

func (s *Server) status() Status {
	return Status{
		SkyrimNetDir: s.Store.Dir(),
		BackupDir:    s.Store.BackupDir(),
		GameRunning:  s.Store.GameRunning(),
		Playthroughs: s.Store.Playthroughs(),
		Version:      s.Version,
	}
}
