package httpapi

import (
	"net/http"
	"os"
)

// File sends a file from disk, with range requests (so audio can be
// sought) and revalidation by modification time, so a browser that has it
// asks again cheaply instead of downloading it again.
func File(w http.ResponseWriter, r *http.Request, path, contentType string) {
	f, err := os.Open(path)
	if err != nil {
		Send(w, http.StatusNotFound, Error{"File not found"})
		return
	}
	defer f.Close()
	info, err := f.Stat()
	if err != nil || !info.Mode().IsRegular() {
		Send(w, http.StatusNotFound, Error{"File not found"})
		return
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "private, no-cache")
	http.ServeContent(w, r, "", info.ModTime(), f)
}
