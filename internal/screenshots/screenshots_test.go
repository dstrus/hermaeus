package screenshots

import (
	"bytes"
	"image"
	"image/jpeg"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"hermaeus/internal/apperr"
	"hermaeus/internal/store"
	"hermaeus/internal/testenv"
)

func setup(t *testing.T) (*testenv.Folder, *store.Store) {
	t.Helper()
	f := testenv.NewFolder(t)
	s := store.New("", func() bool { return false })
	if _, err := s.SetDir(f.Dir); err != nil {
		t.Fatal(err)
	}
	return f, s
}

func TestListEditAndServe(t *testing.T) {
	f, s := setup(t)
	f.Exec(`insert into omnisight_screenshots (id, subject_type, subject_identifier, subject_name, location_name, capture_time_local, capture_timestamp_local, capture_time_game, file_path, description, description_locked)
values (1, 'location', 'Helgen', 'Helgen', 'Helgen', '2026-10-11T02:21:11Z', 10, '7:54 AM, Sundas', '["C:\\Steam\\Data\\SKSE\\Plugins\\SkyrimNet\\screenshots/omnisight\\location_Helgen.png"]', 'Ruins.', 1),
       (2, 'location', 'Riverwood01', '', 'Riverwood', '2026-10-12T02:21:11Z', 20, '', 'E:\gone\location_Riverwood01.png', '', 0)`)
	shots := filepath.Join(f.Dir, "screenshots", "omnisight")
	os.MkdirAll(shots, 0o755)
	os.WriteFile(filepath.Join(shots, "location_Helgen.png"), []byte("\x89PNG"), 0o644)

	l, err := store.Read(s, testenv.SaveID, func(db store.DB) (List, error) { return All(db, s.Dir()) })
	if err != nil {
		t.Fatal(err)
	}
	if len(l.Screenshots) != 2 || l.Screenshots[0].ID != 2 || l.Screenshots[0].Name != "Riverwood01" || l.Screenshots[0].HasImage {
		t.Errorf("newest = %+v", l.Screenshots[0])
	}
	if h := l.Screenshots[1]; !h.HasImage || !h.Locked || h.Description != "Ruins." || h.GameTime != "7:54 AM, Sundas" {
		t.Errorf("Helgen = %+v", h)
	}

	name, text, lock := "Riverwood mill", "  A mill.  ", true
	if _, err := store.Write(s, testenv.SaveID, "test", func(tx store.DB) (int, error) {
		return 0, Update(tx, 2, Changes{Name: &name, Description: &text, Locked: &lock})
	}); err != nil {
		t.Fatal(err)
	}
	if got := f.Query(`select description || '|' || subject_name || '|' || description_locked from omnisight_screenshots where id = 2`); got != "A mill.|Riverwood mill|1" {
		t.Errorf("after editing: %v", got)
	}
	for id, c := range map[int64]Changes{2: {}, 42: {Name: &name}} {
		_, err := store.Write(s, testenv.SaveID, "test", func(tx store.DB) (int, error) { return 0, Update(tx, id, c) })
		if kind, _ := apperr.KindOf(err); kind != apperr.Invalid && kind != apperr.NotFound {
			t.Errorf("Update(%d, %+v) = %v", id, c, err)
		}
	}

	mux := http.NewServeMux()
	Routes(mux, s)
	for path, want := range map[string]int{"/api/playthroughs/test/screenshots/1/image": 200, "/api/playthroughs/test/screenshots/2/image": 404, "/api/playthroughs/test/screenshots/42/image": 404} {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != want || (want == 200 && w.Header().Get("Content-Type") != "image/png") {
			t.Errorf("%s: %d %s", path, w.Code, w.Header().Get("Content-Type"))
		}
	}
}

func TestOlderDatabaseWithoutLaterColumns(t *testing.T) {
	f, s := setup(t)
	f.Exec(`alter table omnisight_screenshots drop column suggested_name;
alter table omnisight_screenshots drop column description_locked;
alter table omnisight_screenshots drop column manual_capture;
insert into omnisight_screenshots (subject_type, subject_identifier, capture_time_local, capture_timestamp_local) values ('location', 'Helgen', '', 0)`)
	l, err := store.Read(s, testenv.SaveID, func(db store.DB) (List, error) { return All(db, s.Dir()) })
	if err != nil || len(l.Screenshots) != 1 || l.Screenshots[0].Name != "Helgen" {
		t.Fatalf("list = %+v, %v", l, err)
	}
	lock := true
	_, err = store.Write(s, testenv.SaveID, "test", func(tx store.DB) (int, error) { return 0, Update(tx, 1, Changes{Locked: &lock}) })
	if kind, _ := apperr.KindOf(err); kind != apperr.Invalid {
		t.Errorf("locking without the column = %v", err)
	}
}

func TestImageFile(t *testing.T) {
	dir := t.TempDir()
	want := filepath.Join(dir, "screenshots", "omnisight", "location_A.png")
	os.MkdirAll(filepath.Dir(want), 0o755)
	os.WriteFile(want, nil, 0o644)

	for _, stored := range []string{
		`C:\Steam\steamapps\common\SkyrimVR\Data\SKSE\Plugins\SkyrimNet\screenshots/omnisight\location_A.png`,
		`D:/MO2/overwrite/skse/plugins/skyrimnet/screenshots/omnisight/location_A.png`,
		`E:\elsewhere\location_A.png`, // found by name in screenshots/omnisight
		`["C:\\x\\SKSE\\Plugins\\SkyrimNet\\screenshots\\omnisight\\location_A.png"]`,
	} {
		if got := ImageFile(dir, stored); got != want {
			t.Errorf("ImageFile(%q) = %q, want %q", stored, got, want)
		}
	}
	secret := filepath.Join(filepath.Dir(dir), "secret.png")
	os.WriteFile(secret, nil, 0o644)
	defer os.Remove(secret)
	if got := ImageFile(dir, `X:\SKSE\Plugins\SkyrimNet\..\secret.png`); got != "" {
		t.Errorf("followed .. out of the SkyrimNet folder: %q", got)
	}
	os.WriteFile(filepath.Join(dir, "screenshots", "omnisight", "notes.txt"), nil, 0o644)
	if got := ImageFile(dir, `C:\SKSE\Plugins\SkyrimNet\screenshots\omnisight\notes.txt`); got != "" {
		t.Errorf("served a file that isn't an image: %q", got)
	}
}

func TestThumbnails(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.png")
	img := image.NewRGBA(image.Rect(0, 0, 1024, 512))
	for i := range img.Pix {
		img.Pix[i] = 200
	}
	var buf bytes.Buffer
	png.Encode(&buf, img)
	os.WriteFile(path, buf.Bytes(), 0o644)

	var th Thumbs
	data, _, ok := th.Of(path)
	if !ok {
		t.Fatal("no thumbnail")
	}
	got, err := jpeg.Decode(bytes.NewReader(data))
	if err != nil || got.Bounds().Dx() != ThumbSize || got.Bounds().Dy() != ThumbSize/2 {
		t.Fatalf("thumbnail = %v, %v", got.Bounds(), err)
	}
	if r, _, _, _ := got.At(10, 10).RGBA(); r>>8 < 190 || r>>8 > 210 {
		t.Errorf("color = %d, want about 200", r>>8)
	}
	if again, _, _ := th.Of(path); &again[0] != &data[0] {
		t.Error("made again rather than kept")
	}
	os.WriteFile(filepath.Join(dir, "b.png"), []byte("not a png"), 0o644)
	if _, _, ok := th.Of(filepath.Join(dir, "b.png")); ok {
		t.Error("a broken image made a thumbnail")
	}
}
