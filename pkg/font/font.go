package font

import (
	"image/color"

	"tinygo.org/x/drivers"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/freemono"
	"tinygo.org/x/tinyfont/freesans"
	"tinygo.org/x/tinyfont/notoemoji"
	"tinygo.org/x/tinyfont/notosans"
	"tinygo.org/x/tinyfont/proggy"
)

type FontType int

const (
	FreemonoRegular18 FontType = iota
	FreemonoBold18
	FreemonoRegular12
	FreemonoBold12
	FreemonoRegular9
	FreemonoBold9
	FreesansRegular12
	FreesansBold12
	FreesansRegular9
	FreesansBold9
	ProggySZ8
	Notoemoji
	Notosans
)

type Face struct {
	display drivers.Displayer
	font    tinyfont.Fonter
	width   int16
	height  int16
}

func New(display drivers.Displayer, typ FontType) Face {
	face := Face{display: display, font: source(typ)}
	if typ == ProggySZ8 {
		face.width = 6
		face.height = 6
		return face
	}
	if glyph := face.font.GetGlyph('0'); glyph != nil {
		face.width = int16(glyph.Info().Width)
		face.height = int16(glyph.Info().Height)
	}
	return face
}

func (f Face) Print(x, y int16, text string) int16 {
	if f.display == nil || f.font == nil {
		return 0
	}
	tinyfont.WriteLine(f.display, f.font, x, y+f.height, text, color.RGBA{255, 255, 255, 255})
	return f.Width(text)
}

func (f Face) Width(text string) int16 {
	if f.font == nil {
		return 0
	}
	if f.width == 6 {
		return int16(len(text)) * f.width
	}
	_, width := tinyfont.LineWidth(f.font, text)
	return int16(width)
}

func (f Face) Height() int16 { return f.height }

func (f Face) Raw() tinyfont.Fonter { return f.font }

func source(typ FontType) tinyfont.Fonter {
	switch typ {
	case FreemonoRegular18:
		return &freemono.Regular18pt7b
	case FreemonoBold18:
		return &freemono.Bold18pt7b
	case FreemonoRegular12:
		return &freemono.Regular12pt7b
	case FreemonoBold12:
		return &freemono.Bold12pt7b
	case FreemonoRegular9:
		return &freemono.Regular9pt7b
	case FreemonoBold9:
		return &freemono.Bold9pt7b
	case FreesansRegular12:
		return &freesans.Regular12pt7b
	case FreesansBold12:
		return &freesans.Bold12pt7b
	case FreesansRegular9:
		return &freesans.Regular9pt7b
	case FreesansBold9:
		return &freesans.Bold9pt7b
	case Notoemoji:
		return &notoemoji.NotoEmojiRegular16pt
	case Notosans:
		return &notosans.Notosans12pt
	default:
		return &proggy.TinySZ8pt7b
	}
}
