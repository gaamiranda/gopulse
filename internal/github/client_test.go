package github

import (
	"testing"
)

func TestParseRemoteURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    *RepoInfo
		wantErr bool
	}{
		{
			name: "HTTPS URL with .git",
			url:  "https://github.com/owner/repo.git",
			want: &RepoInfo{Owner: "owner", Name: "repo"},
		},
		{
			name: "HTTPS URL without .git",
			url:  "https://github.com/owner/repo",
			want: &RepoInfo{Owner: "owner", Name: "repo"},
		},
		{
			name: "SSH URL with .git",
			url:  "git@github.com:owner/repo.git",
			want: &RepoInfo{Owner: "owner", Name: "repo"},
		},
		{
			name: "SSH URL without .git",
			url:  "git@github.com:owner/repo",
			want: &RepoInfo{Owner: "owner", Name: "repo"},
		},
		{
			name: "HTTPS URL with trailing whitespace",
			url:  "https://github.com/owner/repo.git  ",
			want: &RepoInfo{Owner: "owner", Name: "repo"},
		},
		{
			name: "Owner with hyphen",
			url:  "https://github.com/my-org/my-repo.git",
			want: &RepoInfo{Owner: "my-org", Name: "my-repo"},
		},
		{
			name: "Owner with underscore",
			url:  "https://github.com/my_org/my_repo.git",
			want: &RepoInfo{Owner: "my_org", Name: "my_repo"},
		},
		{
			name:    "Invalid URL - not GitHub",
			url:     "https://gitlab.com/owner/repo.git",
			wantErr: true,
		},
		{
			name:    "Invalid URL - missing repo",
			url:     "https://github.com/owner",
			wantErr: true,
		},
		{
			name:    "Invalid URL - empty",
			url:     "",
			wantErr: true,
		},
		{
			name:    "Invalid URL - random string",
			url:     "not a url",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseRemoteURL(tt.url)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseRemoteURL() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseRemoteURL() unexpected error: %v", err)
				return
			}

			if got.Owner != tt.want.Owner {
				t.Errorf("ParseRemoteURL() Owner = %v, want %v", got.Owner, tt.want.Owner)
			}

			if got.Name != tt.want.Name {
				t.Errorf("ParseRemoteURL() Name = %v, want %v", got.Name, tt.want.Name)
			}
		})
	}
}
