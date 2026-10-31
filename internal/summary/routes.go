package summary

import (
	"net/http"

	"hermaeus/internal/httpapi"
	"hermaeus/internal/store"
)

func Routes(mux *http.ServeMux, st *store.Store) {
	mux.Handle("GET /api/playthroughs/{save}/summary", httpapi.JSON(func(r *http.Request, _ httpapi.None) (Summary, error) {
		return store.Read(st, r.PathValue("save"), Of)
	}))
}
