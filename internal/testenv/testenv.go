// Package testenv builds SkyrimNet folders for tests, with a database that
// has SkyrimNet's schema (schema.sql) and whatever rows a test adds.
package testenv

import (
	"database/sql"
	_ "embed"
	"os"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schema string

const (
	SaveID = "test"

	SeranaUUID     = "561512192010693353"  // beyond JavaScript's safe integer range
	MysteriousUUID = "1618829674175326866" // Serana's identity before her name was known
)

// Folder is a SkyrimNet folder with one playthrough, SaveID.
type Folder struct {
	t      *testing.T
	Dir    string
	DBPath string
}

// NewFolder makes a SkyrimNet folder whose playthrough knows Serana.
func NewFolder(t *testing.T) *Folder {
	t.Helper()
	// MO2 folder names often include brackets, which file patterns would misread.
	dir := filepath.Join(t.TempDir(), "[NoDelete] SkyrimNet")
	f := &Folder{t: t, Dir: dir, DBPath: filepath.Join(dir, "data", "SkyrimNet-"+SaveID+".db")}
	if err := os.MkdirAll(filepath.Dir(f.DBPath), 0o755); err != nil {
		t.Fatal(err)
	}
	f.Exec(schema + `
insert into uuid_mappings (uuid, form_id, actor_name, bio_template_name) values (` + SeranaUUID + `, 1, 'Serana', 'serana');`)
	return f
}

// WithAlias records that Serana was first known as "Mysterious Woman",
// under another UUID.
func (f *Folder) WithAlias() {
	f.t.Helper()
	f.Exec(`
insert into uuid_mappings (uuid, form_id, actor_name, bio_template_name) values (` + MysteriousUUID + `, 1, 'Serana', 'serana');
insert into identity_aliases (old_uuid, new_uuid, old_name, new_name) values (` + MysteriousUUID + `, ` + SeranaUUID + `, 'Mysterious Woman', 'Serana');`)
}

// Exec runs SQL on the playthrough's database, outside Hermaeus's store.
func (f *Folder) Exec(query string, args ...any) {
	f.t.Helper()
	db := f.open()
	defer db.Close()
	if _, err := db.Exec(query, args...); err != nil {
		f.t.Fatal(err)
	}
}

// Query returns the first column of the first row, or nil if there's none.
func (f *Folder) Query(query string, args ...any) any {
	f.t.Helper()
	db := f.open()
	defer db.Close()
	var v any
	if err := db.QueryRow(query, args...).Scan(&v); err != nil && err != sql.ErrNoRows {
		f.t.Fatal(err)
	}
	return v
}

func (f *Folder) open() *sql.DB {
	f.t.Helper()
	db, err := sql.Open("sqlite", f.DBPath)
	if err != nil {
		f.t.Fatal(err)
	}
	return db
}
