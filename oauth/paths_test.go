package oauth

import "testing"

func TestPaths_LoginURL(t *testing.T) {
	const root = "http://example.com"

	tests := []struct {
		name  string
		paths Paths
		want  string
	}{
		{name: "no slash", paths: Paths{RootURL: root, LoginPath: "login"}, want: "http://example.com/login"},
		{name: "root slash", paths: Paths{RootURL: root + "/", LoginPath: "login"}, want: "http://example.com/login"},
		{name: "path slash", paths: Paths{RootURL: root, LoginPath: "/login"}, want: "http://example.com/login"},
		{name: "both slashes", paths: Paths{RootURL: root + "/", LoginPath: "/login"}, want: "http://example.com/login"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.paths.LoginURL("").String(); got != tt.want {
				t.Errorf("Paths.LoginURL() = %v, want %v", got, tt.want)
			}
		})
	}
}
