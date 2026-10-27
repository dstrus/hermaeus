// Package server is Hermaeus's web server. It serves the app's pages, and
// mounts each domain's API routes behind the checks every request needs:
// addressed to this machine, and carrying the token from the page.
package server

import (
	"bytes"
	"crypto/subtle"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"path"
	"strings"

	"hermaeus/internal/apperr"
	"hermaeus/internal/characters"
	"hermaeus/internal/httpapi"
	"hermaeus/internal/prefs"
	"hermaeus/internal/store"
)

// tokenPlaceholder is replaced with the token in index.html (see app.html).
var tokenPlaceholder = []byte("__HERMAEUS_TOKEN__")

type Options struct {
	Store   *store.Store
	Prefs   *prefs.Prefs
	Token   string
	Port    int
	Files   fs.FS // the built frontend, with index.html at its root
	Version string
}

type Server struct {
	Options
	index []byte // index.html, with the token filled in
	api   *http.ServeMux
}

func New(o Options) (*Server, error) {
	index, err := fs.ReadFile(o.Files, "index.html")
	if err != nil {
		return nil, fmt.Errorf("the frontend isn't built: %w", err)
	}
	s := &Server{Options: o, index: bytes.ReplaceAll(index, tokenPlaceholder, []byte(o.Token)), api: http.NewServeMux()}
	s.routes()
	return s, nil
}

func (s *Server) routes() {
	m := s.api
	m.Handle("GET /api/status", httpapi.JSON(func(*http.Request, httpapi.None) (Status, error) {
		return s.status(), nil
	}))
	m.Handle("PUT /api/folder", httpapi.JSON(func(_ *http.Request, req FolderRequest) (Status, error) {
		dir, err := s.Store.SetDir(req.Path)
		if err != nil {
			return Status{}, err
		}
		log.Printf("SkyrimNet folder: %s", dir)
		s.Prefs.Update(func(v *prefs.Values) { v.SkyrimNetDir = dir })
		return s.status(), nil
	}))

	characters.Routes(m, s.Store)

	m.Handle("/api/", httpapi.JSON(func(r *http.Request, _ httpapi.None) (httpapi.None, error) {
		return httpapi.None{}, apperr.New(apperr.NotFound, "Unknown endpoint: %s %s", r.Method, r.URL.Path)
	}))
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Only answer requests addressed to this machine by name (blocks DNS-rebinding tricks).
	if r.Host != fmt.Sprintf("localhost:%d", s.Port) && r.Host != fmt.Sprintf("127.0.0.1:%d", s.Port) {
		httpapi.Send(w, http.StatusForbidden, httpapi.Error{Error: "Forbidden"})
		return
	}
	switch {
	case r.URL.Path == "/api/ping":
		// Needs no token: it only says that this is Hermaeus.
		httpapi.Send(w, http.StatusOK, Ping{App: AppID})
	case strings.HasPrefix(r.URL.Path, "/api/"):
		s.guard(w, r)
	default:
		s.page(w, r)
	}
}

// guard requires the per-session token embedded in the page, so other
// websites open in the same browser can't read or change anything.
func (s *Server) guard(w http.ResponseWriter, r *http.Request) {
	if subtle.ConstantTimeCompare([]byte(r.Header.Get(TokenHeader)), []byte(s.Token)) != 1 {
		httpapi.Send(w, http.StatusForbidden, httpapi.Error{Error: "Missing or invalid Hermaeus token. Reload the page."})
		return
	}
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("X-Content-Type-Options", "nosniff")
	s.api.ServeHTTP(w, r)
}

// page serves the frontend's files. Any other path is a view inside the app
// (every view has its own address), so it gets index.html.
func (s *Server) page(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("X-Content-Type-Options", "nosniff")
	name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
	if name != "" && name != "index.html" {
		if info, err := fs.Stat(s.Files, name); err == nil && !info.IsDir() {
			if strings.HasPrefix(name, "_app/immutable/") {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			}
			http.ServeFileFS(w, r, s.Files, name)
			return
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-store") // it carries this session's token
	w.Write(s.index)
}
