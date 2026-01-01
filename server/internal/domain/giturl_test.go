package domain

import "testing"

func TestNormalizeGitURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "SSH format with .git",
			input:    "git@github.com:user/repo.git",
			expected: "https://github.com/user/repo",
		},
		{
			name:     "SSH format without .git",
			input:    "git@github.com:user/repo",
			expected: "https://github.com/user/repo",
		},
		{
			name:     "HTTPS with .git",
			input:    "https://github.com/user/repo.git",
			expected: "https://github.com/user/repo",
		},
		{
			name:     "HTTPS without .git",
			input:    "https://github.com/user/repo",
			expected: "https://github.com/user/repo",
		},
		{
			name:     "HTTPS with trailing slash",
			input:    "https://github.com/user/repo/",
			expected: "https://github.com/user/repo",
		},
		{
			name:     "ssh:// protocol",
			input:    "ssh://git@github.com/user/repo.git",
			expected: "https://github.com/user/repo",
		},
		{
			name:     "git:// protocol",
			input:    "git://github.com/user/repo.git",
			expected: "https://github.com/user/repo",
		},
		{
			name:     "HTTP upgraded to HTTPS",
			input:    "http://github.com/user/repo",
			expected: "https://github.com/user/repo",
		},
		{
			name:     "GitLab SSH format",
			input:    "git@gitlab.com:group/project.git",
			expected: "https://gitlab.com/group/project",
		},
		{
			name:     "Self-hosted GitLab",
			input:    "git@gitlab.example.com:team/repo.git",
			expected: "https://gitlab.example.com/team/repo",
		},
		{
			name:     "Bitbucket SSH format",
			input:    "git@bitbucket.org:user/repo.git",
			expected: "https://bitbucket.org/user/repo",
		},
		{
			name:     "whitespace trimmed",
			input:    "  https://github.com/user/repo  ",
			expected: "https://github.com/user/repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NormalizeGitURL(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeGitURL(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
