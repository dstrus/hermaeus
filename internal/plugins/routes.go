package plugins

import (
	"net/http"

	"hermaeus/internal/httpapi"
)

func (p *Plugins) Routes(mux *http.ServeMux) {
	mux.Handle("GET /api/plugins", httpapi.JSON(func(*http.Request, httpapi.None) (List, error) {
		return p.List()
	}))
	mux.Handle("PUT /api/plugins/order", httpapi.JSON(func(_ *http.Request, o Order) (List, error) {
		return p.SetOrder(o)
	}))
	mux.Handle("PUT /api/plugins/{id}/enabled", httpapi.JSON(func(r *http.Request, s Switch) (List, error) {
		return p.SetEnabled(r.PathValue("id"), s)
	}))
}
