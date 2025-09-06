package main

import (
	"net/url"
	"regexp"
	"strings"
)

var (
	// More inclusive URL character sets that support international domains
	urlChar = `[a-zA-Z0-9._~:/?#[\]@!$&'()*+,;=%-]|\p{L}|\p{N}` // Valid URL chars + Unicode letters/numbers

	// URL pattern components
	schemeStart = `[a-zA-Z]`             // Scheme must start with a letter
	schemeRest  = `[a-zA-Z0-9+.-]*`      // Scheme can contain letters, digits, +, -, .
	schemeSep   = `://`                  // Standard URL scheme separator
	urlRest     = `(?:` + urlChar + `)+` // Authority + path (specific URL characters, not just non-whitespace)

	// SCP-style Git remote pattern components
	gitUser    = `[a-zA-Z0-9_.-]+`            // Git username (letters, digits, underscore, dot, dash)
	gitUserSep = `@`                          // User separator
	gitHost    = `[a-zA-Z0-9_.-]+`            // Git hostname (letters, digits, underscore, dot, dash)
	gitPathSep = `:`                          // Path separator (not //, distinguishes from URL)
	gitPath    = `[a-zA-Z0-9_./-]+(?:\.git)?` // Repository path (specific characters common in Git paths)

	standardURL = schemeStart + schemeRest + schemeSep + urlRest        // Standard URL: scheme://authority+path
	scpGitURL   = gitUser + gitUserSep + gitHost + gitPathSep + gitPath // SCP Git: user@host:path

	combinedRegex = regexp.MustCompile(`(?:` + standardURL + `|` + scpGitURL + `)`)
)

func ExtractURLs(text string) []string {
	var results []string
	seen := make(map[string]bool)

	matches := combinedRegex.FindAllString(text, -1)
	for _, match := range matches {
		trimmed := TrimWrapperPunctuation(match)

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

		if !seen[final] {
			results = append(results, final)
			seen[final] = true
		}
	}

	return results
}
