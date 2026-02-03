package utils

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// GitRemoteInfo contains parsed information from a git remote URL
type GitRemoteInfo struct {
	Host        string // e.g., "gitlab.com" or "gitlab.example.com"
	ProjectPath string // e.g., "group/project" or "group/subgroup/project"
	URL         string // The full GitLab URL, e.g., "https://gitlab.com"
}

// DetectGitRemote attempts to detect GitLab URL and project path from git remote.
// It tries the "origin" remote first.
// Returns nil if detection fails (not a git repo, no remote, not a GitLab URL, etc.)
func DetectGitRemote() *GitRemoteInfo {
	// Try to get the origin remote URL
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return nil
	}

	remoteURL := strings.TrimSpace(string(output))
	if remoteURL == "" {
		return nil
	}

	return ParseGitRemoteURL(remoteURL)
}

// ParseGitRemoteURL parses a git remote URL and extracts host and project path.
// Supports both SSH and HTTPS formats:
//   - SSH:   git@gitlab.com:group/project.git
//   - HTTPS: https://gitlab.com/group/project.git
//
// Returns nil if the URL cannot be parsed.
func ParseGitRemoteURL(remoteURL string) *GitRemoteInfo {
	// Try SSH format: git@host:path.git
	sshRegex := regexp.MustCompile(`^git@([^:]+):(.+?)(?:\.git)?$`)
	if matches := sshRegex.FindStringSubmatch(remoteURL); matches != nil {
		host := matches[1]
		projectPath := matches[2]
		return &GitRemoteInfo{
			Host:        host,
			ProjectPath: projectPath,
			URL:         fmt.Sprintf("https://%s", host),
		}
	}

	// Try HTTPS format: https://host/path.git or https://host/path
	httpsRegex := regexp.MustCompile(`^https?://([^/]+)/(.+?)(?:\.git)?$`)
	if matches := httpsRegex.FindStringSubmatch(remoteURL); matches != nil {
		host := matches[1]
		projectPath := matches[2]
		return &GitRemoteInfo{
			Host:        host,
			ProjectPath: projectPath,
			URL:         fmt.Sprintf("https://%s", host),
		}
	}

	return nil
}
