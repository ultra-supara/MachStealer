package main

import (
	"os/user"
	"path/filepath"
	"testing"
)

func TestGetDefaultPath(t *testing.T) {
	usr, err := user.Current()
	if err != nil {
		t.Skipf("Cannot get current user: %v", err)
	}

	basePath := filepath.Join(usr.HomeDir, "Library/Application Support/Google/Chrome/Default")

	tests := []struct {
		name string
		kind string
		want string
	}{
		{
			name: "cookie path",
			kind: "cookie",
			want: filepath.Join(basePath, "Cookies"),
		},
		{
			name: "logindata path",
			kind: "logindata",
			want: filepath.Join(basePath, "Login Data"),
		},
		{
			name: "invalid kind returns empty",
			kind: "invalid",
			want: "",
		},
		{
			name: "empty kind returns empty",
			kind: "",
			want: "",
		},
		{
			name: "random kind returns empty",
			kind: "somethingelse",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getDefaultPath(tt.kind)
			if got != tt.want {
				t.Errorf("getDefaultPath(%q) = %q, want %q", tt.kind, got, tt.want)
			}
		})
	}
}

func TestGetDefaultPathAllCases(t *testing.T) {
	// Test all switch cases
	kinds := []string{"cookie", "logindata", "other"}
	for _, kind := range kinds {
		result := getDefaultPath(kind)
		if kind == "other" {
			if result != "" {
				t.Errorf("getDefaultPath(%q) should return empty string, got %q", kind, result)
			}
		} else {
			if result == "" {
				t.Errorf("getDefaultPath(%q) should not return empty string", kind)
			}
		}
	}
}

func BenchmarkGetDefaultPath(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = getDefaultPath("cookie")
	}
}
