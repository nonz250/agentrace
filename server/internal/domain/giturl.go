package domain

import (
	"regexp"
	"strings"
)

// NormalizeGitURL converts various git URL formats to a canonical HTTPS format.
// Examples:
//   - git@github.com:user/repo.git -> https://github.com/user/repo
//   - https://github.com/user/repo.git -> https://github.com/user/repo
//   - ssh://git@github.com/user/repo.git -> https://github.com/user/repo
//   - "" -> ""
func NormalizeGitURL(url string) string {
	if url == "" {
		return ""
	}

	url = strings.TrimSpace(url)

	// Handle SSH format: git@host:user/repo.git
	sshRegex := regexp.MustCompile(`^git@([^:]+):(.+)$`)
	if matches := sshRegex.FindStringSubmatch(url); matches != nil {
		host := matches[1]
		path := matches[2]
		url = "https://" + host + "/" + path
	}

	// Handle ssh:// protocol: ssh://git@host/user/repo.git
	if strings.HasPrefix(url, "ssh://") {
		url = strings.TrimPrefix(url, "ssh://")
		// Remove git@ prefix if present
		if strings.HasPrefix(url, "git@") {
			url = strings.TrimPrefix(url, "git@")
		}
		url = "https://" + url
	}

	// Handle git:// protocol: git://host/user/repo.git
	if strings.HasPrefix(url, "git://") {
		url = strings.Replace(url, "git://", "https://", 1)
	}

	// Ensure https:// prefix
	if !strings.HasPrefix(url, "https://") && !strings.HasPrefix(url, "http://") {
		url = "https://" + url
	}

	// Upgrade http to https
	if strings.HasPrefix(url, "http://") {
		url = strings.Replace(url, "http://", "https://", 1)
	}

	// Remove .git suffix
	url = strings.TrimSuffix(url, ".git")

	// Remove trailing slash
	url = strings.TrimSuffix(url, "/")

	return url
}
