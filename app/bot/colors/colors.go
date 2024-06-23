package colors

import "fmt"

// Color is a color.
type Color uint

const (
	// Red is the color red.
	Red Color = 0xf44336
	// Blue is the color blue.
	Blue Color = 0x2196f3
	// Green is the color green.
	Green Color = 0x4caf50
	// Yellow is the color yellow.
	Yellow Color = 0xffeb3b
	// Purple is the color purple.
	Purple Color = 0x9c27b0
	// Pink is the color pink.
	Pink Color = 0xe91e63
	// Orange is the color orange.
	Orange Color = 0xff9800
	// White is the color white.
	White Color = 0xffffff
	// Black is the color black.
	Black Color = 0x000000
)

// colors is a map of colors to their names.
var colors = map[Color]string{
	Red:    "Red",
	Blue:   "Blue",
	Green:  "Green",
	Yellow: "Yellow",
	Purple: "Purple",
	Pink:   "Pink",
	Orange: "Orange",
	White:  "White",
	Black:  "Black",
}

// String returns the color as a (hex) string.
func (c Color) String() string {
	return fmt.Sprintf("#%06x", int(c))
}

// Int returns the color as an integer.
func (c Color) Int() int {
	return int(c)
}

// Uint returns the color as an unsigned integer.
func (c Color) Uint() uint {
	return uint(c)
}

// Name returns the name of the color.
func (c Color) Name() string {
	if name, ok := colors[c]; ok {
		return name
	}
	return "Unknown"
}
