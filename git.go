package main

import (
	"net/url"
	"regexp"
	"strings"
)

var (
	// SCP-style Git remote pattern components
	scpStart        = `^`                   // Anchor to start of string
	scpUser         = `[a-zA-Z0-9_.-]+`     // Username: letters, digits, underscore, dot, dash
	scpUserOptional = `(` + scpUser + `@)?` // Optional capture group: username followed by @
	scpHost         = `([a-zA-Z0-9_.-]+)`   // Capture group: hostname (same chars as username)
	scpSeparator    = `:`                   // Colon separator (distinguishes from URL)
	scpPath         = `([^/].*/?)`          // Capture group: path not starting with /, optional trailing /
	scpEnd          = `$`                   // Anchor to end of string

	// Complete SCP-style pattern: matches git@github.com:user/repo.git format
	scpPattern   = scpStart + scpUserOptional + scpHost + scpSeparator + scpPath + scpEnd
	scpLikeRegex = regexp.MustCompile(scpPattern)
)

func IsGitRemote(s string) bool {
	if strings.HasPrefix(s, "ssh://") || strings.HasPrefix(s, "git://") {
		return true
	}
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		if strings.HasSuffix(strings.ToLower(s), ".git") || strings.HasSuffix(strings.ToLower(s), ".git/") {
			return true
		}
		return false
	}

	if scpLikeRegex.MatchString(s) {
		parts := scpLikeRegex.FindStringSubmatch(s)
		if len(parts) > 3 && strings.Contains(parts[3], "/") {
			return true
		}
	}

	return false
}

func TransformGitRemote(s string) (string, bool) {
	var transformed string

	if matches := scpLikeRegex.FindStringSubmatch(s); len(matches) > 0 && strings.Contains(matches[3], "/") {
		host := matches[2]
		path := matches[3]
		transformed = "https://" + host + "/" + path
	} else {
		u, err := url.Parse(s)
		if err != nil {
			return "", false
		}
		host := u.Hostname()
		if host == "" {
			host = u.Host
		}
		if host == "" {
			return "", false
		}
		path := u.Path
		if path == "" {
			path = "/"
		}
		transformed = "https://" + host + path
	}

	if strings.HasSuffix(strings.ToLower(transformed), ".git") {
		transformed = transformed[:len(transformed)-4]
	}
	transformed = strings.TrimSuffix(transformed, "/")

	if _, err := url.ParseRequestURI(transformed); err != nil {
		return "", false
	}

	return transformed, true
}
