package prompts

import (
	"bufio"
	"io"
	"os"
	"unicode/utf8"

	"fx.prodigy9.co/errutil"
	"golang.org/x/term"
)

// tty owns the terminal while a raw-mode prompt is running. close is idempotent so it is
// safe both as a deferred cleanup and as an explicit restore before bail/cancel (which
// os.Exit and would otherwise skip the defer, leaving the terminal in raw mode).
type tty struct {
	fd     int
	state  *term.State
	in     *bufio.Reader
	out    io.Writer
	closed bool
}

func openTTY() (t *tty, err error) {
	defer errutil.Wrap("prompts", &err)

	fd := int(os.Stdin.Fd())
	state, err := term.MakeRaw(fd)
	if err != nil {
		return nil, err
	}

	out := os.Stderr
	_, _ = io.WriteString(out, hideCursor)
	return &tty{fd, state, bufio.NewReader(os.Stdin), out, false}, nil
}

func (t *tty) close() {
	if t.closed {
		return
	}
	t.closed = true
	_, _ = io.WriteString(t.out, showCursor)
	_ = term.Restore(t.fd, t.state)
}

func (t *tty) write(s string) { _, _ = io.WriteString(t.out, s) }

// rows reports the terminal height, defaulting to 24 when it cannot be determined.
func (t *tty) rows() int {
	if _, h, err := term.GetSize(t.fd); err == nil && h > 0 {
		return h
	}
	return 24
}

// key reads the next keypress, restoring the terminal and bailing on read error.
func (t *tty) key() (key, rune) {
	k, r, err := t.readKey()
	if err != nil {
		t.close()
		bail(err)
	}
	return k, r
}

func (t *tty) readKey() (key, rune, error) {
	b, err := t.in.ReadByte()
	if err != nil {
		return keyNone, 0, err
	}

	switch b {
	case '\r', '\n':
		return keyEnter, 0, nil
	case 0x03:
		return keyInterrupt, 0, nil
	case 0x7f, 0x08:
		return keyBackspace, 0, nil
	case ' ':
		return keySpace, ' ', nil
	case 0x1b:
		return t.readEscape()
	}
	if b < 0x20 {
		return keyNone, 0, nil
	}

	r, err := t.decodeRune(b)
	return keyRune, r, err
}

// readEscape decodes a CSI sequence already begun by an ESC byte. A lone ESC (nothing
// buffered behind it) is reported as keyEsc.
func (t *tty) readEscape() (key, rune, error) {
	if t.in.Buffered() == 0 {
		return keyEsc, 0, nil
	}
	if b, err := t.in.ReadByte(); err != nil {
		return keyNone, 0, err
	} else if b != '[' && b != 'O' {
		return keyEsc, 0, nil
	}

	switch b, err := t.in.ReadByte(); {
	case err != nil:
		return keyNone, 0, err
	case b == 'A':
		return keyUp, 0, nil
	case b == 'B':
		return keyDown, 0, nil
	case b == 'C':
		return keyRight, 0, nil
	case b == 'D':
		return keyLeft, 0, nil
	case b == '3':
		_, _ = t.in.ReadByte() // consume the trailing '~' of ESC [ 3 ~
		return keyBackspace, 0, nil
	}
	return keyNone, 0, nil
}

func (t *tty) decodeRune(first byte) (rune, error) {
	n := 1
	switch {
	case first&0xE0 == 0xC0:
		n = 2
	case first&0xF0 == 0xE0:
		n = 3
	case first&0xF8 == 0xF0:
		n = 4
	}
	if n == 1 {
		return rune(first), nil
	}

	buf := make([]byte, n)
	buf[0] = first
	for i := 1; i < n; i++ {
		c, err := t.in.ReadByte()
		if err != nil {
			return 0, err
		}
		buf[i] = c
	}

	r, _ := utf8.DecodeRune(buf)
	return r, nil
}
