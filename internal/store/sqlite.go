package store

import (
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"

	_ "modernc.org/sqlite"
)

// DB is what domain code reads and writes through: a database opened for
// reading, or a transaction when writing.
type DB interface {
	Query(query string, args ...any) (*sql.Rows, error)
	QueryRow(query string, args ...any) *sql.Row
	Exec(query string, args ...any) (sql.Result, error)
}

// HasTable reports whether the database has the table. Older SkyrimNet
// databases lack some that newer ones have.
func HasTable(db DB, name string) (bool, error) {
	var ok bool
	err := db.QueryRow(`select count(*) > 0 from sqlite_master where type = 'table' and name = ?`, name).Scan(&ok)
	return ok, err
}

func IsNoRows(err error) bool { return errors.Is(err, sql.ErrNoRows) }

func open(path string, readOnly bool) (*sql.DB, error) {
	return openTimeout(path, readOnly, 5000)
}

func openTimeout(path string, readOnly bool, busyMs int) (*sql.DB, error) {
	params := url.Values{"_busy_timeout": {strconv.Itoa(busyMs)}}
	if readOnly {
		params.Set("mode", "ro")
	} else {
		params.Set("mode", "rw")
		params.Set("_txlock", "immediate")
	}
	db, err := sql.Open("sqlite", sqliteURI(path, params))
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("cannot open database: %w", err)
	}
	return db, nil
}

// sqliteURI builds a file: URI so SQLite honors mode=ro / mode=rw. Unlike a
// plain file name, mode=rw never creates an empty database if the file is gone.
func sqliteURI(path string, params url.Values) string {
	p := filepath.ToSlash(path)
	if !strings.HasPrefix(p, "/") {
		p = "/" + p // Windows drive paths: file:///C:/...
	}
	return (&url.URL{Scheme: "file", Path: p, RawQuery: params.Encode()}).String()
}

// HasColumn reports whether the table has the column. SkyrimNet adds
// columns to its tables over time, so older databases lack some.
func HasColumn(db DB, table, column string) (bool, error) {
	var ok bool
	err := db.QueryRow(`select count(*) > 0 from pragma_table_info(?) where name = ?`, table, column).Scan(&ok)
	return ok, err
}
