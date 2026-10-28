package prompts

import (
	"net/http"

	"hermaeus/internal/httpapi"
)

func (p *Prompts) Routes(mux *http.ServeMux) {
	mux.Handle("GET /api/prompts", httpapi.JSON(func(*http.Request, httpapi.None) (PromptList, error) {
		return p.List()
	}))
	mux.Handle("GET /api/prompts/search", httpapi.JSON(func(r *http.Request, _ httpapi.None) (Matches, error) {
		return p.Search(r.URL.Query().Get("q"))
	}))
	// A prompt's path under prompts/, such as components/event_history.prompt.
	mux.Handle("GET /api/prompts/file/{path...}", httpapi.JSON(func(r *http.Request, _ httpapi.None) (Document, error) {
		return p.Read(r.PathValue("path"))
	}))
	mux.Handle("PUT /api/prompts/file/{path...}", httpapi.JSON(func(r *http.Request, e Edit) (Saved, error) {
		return p.Save(r.PathValue("path"), e)
	}))
	mux.Handle("PUT /api/prompts/pins/{path...}", httpapi.JSON(func(r *http.Request, _ httpapi.None) (PinsResponse, error) {
		return p.SetPinned(r.PathValue("path"), true)
	}))
	mux.Handle("DELETE /api/prompts/pins/{path...}", httpapi.JSON(func(r *http.Request, _ httpapi.None) (PinsResponse, error) {
		return p.SetPinned(r.PathValue("path"), false)
	}))
	mux.Handle("DELETE /api/prompts/file/{path...}", httpapi.JSON(func(r *http.Request, rv Revert) (Document, error) {
		return p.RemoveCopy(r.PathValue("path"), rv)
	}))
}
