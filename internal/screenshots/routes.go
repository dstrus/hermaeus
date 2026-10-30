package screenshots

import (
	"bytes"
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
	thumbs := &Thumbs{}
	const base = "/api/playthroughs/{save}/screenshots"
	mux.Handle("GET "+base, httpapi.JSON(func(r *http.Request, _ httpapi.None) (List, error) {
		return store.Read(st, r.PathValue("save"), func(db store.DB) (List, error) { return All(db, st.Dir()) })
	}))
	mux.Handle("PATCH "+base+"/{id}", httpapi.JSON(func(r *http.Request, c Changes) (List, error) {
		id, err := pathID(r)
		if err != nil {
			return List{}, err
		}
		return store.Write(st, r.PathValue("save"), fmt.Sprintf("Saved screenshot %d", id), func(tx store.DB) (List, error) {
			if err := Update(tx, id, c); err != nil {
				return List{}, err
			}
			return All(tx, st.Dir())
		})
	}))
	mux.HandleFunc("GET "+base+"/{id}/image", func(w http.ResponseWriter, r *http.Request) {
		id, err := pathID(r)
		if err != nil {
			httpapi.SendError(w, err)
			return
		}
		stored, err := store.Read(st, r.PathValue("save"), func(db store.DB) (string, error) { return StoredFile(db, id) })
		if err != nil {
			httpapi.SendError(w, err)
			return
		}
		f := ImageFile(st.Dir(), stored)
		if f == "" {
			httpapi.SendError(w, apperr.New(apperr.NotFound, "Screenshot %d’s image isn’t in the SkyrimNet folder", id))
			return
		}
		// ?size=thumb asks for a small copy, for a grid of them.
		if r.URL.Query().Get("size") == "thumb" {
			if data, modified, ok := thumbs.Of(f); ok {
				w.Header().Set("Content-Type", "image/jpeg")
				w.Header().Set("Cache-Control", "private, no-cache")
				http.ServeContent(w, r, "", modified, bytes.NewReader(data))
				return
			}
		}
		httpapi.File(w, r, f, ImageTypes[strings.ToLower(filepath.Ext(f))])
	})
}

func pathID(r *http.Request) (int64, error) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		return 0, apperr.New(apperr.Invalid, "Invalid screenshot id %q", r.PathValue("id"))
	}
	return id, nil
}
