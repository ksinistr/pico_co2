//go:build !tinygo

package display

import (
	"fmt"
	"image"
	"image/png"
	"os"

	"github.com/nfnt/resize"
)

func (v *VirtualDisplay) SavePNG(filename string) (err error) {
	rgbaImage := image.NewRGBA(image.Rect(0, 0, int(v.width), int(v.height)))
	for y := range v.height {
		for x := range v.width {
			rgbaImage.Set(int(x), int(y), v.buffer[y*v.width+x])
		}
	}

	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("create PNG %q: %w", filename, err)
	}

	defer func() {
		if closeErr := file.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("close PNG %q: %w", filename, closeErr)
		}
	}()

	resized := resize.Resize(
		uint(v.width*2),
		uint(v.height*2),
		rgbaImage,
		resize.NearestNeighbor,
	)
	if err := png.Encode(file, resized); err != nil {
		return fmt.Errorf("encode PNG %q: %w", filename, err)
	}

	return nil
}
