package prompts

import "strings"

// maxVisible caps how many options are shown at once; longer lists scroll a window
// around the cursor.
const maxVisible = 10

// runSelect shows an arrow-key menu and returns the chosen option.
func runSelect(label, def string, options []string) string {
	t, err := openTTY()
	if err != nil {
		bail(err)
	}
	defer t.close()

	cursor := max(indexOf(options, def), 0)
	rows, prev := viewport(t, len(options)), 0

	for {
		prev = drawMenu(t, label, options, cursor, rows, prev, nil)

		switch k, _ := t.key(); k {
		case keyUp:
			cursor = wrap(cursor-1, len(options))
		case keyDown:
			cursor = wrap(cursor+1, len(options))
		case keyEnter:
			clearMenu(t, prev)
			t.write(ask(label) + " " + cyan(options[cursor]) + "\r\n")
			return options[cursor]
		case keyInterrupt, keyEsc:
			clearMenu(t, prev)
			t.close()
			cancel()
		}
	}
}

// runMultiSelect shows a checkbox menu (space toggles, enter confirms), pre-checking
// any options listed in defaults.
func runMultiSelect(label string, options, defaults []string) []string {
	t, err := openTTY()
	if err != nil {
		bail(err)
	}
	defer t.close()

	chosen := make([]bool, len(options))
	for i, o := range options {
		chosen[i] = indexOf(defaults, o) >= 0
	}
	cursor, prev := 0, 0
	rows := viewport(t, len(options))

	for {
		prev = drawMenu(t, label, options, cursor, rows, prev, chosen)

		switch k, _ := t.key(); k {
		case keyUp:
			cursor = wrap(cursor-1, len(options))
		case keyDown:
			cursor = wrap(cursor+1, len(options))
		case keySpace:
			chosen[cursor] = !chosen[cursor]
		case keyEnter:
			clearMenu(t, prev)
			result := picked(options, chosen)
			t.write(ask(label) + " " + cyan(strings.Join(result, ", ")) + "\r\n")
			return result
		case keyInterrupt, keyEsc:
			clearMenu(t, prev)
			t.close()
			cancel()
		}
	}
}

// drawMenu redraws the menu in place and returns the number of lines written (so the
// next redraw knows how far up to clear). A nil chosen renders a single-select menu;
// non-nil renders checkboxes.
func drawMenu(t *tty, label string, options []string, cursor, rows, prev int, chosen []bool) int {
	if prev > 0 {
		t.write("\r" + cursorUp(prev) + eraseDown)
	}

	hint := "↑/↓, enter to select"
	if chosen != nil {
		hint = "↑/↓, space to toggle, enter to confirm"
	}
	t.write(ask(label) + " " + dim(hint) + "\r\n")
	lines := 1

	start, end := window(cursor, len(options), rows)
	for i := start; i < end; i++ {
		marker := "  "
		if i == cursor {
			marker = cyan("> ")
		}

		text := options[i]
		if chosen != nil {
			if chosen[i] {
				text = "[x] " + text
			} else {
				text = "[ ] " + text
			}
		}
		if i == cursor {
			text = cyan(text)
		}

		t.write(marker + text + "\r\n")
		lines++
	}
	return lines
}

func clearMenu(t *tty, prev int) {
	if prev > 0 {
		t.write("\r" + cursorUp(prev) + eraseDown)
	}
}

// viewport is how many option rows fit: capped by maxVisible, the terminal height (less
// the header line), and the option count, floored at 3.
func viewport(t *tty, total int) int {
	rows := min(t.rows()-1, maxVisible)
	rows = max(rows, 3)
	rows = min(rows, total)
	return rows
}

// window returns the [start,end) slice of options visible for the given cursor, keeping
// the cursor roughly centered when the list is longer than rows.
func window(cursor, total, rows int) (start, end int) {
	if total <= rows {
		return 0, total
	}

	start = max(cursor-rows/2, 0)
	end = start + rows
	if end > total {
		end = total
		start = end - rows
	}
	return start, end
}

func wrap(i, n int) int {
	if n == 0 {
		return 0
	}
	return (i%n + n) % n
}

func indexOf(options []string, v string) int {
	for i, o := range options {
		if o == v {
			return i
		}
	}
	return -1
}

func picked(options []string, chosen []bool) []string {
	var out []string
	for i, ok := range chosen {
		if ok {
			out = append(out, options[i])
		}
	}
	return out
}
