package main

import (
	"net"
	"net/url"
	"strconv"
	"strings"
	"unicode"
)

var (
	hierarchicalSchemes = map[string]bool{
		"http":   true,
		"https":  true,
		"ftp":    true,
		"ssh":    true,
		"git":    true,
		"ldap":   true,
		"ldaps":  true,
		"telnet": true,
	}

	opaqueSchemes = map[string]bool{
		"mailto": true,
		"urn":    true,
		"news":   true,
		"tel":    true,
		"sip":    true,
		"sips":   true,
	}

	fileScheme = "file"
)

func IsValidRFC3986URL(u *url.URL) bool {
	if u == nil {
		return false
	}

	if u.Scheme == "" {
		return false
	}

	if !isValidScheme(u.Scheme) {
		return false
	}

	scheme := strings.ToLower(u.Scheme)

	if hierarchicalSchemes[scheme] {
		return validateHierarchicalURI(u)
	} else if opaqueSchemes[scheme] {
		return validateOpaqueURI(u)
	} else if scheme == fileScheme {
		return validateFileURI(u)
	} else {
		return validateHierarchicalURI(u)
	}
}

func validateHierarchicalURI(u *url.URL) bool {
	if u.Host == "" {
		return false
	}

	hostname := u.Hostname()
	if hostname == "" || !isValidHost(hostname) {
		return false
	}

	if u.Port() != "" {
		if !isValidPort(u.Port()) {
			return false
		}
	}

	return validateCommonComponents(u)
}

func validateOpaqueURI(u *url.URL) bool {
	if u.Host != "" || u.User != nil {
		return false
	}

	if u.Opaque == "" {
		return false
	}

	return validateCommonComponents(u)
}

func validateFileURI(u *url.URL) bool {
	hostname := u.Hostname()
	if hostname != "" && !isValidHost(hostname) {
		return false
	}

	return validateCommonComponents(u)
}

func validateCommonComponents(u *url.URL) bool {
	if !isValidPercentEncoding(u.Path) {
		return false
	}

	if !isValidPercentEncoding(u.Opaque) {
		return false
	}

	if !isValidPercentEncoding(u.RawQuery) {
		return false
	}

	if !isValidPercentEncoding(u.Fragment) {
		return false
	}

	if strings.Contains(u.RawPath, " ") {
		return false
	}

	if strings.Contains(u.Opaque, " ") {
		return false
	}

	return true
}

func isValidScheme(scheme string) bool {
	if len(scheme) == 0 {
		return false
	}

	if !isAlpha(rune(scheme[0])) {
		return false
	}

	for _, r := range scheme[1:] {
		if !isAlpha(r) && !unicode.IsDigit(r) && r != '+' && r != '-' && r != '.' {
			return false
		}
	}

	return true
}

func isValidHost(host string) bool {
	if host == "" {
		return false
	}

	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		ipv6 := host[1 : len(host)-1]
		return net.ParseIP(ipv6) != nil
	}

	if strings.Count(host, ".") == 3 && looksLikeIPv4(host) {
		return isValidIPv4(host)
	}

	if ip := net.ParseIP(host); ip != nil && strings.Contains(host, ":") {
		return true
	}

	return isValidHostname(host)
}

func looksLikeIPv4(host string) bool {
	parts := strings.Split(host, ".")
	if len(parts) != 4 {
		return false
	}

	for _, part := range parts {
		if part == "" {
			return false
		}
		for _, r := range part {
			if !unicode.IsDigit(r) {
				return false
			}
		}
	}

	return true
}

func isValidIPv4(ip string) bool {
	parts := strings.Split(ip, ".")
	if len(parts) != 4 {
		return false
	}

	for _, part := range parts {
		if part == "" {
			return false
		}

		if len(part) > 1 && part[0] == '0' {
			return false
		}

		octet, err := strconv.Atoi(part)
		if err != nil || octet < 0 || octet > 255 {
			return false
		}
	}

	return true
}

func isValidHostname(hostname string) bool {
	if hostname == "" || len(hostname) > 255 {
		return false
	}

	labels := strings.Split(hostname, ".")
	for _, label := range labels {
		if !isValidHostnameLabel(label) {
			return false
		}
	}

	return true
}

func isValidHostnameLabel(label string) bool {
	if label == "" || len(label) > 63 {
		return false
	}

	if !isAlphaNumeric(rune(label[0])) || !isAlphaNumeric(rune(label[len(label)-1])) {
		return false
	}

	for _, r := range label {
		if !isAlphaNumeric(r) && r != '-' {
			return false
		}
	}

	return true
}

func isAlphaNumeric(r rune) bool {
	return isAlpha(r) || unicode.IsDigit(r)
}

func isValidPort(port string) bool {
	if port == "" {
		return false
	}

	portNum, err := strconv.Atoi(port)
	if err != nil {
		return false
	}

	return portNum >= 1 && portNum <= 65535
}

func isValidPercentEncoding(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == '%' {
			if i+2 >= len(s) {
				return false
			}
			if !isHex(s[i+1]) || !isHex(s[i+2]) {
				return false
			}
			i += 2
		}
	}
	return true
}

func isHex(c byte) bool {
	return (c >= '0' && c <= '9') ||
		(c >= 'A' && c <= 'F') ||
		(c >= 'a' && c <= 'f')
}

func isAlpha(r rune) bool {
	return (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z')
}
