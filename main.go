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
	err := ExtractURLs(r, func(url string) {
		fmt.Fprintf(w, "%s\n", url)
	})

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error processing input: %v\n", err)
		os.Exit(1)
	}
}
