package main

import (
	"bufio"
	"io"
	"net/url"
	"regexp"
	"strings"
)

var (
	urlCharASCII   = `[a-zA-Z0-9._~:/?#[\]@!$&'()*+,;=%-]` // Standard ASCII URL characters
	urlCharUnicode = `\p{L}|\p{N}`                         // Unicode letters and numbers
	urlChar        = urlCharASCII + `|` + urlCharUnicode   // Combined character class

	schemeStart = `[a-zA-Z]`                                     // Scheme must start with a letter
	schemeRest  = `[a-zA-Z0-9+.-]*`                              // Scheme can contain letters, digits, +, -, .
	schemeSep   = `://`                                          // Standard URL scheme separator
	urlRest     = `(?:` + urlChar + `)+`                         // Authority + path (specific URL characters, not just non-whitespace)
	standardURL = schemeStart + schemeRest + schemeSep + urlRest // Standard URL: scheme://authority+path

	gitUser    = `[a-zA-Z0-9_.-]+`                                     // Git username (letters, digits, underscore, dot, dash)
	gitUserSep = `@`                                                   // User separator
	gitHost    = `[a-zA-Z0-9_.-]+`                                     // Git hostname (letters, digits, underscore, dot, dash)
	gitPathSep = `:`                                                   // Path separator (not //, distinguishes from URL)
	gitPath    = `[a-zA-Z0-9_./-]+(?:\.git)?`                          // Repository path (specific characters common in Git paths)
	scpGitURL  = gitUser + gitUserSep + gitHost + gitPathSep + gitPath // SCP Git: user@host:path

	combinedPattern = `(?:` + standardURL + `|` + scpGitURL + `)`
	combinedRegex   = regexp.MustCompile(combinedPattern)
)

func findURLs(text string, callback func(string)) {
	indices := combinedRegex.FindAllStringIndex(text, -1)

	for _, match := range indices {
		start, end := match[0], match[1]
		rawMatch := text[start:end]

		trimmed := TrimWrapperPunctuation(rawMatch)

		var final string
		if IsGitRemote(trimmed) {
			if transformed, ok := TransformGitRemote(trimmed); ok {
				final = transformed
			} else {
				continue
			}
		} else {
			normalized := NormalizeURL(trimmed)
			u, err := url.Parse(normalized)
			if err != nil {
				continue
			}
			if !IsValidRFC3986URL(u) {
				continue
			}
			final = normalized
		}

		finalURL, err := url.Parse(final)
		if err != nil {
			continue
		}
		scheme := strings.ToLower(finalURL.Scheme)
		if scheme != "http" && scheme != "https" {
			continue
		}

		callback(final)
	}
}

func ExtractURLs(reader io.Reader, callback func(string)) error {
	scanner := bufio.NewScanner(reader)

	// Increase buffer size to handle large lines (default is ~64KB)
	const maxTokenSize = 10 * 1024 * 1024 // 10MB
	buf := make([]byte, maxTokenSize)
	scanner.Buffer(buf, maxTokenSize)

	seen := make(map[string]bool)

	for scanner.Scan() {
		line := scanner.Text()

		findURLs(line, func(url string) {
			if !seen[url] {
				seen[url] = true
				callback(url)
			}
		})
	}

	return scanner.Err()
}
