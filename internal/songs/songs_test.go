package songs

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"hermaeus/internal/apperr"
	"hermaeus/internal/store"
	"hermaeus/internal/testenv"
)

func TestSongs(t *testing.T) {
	f := testenv.NewFolder(t)
	s := store.New("", func() bool { return false })
	if _, err := s.SetDir(f.Dir); err != nil {
		t.Fatal(err)
	}
	f.Exec(`insert into bard_songs (id, title, lyrics, style_tags, gender, audio_filename, game_time_created, composer_uuid, composer_name, play_count) values
  (1, 'Ragnar', 'la la', 'folk', 'male', 'song.mp3', 0, ` + testenv.SeranaUUID + `, 'Serana', 2),
  (2, 'Gone', 'oh', 'dirge', 'female', '../escape.mp3', 0, 0, '', 0)`)
	os.MkdirAll(filepath.Join(f.Dir, "data", "bard_songs"), 0o755)
	os.WriteFile(filepath.Join(f.Dir, "data", "bard_songs", "song.mp3"), []byte("0123456789"), 0o644)
	os.WriteFile(filepath.Join(f.Dir, "data", "escape.mp3"), nil, 0o644)

	all := func() List {
		l, err := store.Read(s, testenv.SaveID, func(db store.DB) (List, error) { return All(db, s.Dir()) })
		if err != nil {
			t.Fatal(err)
		}
		return l
	}
	l := all()
	if len(l.Songs) != 2 || l.Songs[1].ComposerUUID != testenv.SeranaUUID || !l.Songs[1].HasAudio || l.Songs[1].Plays != 2 || l.Songs[0].HasAudio || l.Songs[0].ComposerUUID != "" {
		t.Errorf("songs = %+v", l.Songs)
	}

	title, lyrics, empty := "Ragnar the Red", "la la\r\nla", " "
	write := func(fn func(tx store.DB) error) error {
		_, err := store.Write(s, testenv.SaveID, "test", func(tx store.DB) (int, error) { return 0, fn(tx) })
		return err
	}
	if err := write(func(tx store.DB) error { return Update(tx, 1, Changes{Title: &title, Lyrics: &lyrics}) }); err != nil {
		t.Fatal(err)
	}
	if got := f.Query(`select title || '|' || lyrics from bard_songs where id = 1`); got != "Ragnar the Red|la la\nla" {
		t.Errorf("after editing: %q", got)
	}
	for _, c := range []Changes{{}, {Title: &empty}, {Lyrics: &empty}} {
		if kind, _ := apperr.KindOf(write(func(tx store.DB) error { return Update(tx, 1, c) })); kind != apperr.Invalid {
			t.Errorf("Update(%+v) wasn't refused", c)
		}
	}

	mux := http.NewServeMux()
	Routes(mux, s)
	r := httptest.NewRequest("GET", "/api/playthroughs/test/songs/1/audio", nil)
	r.Header.Set("Range", "bytes=2-4")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	if w.Code != 206 || w.Body.String() != "234" || w.Header().Get("Content-Type") != "audio/mpeg" {
		t.Errorf("audio range: %d %q", w.Code, w.Body)
	}
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest("GET", "/api/playthroughs/test/songs/2/audio", nil))
	if w.Code != 404 {
		t.Errorf("a name with .. was served: %d", w.Code)
	}

	if err := write(func(tx store.DB) error { return Delete(tx, 2) }); err != nil {
		t.Fatal(err)
	}
	if kind, _ := apperr.KindOf(write(func(tx store.DB) error { return Delete(tx, 2) })); kind != apperr.NotFound {
		t.Error("deleting twice wasn't refused")
	}
	if l := all(); len(l.Songs) != 1 {
		t.Errorf("after deleting: %+v", l.Songs)
	}
}

func TestAudioFileOnlyPlainNames(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "data", "bard_songs"), 0o755)
	os.WriteFile(filepath.Join(dir, "data", "bard_songs", "a.mp3"), nil, 0o644)
	os.WriteFile(filepath.Join(dir, "data", "x.mp3"), nil, 0o644)
	if AudioFile(dir, "a.mp3") == "" {
		t.Error("a.mp3 not found")
	}
	for _, bad := range []string{"../x.mp3", `..\x.mp3`, "sub/a.mp3", "a.txt", ""} {
		if f := AudioFile(dir, bad); f != "" {
			t.Errorf("AudioFile(%q) = %q, want refused", bad, f)
		}
	}
}
