// Package prefs remembers Hermaeus's own settings between runs, such as the
// SkyrimNet folder chosen in the browser, so it doesn't have to be chosen
// again. SkyrimNet's settings are a different thing entirely.
package prefs

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sync"
)

type Values struct {
	SkyrimNetDir string `json:"skyrimNetDir,omitempty"`
}

// Prefs holds the remembered values and saves every change to its file.
type Prefs struct {
	mu   sync.Mutex
	path string // "" keeps them in memory only
	v    Values
}

// Path is where the preferences are kept, in the OS's usual place for
// program settings, or "" if it doesn't have one.
func Path() string {
	dir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "Hermaeus", "settings.json")
}

// Load reads the preferences at path. A missing or unreadable file means
// nothing is remembered yet.
func Load(path string) *Prefs {
	p := &Prefs{path: path}
	if path != "" {
		if b, err := os.ReadFile(path); err == nil {
			json.Unmarshal(b, &p.v)
		}
	}
	return p
}

func (p *Prefs) Get() Values {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.v
}

// Update changes the values and saves them. Failing to save is logged
// rather than returned: it only means the choice won't be remembered.
func (p *Prefs) Update(change func(*Values)) {
	p.mu.Lock()
	defer p.mu.Unlock()
	change(&p.v)
	if p.path == "" {
		return
	}
	b, _ := json.MarshalIndent(p.v, "", "  ")
	err := os.MkdirAll(filepath.Dir(p.path), 0o755)
	if err == nil {
		err = os.WriteFile(p.path, b, 0o644)
	}
	if err != nil {
		log.Printf("Could not save settings for next time: %v", err)
	}
}
