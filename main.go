package main

import (
	"flag"
	"fmt"
	"io"
	"os"
)

func main() {
	enableGit := flag.Bool("git", false, "enable Git remote transformation to HTTPS URLs")
	flag.Parse()
	process(os.Stdin, os.Stdout, *enableGit)
}

func process(r io.Reader, w io.Writer, enableGitRemotes bool) {
	err := ExtractURLs(r, func(url string) {
		fmt.Fprintf(w, "%s\n", url)
	}, enableGitRemotes)

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error processing input: %v\n", err)
		os.Exit(1)
	}
}
