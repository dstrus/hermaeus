package memories

import (
	"database/sql"
	"strings"

	"hermaeus/internal/apperr"
	"hermaeus/internal/store"
)

// World knowledge is what any NPC can draw on: facts and lore that belong to
// no character. SkyrimNet keeps them as memories with no actor, most of them
// in knowledge packs (knowledge_packs, with the pack's id on each entry);
// a pack can be switched off as a whole.

type Pack struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
	Enabled     bool   `json:"enabled"`
	Source      string `json:"source,omitempty"` // the .sknpack file it was loaded from
	Entries     int    `json:"entries"`
}

type KnowledgeResponse struct {
	Packs   []Pack   `json:"packs"`   // by name
	Entries []Memory `json:"entries"` // by title, with Pack set for those in a pack
}

// Knowledge lists the packs and every world-knowledge entry.
func Knowledge(db store.DB) (KnowledgeResponse, error) {
	out := KnowledgeResponse{Packs: []Pack{}, Entries: []Memory{}}
	if ok, err := store.HasTable(db, "knowledge_packs"); err != nil || !ok {
		return out, err
	}
	rows, err := db.Query(`select p.id, p.name, coalesce(p.description, ''), coalesce(p.version, ''), coalesce(p.is_enabled, 1),
			coalesce(p.source_path, ''), (select count(*) from memories m where m.pack_id = p.id)
		from knowledge_packs p order by p.name collate nocase`)
	if err != nil {
		return out, err
	}
	for rows.Next() {
		var p Pack
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Version, &p.Enabled, &p.Source, &p.Entries); err != nil {
			rows.Close()
			return out, err
		}
		out.Packs = append(out.Packs, p)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return out, err
	}
	p, err := everyone(db)
	if err != nil {
		return out, err
	}
	list, err := listMemories(db, p, "(m.pack_id is not null or coalesce(m.actor_uuid, 0) = 0)", nil)
	if err != nil {
		return out, err
	}
	out.Entries = list.Memories
	// By title, as the dashboard lists them; a pack's own order is its ids.
	sortEntries(out.Entries)
	return out, nil
}

func sortEntries(entries []Memory) {
	less := func(i, j int) bool {
		a, b := strings.ToLower(entries[i].Title), strings.ToLower(entries[j].Title)
		if a != b {
			return a < b
		}
		return entries[i].ID < entries[j].ID
	}
	for i := 1; i < len(entries); i++ { // insertion sort keeps it simple and stable
		for j := i; j > 0 && less(j, j-1); j-- {
			entries[j], entries[j-1] = entries[j-1], entries[j]
		}
	}
}

type PackChanges struct {
	Enabled bool `json:"enabled"`
}

// SetPackEnabled switches a knowledge pack on or off.
func SetPackEnabled(tx store.DB, id int64, enabled bool) error {
	r, err := tx.Exec(`update knowledge_packs set is_enabled = ?, updated_at = datetime('now') where id = ?`, enabled, id)
	if err != nil {
		return err
	}
	if n, _ := r.RowsAffected(); n == 0 {
		return apperr.New(apperr.NotFound, "Knowledge pack %d not found", id)
	}
	return nil
}

// NewEntry is a world-knowledge entry to add, to a pack or to none.
type NewEntry struct {
	Pack    *int64        `json:"pack,omitempty"`
	Changes MemoryChanges `json:"changes"`
}

// AddKnowledge adds a world-knowledge entry: a memory of no character,
// of type KNOWLEDGE, in the pack if one is named.
func AddKnowledge(tx store.DB, e NewEntry) (SaveResult, error) {
	if e.Changes.Content == nil || strings.TrimSpace(*e.Changes.Content) == "" {
		return SaveResult{}, apperr.New(apperr.Invalid, "An entry’s text can’t be empty")
	}
	if e.Pack != nil {
		var n int
		if err := tx.QueryRow(`select count(*) from knowledge_packs where id = ?`, *e.Pack).Scan(&n); err != nil {
			return SaveResult{}, err
		}
		if n == 0 {
			return SaveResult{}, apperr.New(apperr.NotFound, "Knowledge pack %d not found", *e.Pack)
		}
	}
	sets, args, err := e.Changes.columns()
	if err != nil {
		return SaveResult{}, err
	}
	t, err := latestGameTime(tx)
	if err != nil {
		return SaveResult{}, err
	}
	var pack sql.NullInt64
	if e.Pack != nil {
		pack = sql.NullInt64{Int64: *e.Pack, Valid: true}
	}
	r, err := tx.Exec(`insert into memories (actor_uuid, content, game_time, creation_time, related_event_ids, related_actors,
			memory_type, importance_score, tags, always_inject, is_active, pack_id)
		values (0, '', ?, `+nowSQL+`, '[]', '[]', 'KNOWLEDGE', 0.8, '[]', 0, 1, ?)`, t, pack)
	if err != nil {
		return SaveResult{}, err
	}
	id, err := r.LastInsertId()
	if err != nil {
		return SaveResult{}, err
	}
	if _, err := tx.Exec(`update memories set `+strings.Join(sets, ", ")+` where id = ?`, append(args, id)...); err != nil {
		return SaveResult{}, err
	}
	return SaveResult{ID: id, Reembed: true}, nil
}
