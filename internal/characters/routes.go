package characters

import (
	"net/http"

	"hermaeus/internal/httpapi"
	"hermaeus/internal/store"
)

func Routes(mux *http.ServeMux, st *store.Store) {
	mux.Handle("GET /api/playthroughs/{save}/pins", httpapi.JSON(func(r *http.Request, _ httpapi.None) (PinsResponse, error) {
		return store.Read(st, r.PathValue("save"), Pins)
	}))
	mux.Handle("PUT /api/playthroughs/{save}/pins/{uuid}", setPinned(st, true))
	mux.Handle("DELETE /api/playthroughs/{save}/pins/{uuid}", setPinned(st, false))
}

// setPinned pins or unpins the NPC in the path, and replies with the pins
// as they are afterward.
func setPinned(st *store.Store, pinned bool) http.Handler {
	return httpapi.JSON(func(r *http.Request, _ httpapi.None) (PinsResponse, error) {
		uuid := r.PathValue("uuid")
		what := "Unpinned NPC " + uuid
		if pinned {
			what = "Pinned NPC " + uuid
		}
		return store.Write(st, r.PathValue("save"), what, func(tx store.DB) (PinsResponse, error) {
			if err := SetPinned(tx, uuid, pinned); err != nil {
				return PinsResponse{}, err
			}
			return Pins(tx)
		})
	})
}
