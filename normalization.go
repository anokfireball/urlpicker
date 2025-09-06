package main

import (
	"net/url"
	"strconv"
	"strings"

	"golang.org/x/net/idna"
)

func NormalizeHost(host string) string {
	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		return host
	}

	colonIndex := strings.LastIndex(host, ":")
	var hostPart, portPart string
	if colonIndex != -1 {
		hostPart = host[:colonIndex]
		portPart = host[colonIndex:]
	} else {
		hostPart = host
	}

	hostPart = strings.ToLower(hostPart)

	if punycodeHost, err := idna.ToASCII(hostPart); err == nil {
		hostPart = punycodeHost
	}

	return hostPart + portPart
}

func RemoveCredentials(urlStr string) string {
	if IsGitRemote(urlStr) {
		if transformed, ok := TransformGitRemote(urlStr); ok {
			return transformed
		}
		return urlStr
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return urlStr
	}

	u.User = nil

	return u.String()
}

func NormalizeURL(urlStr string) string {
	trimmed := TrimWrapperPunctuation(urlStr)
	withoutCreds := RemoveCredentials(trimmed)

	u, err := url.Parse(withoutCreds)
	if err != nil {
		return withoutCreds
	}

	if u.Host != "" {
		u.Host = NormalizeHost(u.Host)
	}

	u.Scheme = strings.ToLower(u.Scheme)

	u.Host = removeDefaultPort(u.Scheme, u.Host)

	if u.RawPath != "" {
		u.RawPath = decodeUnreservedChars(u.RawPath)
	} else if u.Path != "" {
		u.Path = decodeUnreservedChars(u.Path)
	}

	if u.Path != "" && (strings.Contains(u.Path, "/./") || strings.Contains(u.Path, "/../") || strings.HasSuffix(u.Path, "/.") || strings.HasSuffix(u.Path, "/..")) {
		u.Path = normalizeDotSegments(u.Path)
	}

	return u.String()
}

func normalizeDotSegments(path string) string {
	segments := strings.Split(path, "/")
	var result []string

	for _, segment := range segments {
		switch segment {
		case "", ".":
			if segment == "" && len(result) == 0 {
				result = append(result, "")
			}
		case "..":
			if len(result) > 1 {
				result = result[:len(result)-1]
			}
		default:
			result = append(result, segment)
		}
	}

	normalized := strings.Join(result, "/")
	if strings.HasSuffix(path, "/") && !strings.HasSuffix(normalized, "/") && normalized != "" {
		normalized += "/"
	}
	return normalized
}

func removeDefaultPort(scheme, host string) string {
	defaultPorts := map[string]string{
		"http":  "80",
		"https": "443",
		"ftp":   "21",
		"ssh":   "22",
	}

	if defaultPort, exists := defaultPorts[scheme]; exists {
		if strings.HasSuffix(host, ":"+defaultPort) {
			return strings.TrimSuffix(host, ":"+defaultPort)
		}
	}
	return host
}

func decodeUnreservedChars(s string) string {
	result := strings.Builder{}
	for i := 0; i < len(s); i++ {
		if s[i] == '%' && i+3 <= len(s) {
			if hex := s[i+1 : i+3]; len(hex) == 2 {
				if val, err := strconv.ParseUint(hex, 16, 8); err == nil {
					char := byte(val)
					if (char >= 'A' && char <= 'Z') || (char >= 'a' && char <= 'z') ||
						(char >= '0' && char <= '9') || char == '-' || char == '.' ||
						char == '_' || char == '~' {
						result.WriteByte(char)
						i += 2
						continue
					}
				}
			}
		}
		result.WriteByte(s[i])
	}
	return result.String()
}
