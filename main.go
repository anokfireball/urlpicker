package main

import (
	"fmt"
	"io"
	"os"
)

func main() {
	process(os.Stdin, os.Stdout)
}

func process(r io.Reader, w io.Writer) {
	input, err := io.ReadAll(r)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input: %v\n", err)
		os.Exit(1)
	}

	urls := ExtractURLs(string(input))

	for _, url := range urls {
		fmt.Fprintf(w, "%s\n", url)
	}
}
