// Package layout positions text and values on a pixel display.
//
//	left                 centered                 right
//	|text                   text                   text|
//	0                                                width
package layout

import (
	"strings"

	"pico_co2/pkg/font"
	"tinygo.org/x/drivers"
)

// Center draws text centered across the display and returns its left edge.
//
//	|--------------- text ---------------|
//	                ^
//	              result
func Center(display drivers.Displayer, face font.Face, y int16, text string) int16 {
	width, _ := display.Size()
	x := (width - face.Width(text)) / 2
	face.Print(x, y, text)
	return x
}

// Right draws text against the right display edge and returns its left edge.
//
//	|------------------------------- text|
//	                                ^
//	                              result
func Right(display drivers.Displayer, face font.Face, y int16, text string) int16 {
	width, _ := display.Size()
	x := width - face.Width(text)
	face.Print(x, y, text)
	return x
}

// Three draws three strings with equal gaps and returns their left edges.
//
//	|left          middle           right|
//	^             ^                ^
func Three(display drivers.Displayer, face font.Face, y int16, left, middle, right string) (int16, int16, int16) {
	positions := Row(display, y,
		Cell{Face: face, Text: left},
		Cell{Face: face, Text: middle},
		Cell{Face: face, Text: right},
	)
	return positions[0], positions[1], positions[2]
}

// Cell is one value in a Row. Unit is drawn five pixels below Text. YOffset
// preserves intentional baseline differences between mixed font sizes.
//
//	Text unit
//	^    ^
//	y+Y  y+Y+5
type Cell struct {
	Face     font.Face
	Text     string
	Unit     string
	UnitFace font.Face
	YOffset  int16
}

// Row distributes cells across the full display with equal gaps. It returns
// the left edge of every cell.
//
//	|24 c          55 %             1200|
//	^              ^                ^
//	positions[0]   positions[1]     positions[2]
func Row(display drivers.Displayer, y int16, cells ...Cell) []int16 {
	if len(cells) == 0 {
		return nil
	}
	width, _ := display.Size()
	positions := make([]int16, len(cells))
	used := int16(0)
	for _, cell := range cells {
		used += cellWidth(cell)
	}
	gap := int16(0)
	if len(cells) > 1 {
		gap = (width - used) / int16(len(cells)-1)
	}

	x := int16(0)
	for i, cell := range cells {
		positions[i] = x
		drawCell(cell, x, y)
		x += cellWidth(cell) + gap
	}
	return positions
}

// Wrap splits text at spaces and, when needed, splits words to fit width.
//
//	"sensor timeout while reading"
//	+--------------+
//	|sensor timeout|
//	|while reading |
//	+--------------+
func Wrap(text string, face font.Face, width int16) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}
	lines := make([]string, 0, len(words))
	for _, word := range words {
		parts := splitWord(word, face, width)
		for _, part := range parts {
			if len(lines) == 0 {
				lines = append(lines, part)
				continue
			}
			last := len(lines) - 1
			candidate := lines[last] + " " + part
			if face.Width(candidate) > width {
				lines = append(lines, part)
				continue
			}
			lines[last] = candidate
		}
	}
	return lines
}

func cellWidth(cell Cell) int16 {
	width := cell.Face.Width(cell.Text)
	if cell.Unit != "" {
		width += 1 + cell.UnitFace.Width(cell.Unit)
	}
	return width
}

func drawCell(cell Cell, x, y int16) {
	y += cell.YOffset
	cell.Face.Print(x, y, cell.Text)
	if cell.Unit != "" {
		cell.UnitFace.Print(x+cell.Face.Width(cell.Text)+1, y+5, cell.Unit)
	}
}

func splitWord(word string, face font.Face, width int16) []string {
	if width <= 0 || face.Width(word) <= width {
		return []string{word}
	}
	parts := make([]string, 0, len(word))
	for len(word) > 0 {
		end := 1
		for end < len(word) && face.Width(word[:end+1]) <= width {
			end++
		}
		parts = append(parts, word[:end])
		word = word[end:]
	}
	return parts
}

// LongText draws wrapped lines from (x,y) and clips complete lines below the
// display. Text wider than one line follows the same shape as Wrap.
//
//	(x,y) +----------------+
//	      |sensor timeout  |
//	      |while reading   |
//	      +----------------+
func LongText(display drivers.Displayer, face font.Face, x, y int16, text string) {
	width, height := display.Size()
	lineHeight := face.Height() + 1
	maxLines := int((height - y) / lineHeight)
	for i, line := range Wrap(text, face, width-x) {
		if i >= maxLines {
			return
		}
		face.Print(x, y+int16(i)*lineHeight, line)
	}
}
