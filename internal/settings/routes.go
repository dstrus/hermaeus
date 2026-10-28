package settings

import (
	"net/http"

	"hermaeus/internal/httpapi"
)

func (s *Settings) Routes(mux *http.ServeMux) {
	mux.Handle("GET /api/settings", httpapi.JSON(func(*http.Request, httpapi.None) (FileList, error) {
		return s.List()
	}))
	mux.Handle("GET /api/settings/{name}", httpapi.JSON(func(r *http.Request, _ httpapi.None) (FileValues, error) {
		return s.Read(r.PathValue("name"))
	}))
	mux.Handle("PATCH /api/settings/{name}", httpapi.JSON(func(r *http.Request, e Edit) (FileValues, error) {
		return s.Save(r.PathValue("name"), e)
	}))
	mux.Handle("GET /api/toggles", httpapi.JSON(func(*http.Request, httpapi.None) (ToggleList, error) {
		return s.Toggles()
	}))
	mux.Handle("PUT /api/toggles/{id}", httpapi.JSON(func(r *http.Request, req ToggleRequest) (ToggleList, error) {
		return s.SetToggle(r.PathValue("id"), req.On)
	}))
}
