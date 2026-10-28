package settings

import (
	"path/filepath"
	"strings"

	"hermaeus/internal/apperr"
)

// Quick toggles switch SkyrimNet's main features on and off, as the quick
// toggles in SkyrimNet's dashboard do. Each is one setting; switching it
// goes through Save, so the file is backed up and nothing changes while
// Skyrim runs. The dashboard also has one for actions, but where it keeps
// that switch isn't known, so it's left out here.

type Toggle struct {
	ID     string `json:"id"`
	Label  string `json:"label"` // short, for a row of them
	What   string `json:"what"`  // a sentence saying what it switches
	File   string `json:"file"`  // settings name, as in File
	Key    string `json:"key"`
	Master bool   `json:"master,omitempty"` // when off, SkyrimNet makes no AI calls at all

	On        bool  `json:"on"`
	Available bool  `json:"available"` // its file exists, so it can be switched here
	modified  int64 // the file's time, for saving
}

var toggles = []Toggle{
	{ID: "skyrimnet", Label: "AI", What: "SkyrimNet’s AI altogether. Off, SkyrimNet makes no AI calls at all.", File: "game", Key: "general.globalAIEnabled", Master: true},
	{ID: "gamemaster", Label: "GameMaster", What: "The GameMaster agent, which sets scenes and moves the story along.", File: "game", Key: "gamemaster.agentEnabled"},
	{ID: "npc_reactions", Label: "NPC reactions", What: "NPCs reacting to what happens around them.", File: "Events", Key: "global.npcReactionsEnabled"},
	{ID: "player_reactions", Label: "Player reactions", What: "AI-written lines spoken by your own character.", File: "PlayerDialogue", Key: "enabled"},
	{ID: "narration", Label: "Narration", What: "A narrator describing what happens.", File: "game", Key: "narration.enabled"},
	{ID: "bard_singing", Label: "Bards", What: "Bards singing AI-written songs.", File: "BardSinging", Key: "bard_singing.enabled"},
	{ID: "universal_translator", Label: "Translator", What: "The Universal Translator, for speaking with NPCs in other languages.", File: "UniversalTranslator", Key: "enabled"},
	{ID: "dynamic_bios", Label: "Dynamic bios", What: "SkyrimNet rewriting characters’ dynamic bios as you play.", File: "DynamicBio", Key: "enabled"},
	{ID: "memory_generation", Label: "Memories", What: "Making new memories from events, in the background.", File: "Memory", Key: "performance.enable_background_processing"},
	{ID: "omnisight", Label: "OmniSight", What: "OmniSight: NPCs seeing what’s on screen.", File: "OmniSight", Key: "omnisight.enabled"},
}

type ToggleList struct {
	Toggles []Toggle `json:"toggles"`
}

type ToggleRequest struct {
	On bool `json:"on"`
}

// Toggles reads the quick toggles' current state. One whose file doesn't
// exist yet shows its default, but can't be switched.
func (s *Settings) Toggles() (ToggleList, error) {
	dir, err := s.configDir()
	if err != nil {
		return ToggleList{}, err
	}
	defs := loadDefaults(dir)
	files := map[string]FileValues{}
	out := ToggleList{Toggles: make([]Toggle, 0, len(toggles))}
	for _, t := range toggles {
		f, ok := files[t.File]
		if !ok {
			if f, err = read(dir, t.File, defs); err != nil {
				return ToggleList{}, err
			}
			files[t.File] = f
		}
		t.Available, t.modified = f.File.Exists, f.File.Modified
		for _, v := range f.Values {
			if strings.EqualFold(v.Key, t.Key) {
				val := v.Default
				if v.Set {
					val = v.Value
				}
				t.On = val == true
			}
		}
		out.Toggles = append(out.Toggles, t)
	}
	return out, nil
}

// SetToggle switches one quick toggle on or off.
func (s *Settings) SetToggle(id string, on bool) (ToggleList, error) {
	list, err := s.Toggles()
	if err != nil {
		return ToggleList{}, err
	}
	for _, t := range list.Toggles {
		if t.ID != id {
			continue
		}
		if !t.Available {
			dir, _ := s.configDir()
			return ToggleList{}, apperr.New(apperr.NotFound, "%s doesn't exist yet. SkyrimNet creates it the next time it starts.", filepath.Base(filePath(dir, t.File)))
		}
		if _, err := s.Save(t.File, Edit{Modified: t.modified, Changes: []Change{{Key: t.Key, Value: on}}}); err != nil {
			return ToggleList{}, err
		}
		return s.Toggles()
	}
	return ToggleList{}, apperr.New(apperr.NotFound, "There's no quick toggle %q", id)
}
