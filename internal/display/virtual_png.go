//go:build !tinygo

package display

import (
	"image"
	"image/png"
	"os"

	"github.com/nfnt/resize"
)

func (v *VirtualDisplay) SavePNG(filename string) error {
	image := image.NewRGBA(image.Rect(0, 0, int(v.width), int(v.height)))
	for y := int16(0); y < v.height; y++ {
		for x := int16(0); x < v.width; x++ {
			image.Set(int(x), int(y), v.buffer[y*v.width+x])
		}
	}
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	return png.Encode(file, resize.Resize(uint(v.width*2), uint(v.height*2), image, resize.NearestNeighbor))
}
