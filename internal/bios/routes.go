package bios

import (
	"net/http"

	"hermaeus/internal/httpapi"
)

func (b *Bios) Routes(mux *http.ServeMux) {
	mux.Handle("GET /api/bios", httpapi.JSON(func(*http.Request, httpapi.None) (Catalog, error) {
		return b.CatalogOf()
	}))
	const base = "/api/playthroughs/{save}/bios"
	mux.Handle("GET "+base, httpapi.JSON(func(r *http.Request, _ httpapi.None) (SaveBios, error) {
		return b.List(r.PathValue("save"))
	}))
	mux.Handle("GET "+base+"/{template}", httpapi.JSON(func(r *http.Request, _ httpapi.None) (BioResponse, error) {
		return b.Read(r.PathValue("save"), r.PathValue("template"))
	}))
	mux.Handle("POST "+base+"/{template}", httpapi.JSON(func(r *http.Request, _ httpapi.None) (Version, error) {
		return b.Create(r.PathValue("save"), r.PathValue("template"))
	}))
	mux.Handle("PATCH "+base+"/{template}/{kind}", httpapi.JSON(func(r *http.Request, e Edit) (Version, error) {
		return b.Save(r.PathValue("save"), r.PathValue("template"), r.PathValue("kind"), e)
	}))
	mux.Handle("GET "+base+"/{template}/{kind}/history", httpapi.JSON(func(r *http.Request, _ httpapi.None) (HistoryResponse, error) {
		return b.History(r.PathValue("save"), r.PathValue("template"), r.PathValue("kind"))
	}))
	mux.Handle("POST "+base+"/{template}/{kind}/restore", httpapi.JSON(func(r *http.Request, rs Restore) (Version, error) {
		return b.RestoreEarlier(r.PathValue("save"), r.PathValue("template"), r.PathValue("kind"), rs)
	}))
}
