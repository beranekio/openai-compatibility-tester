package suites

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"io"
)

// smallPNG is an 8x8 RGBA PNG with transparent pixels for DALL-E 2 edit compatibility.
var smallPNG = func() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			if x < 4 {
				img.SetRGBA(x, y, color.RGBA{R: 255, A: 0})
			} else {
				img.SetRGBA(x, y, color.RGBA{R: 255, A: 255})
			}
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		panic(err)
	}
	return buf.Bytes()
}()

type namedPNGReader struct {
	r *bytes.Reader
}

func (r *namedPNGReader) Read(p []byte) (int, error) {
	return r.r.Read(p)
}

func (r *namedPNGReader) Filename() string {
	return "test.png"
}

func (r *namedPNGReader) ContentType() string {
	return "image/png"
}

func smallPNGReader() io.Reader {
	return &namedPNGReader{r: bytes.NewReader(smallPNG)}
}