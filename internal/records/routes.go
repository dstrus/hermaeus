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
	// Every record of every kind, and one record's copies by layer.
	mux.Handle("GET /api/content", httpapi.JSON(func(*http.Request, httpapi.None) (ContentList, error) {
		return r.Content(), nil
	}))
	mux.Handle("GET /api/content/file/{path...}", httpapi.JSON(func(req *http.Request, _ httpapi.None) ([]Version, error) {
		return r.Versions(req.PathValue("path"))
	}))
	mux.Handle("PUT /api/content/file/{path...}", httpapi.JSON(func(req *http.Request, e RecordEdit) ([]Version, error) {
		return r.SaveRecord(req.PathValue("path"), e)
	}))
	// A record as its form edits it: an object, in the file's order.
	mux.Handle("GET /api/content/form/{path...}", httpapi.JSON(func(req *http.Request, _ httpapi.None) (RecordForm, error) {
		return r.Form(req.PathValue("path"))
	}))
	mux.Handle("PUT /api/content/form/{path...}", httpapi.JSON(func(req *http.Request, e FormEdit) (RecordForm, error) {
		return r.SaveForm(req.PathValue("path"), e)
	}))
	mux.Handle("PUT /api/content/enabled/{path...}", httpapi.JSON(func(req *http.Request, s Switch) ([]Version, error) {
		return r.SetEnabled(req.PathValue("path"), s)
	}))
	mux.Handle("DELETE /api/content/file/{path...}", httpapi.JSON(func(req *http.Request, rv Revert) ([]Version, error) {
		return r.RevertRecord(req.PathValue("path"), rv)
	}))
	// Voice effects, with the kinds of stage they're built from.
	mux.Handle("GET /api/voice-effects", httpapi.JSON(func(*http.Request, httpapi.None) (VoiceEffectList, error) {
		return r.VoiceEffects(), nil
	}))
	mux.Handle("PUT /api/voice-effects/{id}", httpapi.JSON(func(req *http.Request, e EffectEdit) (VoiceEffectList, error) {
		return r.SaveEffect(req.PathValue("id"), e)
	}))
	// Every file a layer above SkyrimNet's defaults replaces or adds.
	mux.Handle("GET /api/overrides", httpapi.JSON(func(*http.Request, httpapi.None) (OverrideList, error) {
		return r.Overrides(), nil
	}))
}
