package filters

import (
	"net/http"

	"hermaeus/internal/httpapi"
)

// Created is the lists after a new one is added, and the new one.
type Created struct {
	Lists   ListsResponse `json:"lists"`
	Created List          `json:"created"`
}

func (f *Filters) Routes(mux *http.ServeMux) {
	// A filter's lists, by kind: memory or actor.
	mux.Handle("GET /api/filters/{kind}", httpapi.JSON(func(r *http.Request, _ httpapi.None) (ListsResponse, error) {
		return f.Lists(r.PathValue("kind"))
	}))
	mux.Handle("POST /api/filters/{kind}", httpapi.JSON(func(r *http.Request, e ListEdit) (Created, error) {
		lists, list, err := f.Create(r.PathValue("kind"), e)
		return Created{Lists: lists, Created: list}, err
	}))
	mux.Handle("PUT /api/filters/{kind}/{id}", httpapi.JSON(func(r *http.Request, e ListEdit) (ListsResponse, error) {
		return f.Save(r.PathValue("kind"), r.PathValue("id"), e)
	}))
	mux.Handle("DELETE /api/filters/{kind}/{id}", httpapi.JSON(func(r *http.Request, rm Removal) (ListsResponse, error) {
		return f.Remove(r.PathValue("kind"), r.PathValue("id"), rm)
	}))
}
