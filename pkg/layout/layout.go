package layout

import (
	"strings"

	"pico_co2/pkg/font"
	"tinygo.org/x/drivers"
)

func Left(face font.Face, x, y int16, text string) int16 { return face.Print(x, y, text) }

func Center(display drivers.Displayer, face font.Face, y int16, text string) int16 {
	width, _ := display.Size()
	x := (width - face.Width(text)) / 2
	face.Print(x, y, text)
	return x
}

func Right(display drivers.Displayer, face font.Face, y int16, text string) int16 {
	width, _ := display.Size()
	x := width - face.Width(text)
	face.Print(x, y, text)
	return x
}

func Three(display drivers.Displayer, face font.Face, y int16, left, middle, right string) (int16, int16, int16) {
	width, _ := display.Size()
	leftX := int16(0)
	rightX := width - face.Width(right)
	middleX := face.Width(left) + (width-face.Width(left)-face.Width(middle)-face.Width(right))/2
	face.Print(leftX, y, left)
	face.Print(middleX, y, middle)
	face.Print(rightX, y, right)
	return leftX, middleX, rightX
}

func NumberWithUnit(number, unit font.Face, x, y int16, value, suffix string) int16 {
	number.Print(x, y, value)
	unitX := x + number.Width(value) + 1
	unit.Print(unitX, y+5, suffix)
	return unitX + unit.Width(suffix)
}

func Wrap(text string, face font.Face, width int16) []string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return nil
	}
	lines := []string{words[0]}
	for _, word := range words[1:] {
		last := len(lines) - 1
		candidate := lines[last] + " " + word
		if face.Width(candidate) > width {
			lines = append(lines, word)
			continue
		}
		lines[last] = candidate
	}
	return lines
}

func LongText(display drivers.Displayer, face font.Face, x, y int16, text string) {
	width, height := display.Size()
	lineHeight := face.Height() + 1
	maxLines := int((height - y) / lineHeight)
	for i, line := range Wrap(text, face, width-x) {
		if i >= maxLines {
			return
		}
		face.Print(x, y+face.Height()+int16(i)*lineHeight, line)
	}
}
