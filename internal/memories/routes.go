package memories

import (
	"fmt"
	"net/http"
	"strconv"

	"hermaeus/internal/apperr"
	"hermaeus/internal/httpapi"
	"hermaeus/internal/store"
)

func Routes(mux *http.ServeMux, st *store.Store) {
	const base = "/api/playthroughs/{save}"

	mux.Handle("GET "+base+"/memories", httpapi.JSON(func(r *http.Request, _ httpapi.None) (MemoriesResponse, error) {
		return store.Read(st, r.PathValue("save"), All)
	}))
	mux.Handle("GET "+base+"/knowledge", httpapi.JSON(func(r *http.Request, _ httpapi.None) (KnowledgeResponse, error) {
		return store.Read(st, r.PathValue("save"), Knowledge)
	}))
	mux.Handle("POST "+base+"/knowledge", httpapi.JSON(func(r *http.Request, e NewEntry) (SaveResult, error) {
		return store.Write(st, r.PathValue("save"), "Added a world-knowledge entry", func(tx store.DB) (SaveResult, error) {
			return AddKnowledge(tx, e)
		})
	}))
	mux.Handle("PUT "+base+"/knowledge/packs/{id}", httpapi.JSON(func(r *http.Request, c PackChanges) (KnowledgeResponse, error) {
		id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
		if err != nil {
			return KnowledgeResponse{}, apperr.New(apperr.Invalid, "Invalid pack id %q", r.PathValue("id"))
		}
		what := "Switched off knowledge pack " + r.PathValue("id")
		if c.Enabled {
			what = "Switched on knowledge pack " + r.PathValue("id")
		}
		return store.Write(st, r.PathValue("save"), what, func(tx store.DB) (KnowledgeResponse, error) {
			if err := SetPackEnabled(tx, id, c.Enabled); err != nil {
				return KnowledgeResponse{}, err
			}
			return Knowledge(tx)
		})
	}))
	mux.Handle("GET "+base+"/diary", httpapi.JSON(func(r *http.Request, _ httpapi.None) (DiaryResponse, error) {
		return store.Read(st, r.PathValue("save"), AllDiary)
	}))
	mux.Handle("GET "+base+"/characters/{uuid}/memories", httpapi.JSON(func(r *http.Request, _ httpapi.None) (MemoriesResponse, error) {
		return store.Read(st, r.PathValue("save"), func(db store.DB) (MemoriesResponse, error) {
			return ForCharacter(db, r.PathValue("uuid"))
		})
	}))
	mux.Handle("POST "+base+"/characters/{uuid}/memories", httpapi.JSON(func(r *http.Request, m MemoryChanges) (SaveResult, error) {
		return store.Write(st, r.PathValue("save"), "Added a memory for "+r.PathValue("uuid"), func(tx store.DB) (SaveResult, error) {
			return AddMemory(tx, r.PathValue("uuid"), m)
		})
	}))
	mux.Handle("PATCH "+base+"/memories/{id}", httpapi.JSON(func(r *http.Request, c MemoryChanges) (SaveResult, error) {
		id, err := pathID(r)
		if err != nil {
			return SaveResult{}, err
		}
		return store.Write(st, r.PathValue("save"), fmt.Sprintf("Saved memory %d", id), func(tx store.DB) (SaveResult, error) {
			return UpdateMemory(tx, id, c, true)
		})
	}))
	mux.Handle("POST "+base+"/memories/update", httpapi.JSON(func(r *http.Request, b BulkChanges) (Count, error) {
		return store.Write(st, r.PathValue("save"), fmt.Sprintf("Updated memories %v", b.IDs), func(tx store.DB) (Count, error) {
			return UpdateMemories(tx, b)
		})
	}))
	mux.Handle("POST "+base+"/memories/delete", httpapi.JSON(func(r *http.Request, ids IDs) (Count, error) {
		return store.Write(st, r.PathValue("save"), fmt.Sprintf("Deleted memories %v", ids.IDs), func(tx store.DB) (Count, error) {
			return DeleteMemories(tx, ids.IDs)
		})
	}))
	mux.Handle("POST "+base+"/memories/move", httpapi.JSON(func(r *http.Request, ids IDs) (Count, error) {
		return store.Write(st, r.PathValue("save"), fmt.Sprintf("Moved memories %v to their NPC's current identity", ids.IDs), func(tx store.DB) (Count, error) {
			return MoveToCurrentIdentity(tx, ids.IDs)
		})
	}))

	mux.Handle("GET "+base+"/characters/{uuid}/diary", httpapi.JSON(func(r *http.Request, _ httpapi.None) (DiaryResponse, error) {
		return store.Read(st, r.PathValue("save"), func(db store.DB) (DiaryResponse, error) {
			return Diary(db, r.PathValue("uuid"))
		})
	}))
	mux.Handle("POST "+base+"/characters/{uuid}/diary", httpapi.JSON(func(r *http.Request, c DiaryChanges) (SaveResult, error) {
		return store.Write(st, r.PathValue("save"), "Added a diary entry for "+r.PathValue("uuid"), func(tx store.DB) (SaveResult, error) {
			return AddDiaryEntry(tx, r.PathValue("uuid"), c)
		})
	}))
	mux.Handle("PATCH "+base+"/diary/{id}", httpapi.JSON(func(r *http.Request, c DiaryChanges) (SaveResult, error) {
		id, err := pathID(r)
		if err != nil {
			return SaveResult{}, err
		}
		return store.Write(st, r.PathValue("save"), fmt.Sprintf("Saved diary entry %d", id), func(tx store.DB) (SaveResult, error) {
			return UpdateDiary(tx, id, c)
		})
	}))
	mux.Handle("POST "+base+"/diary/delete", httpapi.JSON(func(r *http.Request, d DeleteDiary) (DeletedDiary, error) {
		return store.Write(st, r.PathValue("save"), fmt.Sprintf("Deleted diary entries %v", d.IDs), func(tx store.DB) (DeletedDiary, error) {
			return DeleteDiaryEntries(tx, d)
		})
	}))
}

func pathID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		return 0, apperr.New(apperr.Invalid, "Invalid ID %q", r.PathValue("id"))
	}
	return id, nil
}
