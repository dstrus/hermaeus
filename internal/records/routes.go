package records

import (
	"net/http"

	"hermaeus/internal/httpapi"
)

func (r *Records) Routes(mux *http.ServeMux) {
	mux.Handle("GET /api/actions", httpapi.JSON(func(*http.Request, httpapi.None) (ActionList, error) {
		return r.Actions(), nil
	}))
	mux.Handle("GET /api/triggers", httpapi.JSON(func(*http.Request, httpapi.None) (TriggerList, error) {
		return r.Triggers(), nil
	}))
}
