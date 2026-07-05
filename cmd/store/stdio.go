package store

import (
	"bytes"
	"io"
	"os"
)

// openSource returns a reader for an upload body. "-" (or empty) reads stdin,
// buffered so its length is known; a path streams the file with its stat size.
func openSource(src string) (io.Reader, int64, func() error, error) {
	if src == "" || src == "-" {
		buf, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, 0, nil, err
		}
		return bytes.NewReader(buf), int64(len(buf)), func() error { return nil }, nil
	}

	file, err := os.Open(src)
	if err != nil {
		return nil, 0, nil, err
	}

	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, 0, nil, err
	}
	return file, info.Size(), file.Close, nil
}

// openDest returns a writer for a download body. "-" (or empty) writes stdout;
// a path creates the file.
func openDest(dest string) (io.Writer, func() error, error) {
	if dest == "" || dest == "-" {
		return os.Stdout, func() error { return nil }, nil
	}

	file, err := os.Create(dest)
	if err != nil {
		return nil, nil, err
	}
	return file, file.Close, nil
}

func arg(args []string, i int) string {
	if i < len(args) {
		return args[i]
	}
	return ""
}
