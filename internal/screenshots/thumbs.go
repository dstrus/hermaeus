package screenshots

import (
	"bytes"
	"image"
	"image/jpeg"
	_ "image/png" // screenshots are PNG
	"os"
	"sync"
	"time"
)

// ThumbSize is the longer side of a thumbnail, in pixels: enough for a grid
// tile on a high-density screen.
const ThumbSize = 320

// Thumbs makes small JPEG copies of screenshots for a grid of them, where
// the full images (a few megabytes each) would be slow to load. Each is
// made once per image file and kept in memory, a few dozen kilobytes apiece.
type Thumbs struct {
	mu   sync.Mutex
	made map[string]thumb
}

type thumb struct {
	modified time.Time
	jpeg     []byte
}

// most bounds the cache; past it, the cache starts over.
const most = 2000

// Of is the thumbnail of the image at path, made or kept. ok is false when
// the image can't be read as one Go decodes (BMP and WebP aren't), and the
// full image should be sent instead.
func (t *Thumbs) Of(path string) (data []byte, modified time.Time, ok bool) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, time.Time{}, false
	}
	t.mu.Lock()
	if th, found := t.made[path]; found && th.modified.Equal(info.ModTime()) {
		t.mu.Unlock()
		return th.jpeg, th.modified, true
	}
	t.mu.Unlock()

	f, err := os.Open(path)
	if err != nil {
		return nil, time.Time{}, false
	}
	img, _, err := image.Decode(f)
	f.Close()
	if err != nil {
		return nil, time.Time{}, false
	}
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, shrink(img, ThumbSize), &jpeg.Options{Quality: 82}); err != nil {
		return nil, time.Time{}, false
	}
	t.mu.Lock()
	if t.made == nil || len(t.made) >= most {
		t.made = map[string]thumb{}
	}
	t.made[path] = thumb{info.ModTime(), buf.Bytes()}
	t.mu.Unlock()
	return buf.Bytes(), info.ModTime(), true
}

// shrink scales img down so its longer side is at most size, averaging the
// pixels each new one covers. A smaller image is returned as it is.
func shrink(img image.Image, size int) image.Image {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w <= size && h <= size {
		return img
	}
	nw, nh := size, h*size/w
	if h > w {
		nw, nh = w*size/h, size
	}
	nw, nh = max(nw, 1), max(nh, 1)
	out := image.NewRGBA(image.Rect(0, 0, nw, nh))
	for y := 0; y < nh; y++ {
		y0, y1 := b.Min.Y+y*h/nh, b.Min.Y+(y+1)*h/nh
		for x := 0; x < nw; x++ {
			x0, x1 := b.Min.X+x*w/nw, b.Min.X+(x+1)*w/nw
			var r, g, bl, a, n uint64
			for sy := y0; sy < max(y1, y0+1); sy++ {
				for sx := x0; sx < max(x1, x0+1); sx++ {
					cr, cg, cb, ca := img.At(sx, sy).RGBA()
					r, g, bl, a, n = r+uint64(cr), g+uint64(cg), bl+uint64(cb), a+uint64(ca), n+1
				}
			}
			i := out.PixOffset(x, y)
			out.Pix[i], out.Pix[i+1], out.Pix[i+2], out.Pix[i+3] = uint8(r/n>>8), uint8(g/n>>8), uint8(bl/n>>8), uint8(a/n>>8)
		}
	}
	return out
}
