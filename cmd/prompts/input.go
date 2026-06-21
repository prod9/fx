package prompts

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// readLine reads a single trimmed line in cooked mode, so the terminal's own line
// editing (backspace, kill-word, etc.) works without us reimplementing it.
func readLine(label string) string {
	fmt.Fprint(os.Stderr, ask(label)+": ")

	line, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil && line == "" {
		bail(err)
	}
	return strings.TrimSpace(line)
}

// readConfirm reads a y/n answer in cooked mode. Empty input defaults to no.
func readConfirm(what string) bool {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Fprintf(os.Stderr, "%s %s ", ask(what), dim("[y/N]"))

		line, err := reader.ReadString('\n')
		if err != nil && line == "" {
			bail(err)
		}
		switch strings.ToLower(strings.TrimSpace(line)) {
		case "y", "yes":
			return true
		case "n", "no", "":
			return false
		}
	}
}

// readSecret reads a masked secret in raw mode, echoing one '*' per rune.
func readSecret(label string) string {
	t, err := openTTY()
	if err != nil {
		bail(err)
	}
	defer t.close()

	t.write(ask(label) + ": ")
	var buf []rune
	for {
		switch k, r := t.key(); k {
		case keyEnter:
			t.write("\r\n")
			return string(buf)
		case keyInterrupt, keyEsc:
			t.write("\r\n")
			t.close()
			cancel()
		case keyBackspace:
			if len(buf) > 0 {
				buf = buf[:len(buf)-1]
				t.write("\b \b")
			}
		case keyRune:
			buf = append(buf, r)
			t.write("*")
		}
	}
}
