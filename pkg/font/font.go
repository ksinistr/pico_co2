// Package font selects TinyGo bitmap fonts and draws text using top-left
// coordinates.
//
//	(x,y) text
//	  +---------+
//	  |ABC      |
//	  +---------+
package font

import (
	"image/color"

	"tinygo.org/x/drivers"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/freemono"
	"tinygo.org/x/tinyfont/freesans"
	"tinygo.org/x/tinyfont/freeserif"
	"tinygo.org/x/tinyfont/notoemoji"
	"tinygo.org/x/tinyfont/notosans"
	"tinygo.org/x/tinyfont/proggy"
	"tinygo.org/x/tinyfont/shnm"
)

// FontType identifies one of the bitmap fonts compiled into the program.
type FontType int

const (
	FreemonoRegular18 FontType = iota
	FreemonoBold18
	FreemonoRegular12
	FreemonoBold12
	FreemonoRegular9
	FreemonoBold9
	FreesansRegular18
	FreesansBold18
	FreesansRegular12
	FreesansBold12
	FreesansRegular9
	FreesansBold9
	FreeserifRegular12
	FreeserifBold12
	ProggySZ8
	TomThumb
	Shnmk12
	Notoemoji
	Notosans
)

// Face couples a font with the display it draws onto.
type Face struct {
	display drivers.Displayer
	font    tinyfont.Fonter
	fixed   int16
	height  int16
}

// New returns a face bound to display.
func New(display drivers.Displayer, typ FontType) Face {
	face := Face{display: display, font: source(typ)}
	if typ == ProggySZ8 {
		face.fixed = 6
		face.height = 6

		return face
	}

	if glyph := face.font.GetGlyph('0'); glyph != nil {
		face.height = int16(glyph.Info().Height)
	}

	return face
}

// Print draws text with (x,y) as its top-left corner and returns its width.
//
//	(x,y)
//	  +----------+
//	  |text      |
//	  +----------+
//	  |< width ->|
func (f Face) Print(x, y int16, text string) int16 {
	if f.display == nil || f.font == nil {
		return 0
	}

	tinyfont.WriteLine(f.display, f.font, x, y+f.height, text, color.RGBA{255, 255, 255, 255})

	return f.Width(text)
}

// Width returns the horizontal pixel extent of text without drawing it.
// Proggy is an ASCII-only fixed-width face, so its width is measured in bytes.
//
//	|<---- Width("text") ---->|
//	text
func (f Face) Width(text string) int16 {
	if f.font == nil {
		return 0
	}

	if f.fixed > 0 {
		return int16(len(text)) * f.fixed
	}

	_, width := tinyfont.LineWidth(f.font, text)

	return int16(width)
}

// Height returns the face's pixel height.
//
//	+------+
//	| text | ^
//	+------+ | Height
func (f Face) Height() int16 { return f.height }

func source(typ FontType) tinyfont.Fonter {
	fonts := [...]tinyfont.Fonter{
		&freemono.Regular18pt7b,
		&freemono.Bold18pt7b,
		&freemono.Regular12pt7b,
		&freemono.Bold12pt7b,
		&freemono.Regular9pt7b,
		&freemono.Bold9pt7b,
		&freesans.Regular18pt7b,
		&freesans.Bold18pt7b,
		&freesans.Regular12pt7b,
		&freesans.Bold12pt7b,
		&freesans.Regular9pt7b,
		&freesans.Bold9pt7b,
		&freeserif.Regular12pt7b,
		&freeserif.Bold12pt7b,
		&proggy.TinySZ8pt7b,
		&tinyfont.TomThumb,
		&shnm.Shnmk12,
		&notoemoji.NotoEmojiRegular16pt,
		&notosans.Notosans12pt,
	}
	if typ >= 0 && typ < FontType(len(fonts)) {
		return fonts[typ]
	}

	return &proggy.TinySZ8pt7b
}
