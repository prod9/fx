package prompts

import (
	"os"
	"strconv"
)

// Minimal ANSI control sequences. Rendering goes to stderr so stdout stays clean for
// piping. We honor NO_COLOR but otherwise assume a VT-capable terminal.

const csi = "\x1b["

const (
	eraseDown  = csi + "J" // erase from cursor to end of screen
	hideCursor = csi + "?25l"
	showCursor = csi + "?25h"
)

var useColor = os.Getenv("NO_COLOR") == ""

func cursorUp(n int) string {
	if n <= 0 {
		return ""
	}
	return csi + strconv.Itoa(n) + "A"
}

func paint(code, s string) string {
	if !useColor {
		return s
	}
	return csi + code + "m" + s + csi + "0m"
}

func cyan(s string) string { return paint("36", s) }
func dim(s string) string  { return paint("2", s) }

// ask renders the leading "? <label>" question prefix.
func ask(label string) string { return cyan("?") + " " + label }
