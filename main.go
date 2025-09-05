package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// based on RFC 3986
var urlRegex = regexp.MustCompile(
	`https?://` + // scheme
		`[` +
		`A-Za-z0-9` + // unreserved: ALPHA DIGIT
		`\-\._~` + // unreserved: - . _ ~
		`!\$&'\(\)\*\+,;=` + // sub-delims (including quotes to detect them)
		`:/\?\#\[\]@` + // pchar extras: : @ and path/query/fragment: / ? # [ ]
		`%` + // for percent-encoded sequences
		`"` + // include double quote to detect it
		`]+`,
)

// Git SSH URL pattern: user@host:path.git
var gitSSHRegex = regexp.MustCompile(
	`[a-zA-Z0-9._-]+@[a-zA-Z0-9.-]+:[a-zA-Z0-9._/-]+\.git`,
)

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	seen := make(map[string]struct{})

	for scanner.Scan() {
		line := scanner.Text()
		urls := extractURLs(line)
		for _, url := range urls {
			if _, ok := seen[url]; ok {
				continue
			}
			seen[url] = struct{}{}
			fmt.Println(url)
		}
	}
}

func convertGitURLToWeb(input string) (string, bool) {
	// Check if it's a Git SSH URL
	if gitSSHRegex.MatchString(input) {
		// Parse SSH format: user@host:path.git
		parts := strings.Split(input, "@")
		if len(parts) != 2 {
			return "", false
		}

		hostAndPath := parts[1]
		colonIndex := strings.Index(hostAndPath, ":")
		if colonIndex == -1 {
			return "", false
		}

		host := hostAndPath[:colonIndex]
		path := hostAndPath[colonIndex+1:]

		// Remove .git suffix
		if strings.HasSuffix(path, ".git") {
			path = strings.TrimSuffix(path, ".git")
		}

		return fmt.Sprintf("https://%s/%s", host, path), true
	}

	// Check if it's an HTTPS URL with .git suffix
	if strings.HasPrefix(input, "http") && strings.HasSuffix(input, ".git") {
		return strings.TrimSuffix(input, ".git"), true
	}

	return "", false
}

func extractURLs(input string) []string {
	// Extract regular HTTP/HTTPS URLs
	httpMatches := urlRegex.FindAllString(input, -1)

	// Extract Git SSH URLs
	sshMatches := gitSSHRegex.FindAllString(input, -1)

	// Combine all matches
	allMatches := append(httpMatches, sshMatches...)

	if len(allMatches) == 0 {
		return nil
	}

	seen := make(map[string]struct{})
	urls := make([]string, 0, len(allMatches))

	for _, raw := range allMatches {
		var finalURL string

		// Try to convert as Git URL first
		if webURL, isGit := convertGitURLToWeb(raw); isGit {
			finalURL = webURL
		} else {
			// Handle as regular HTTP URL
			finalURL = sanitizeURL(raw)
		}

		if finalURL != "" {
			if _, exists := seen[finalURL]; !exists {
				seen[finalURL] = struct{}{}
				urls = append(urls, finalURL)
			}
		}
	}
	return urls
}

func sanitizeURL(s string) string {
	if s == "" {
		return ""
	}

	s = trimDelimiterPairs(s)

	trimChars := `"'.,;:>`
	s = strings.TrimRight(s, trimChars)

	s = trimUnbalancedClosers(s)

	if strings.Contains(s, `"`) {
		return ""
	}

	if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
		return ""
	}

	s = fixPercentEncoding(s)

	if len(s) <= len("https://") {
		return ""
	}

	return s
}

func trimUnbalancedClosers(s string) string {
	closers := map[rune]rune{')': '(', ']': '[', '}': '{', '>': '<'}
	runes := []rune(s)

	for len(runes) > 0 {
		lastRune := runes[len(runes)-1]
		opener, isCloser := closers[lastRune]

		if !isCloser {
			break
		}

		balance := 0
		for _, r := range runes {
			if r == opener {
				balance++
			} else if r == lastRune {
				balance--
			}
		}

		if balance < 0 {
			runes = runes[:len(runes)-1]
		} else {
			break
		}
	}

	return string(runes)
}

func fixPercentEncoding(s string) string {
	for hasDanglingPercent(s) && len(s) > 0 {
		s = s[:len(s)-1]
	}
	return s
}

func hasDanglingPercent(s string) bool {
	n := len(s)
	if n == 0 {
		return false
	}

	if s[n-1] == '%' {
		return true
	}

	if n >= 2 && s[n-2] == '%' {
		return true
	}

	if n >= 3 && s[n-3] == '%' {
		if !isHex(s[n-2]) || !isHex(s[n-1]) {
			return true
		}
	}

	return false
}

func isHex(b byte) bool {
	return (b >= '0' && b <= '9') ||
		(b >= 'a' && b <= 'f') ||
		(b >= 'A' && b <= 'F')
}

func trimDelimiterPairs(s string) string {
	pairs := []struct{ left, right byte }{
		{'"', '"'},
		{'\'', '\''},
		{'(', ')'},
		{'[', ']'},
		{'{', '}'},
		{'<', '>'},
	}

	changed := true
	for changed {
		changed = false
		for _, p := range pairs {
			if len(s) >= 2 && s[0] == p.left && s[len(s)-1] == p.right {
				s = s[1 : len(s)-1]
				changed = true
				break
			}
		}
	}

	return s
}
