package prompts

// key is a decoded keypress. Runes carry their value alongside keyRune.
type key int

const (
	keyNone key = iota
	keyEnter
	keyUp
	keyDown
	keyLeft
	keyRight
	keyBackspace
	keySpace
	keyEsc
	keyInterrupt // ctrl-c
	keyRune
)
