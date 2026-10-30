package songs

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"hermaeus/internal/apperr"
	"hermaeus/internal/httpapi"
	"hermaeus/internal/store"
)

func Routes(mux *http.ServeMux, st *store.Store) {
	const base = "/api/playthroughs/{save}/songs"
	mux.Handle("GET "+base, httpapi.JSON(func(r *http.Request, _ httpapi.None) (List, error) {
		return store.Read(st, r.PathValue("save"), func(db store.DB) (List, error) { return All(db, st.Dir()) })
	}))
	mux.Handle("PATCH "+base+"/{id}", httpapi.JSON(func(r *http.Request, c Changes) (List, error) {
		id, err := pathID(r)
		if err != nil {
			return List{}, err
		}
		return store.Write(st, r.PathValue("save"), fmt.Sprintf("Saved song %d", id), func(tx store.DB) (List, error) {
			if err := Update(tx, id, c); err != nil {
				return List{}, err
			}
			return All(tx, st.Dir())
		})
	}))
	mux.Handle("DELETE "+base+"/{id}", httpapi.JSON(func(r *http.Request, _ httpapi.None) (List, error) {
		id, err := pathID(r)
		if err != nil {
			return List{}, err
		}
		return store.Write(st, r.PathValue("save"), fmt.Sprintf("Deleted song %d", id), func(tx store.DB) (List, error) {
			if err := Delete(tx, id); err != nil {
				return List{}, err
			}
			return All(tx, st.Dir())
		})
	}))
	mux.HandleFunc("GET "+base+"/{id}/audio", func(w http.ResponseWriter, r *http.Request) {
		id, err := pathID(r)
		if err != nil {
			httpapi.SendError(w, err)
			return
		}
		name, err := store.Read(st, r.PathValue("save"), func(db store.DB) (string, error) { return StoredFile(db, id) })
		if err != nil {
			httpapi.SendError(w, err)
			return
		}
		f := AudioFile(st.Dir(), name)
		if f == "" {
			httpapi.SendError(w, apperr.New(apperr.NotFound, "Song %d’s audio isn’t in data/bard_songs", id))
			return
		}
		httpapi.File(w, r, f, AudioTypes[strings.ToLower(filepath.Ext(f))])
	})
}

func pathID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		return 0, apperr.New(apperr.Invalid, "Invalid song id %q", r.PathValue("id"))
	}
	return id, nil
}
