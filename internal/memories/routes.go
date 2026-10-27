package memories

import (
	"net/http"

	"hermaeus/internal/httpapi"
	"hermaeus/internal/store"
)

func Routes(mux *http.ServeMux, st *store.Store) {
	mux.Handle("GET /api/playthroughs/{save}/characters/{uuid}/memories", httpapi.JSON(func(r *http.Request, _ httpapi.None) (MemoriesResponse, error) {
		return store.Read(st, r.PathValue("save"), func(db store.DB) (MemoriesResponse, error) {
			return ForCharacter(db, r.PathValue("uuid"))
		})
	}))
	mux.Handle("GET /api/playthroughs/{save}/characters/{uuid}/diary", httpapi.JSON(func(r *http.Request, _ httpapi.None) (DiaryResponse, error) {
		return store.Read(st, r.PathValue("save"), func(db store.DB) (DiaryResponse, error) {
			return Diary(db, r.PathValue("uuid"))
		})
	}))
}
