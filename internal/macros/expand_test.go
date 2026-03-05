package macros

import (
	"testing"
)

func TestParseFileMacro(t *testing.T) {
	tests := []struct {
		input         string
		expectPath    string
		expectDefault string
		expectHasDef  bool
	}{
		{"/etc/passwd", "/etc/passwd", "", false},
		{"/etc/passwd:default", "/etc/passwd", "default", true},
		{"C:\\secrets\\pwd", "C:\\secrets\\pwd", "", false},
		{"C:\\secrets\\pwd:default", "C:\\secrets\\pwd", "default", true},
		{"C:foo", "C:foo", "", false},
		{"C:foo:default", "C:foo", "default", true},
		{"/path/with:colon", "/path/with", "colon", true},
		{":default", "", "default", true},
		{":", "", "", true},
		{"", "", "", false},
	}

	for _, tt := range tests {
		path, def, hasDef := parseFileMacro(tt.input)
		if path != tt.expectPath {
			t.Errorf("input=%q: expected path=%q, got %q", tt.input, tt.expectPath, path)
		}
		if def != tt.expectDefault {
			t.Errorf("input=%q: expected default=%q, got %q", tt.input, tt.expectDefault, def)
		}
		if hasDef != tt.expectHasDef {
			t.Errorf("input=%q: expected hasDef=%v, got %v", tt.input, tt.expectHasDef, hasDef)
		}
	}
}

func TestFileMacroPattern(t *testing.T) {
	tests := []struct {
		input string
		match bool
	}{
		{"{FILE:/etc/passwd}", true},
		{"{FILE:C:\\secrets\\pwd}", true},
		{"{FILE:C:\\secrets\\pwd:default}", true},
		{"{FILE:/path/with:colon}", true},
		// Must NOT match
		{"http://{FILE:foo}/x", false},
		{"prefix{FILE:/etc/passwd}", false},
		{"{FILE:/etc/passwd}suffix", false},
		{"plaintext", false},
		{"", false},
	}

	for _, tt := range tests {
		matched := FileMacroRegex.MatchString(tt.input)
		if matched != tt.match {
			t.Errorf("input=%q: expected match=%v, got %v", tt.input, tt.match, matched)
		}
	}
}

func TestFieldError(t *testing.T) {
	tests := []struct {
		name     string
		err      *FieldError
		expected string
	}{
		{
			name:     "path only",
			err:      &FieldError{Path: "some.path", Message: "some error"},
			expected: `path "some.path": some error`,
		},
		{
			name:     "path and line",
			err:      &FieldError{Path: "some.path", Line: 3, Message: "some error"},
			expected: `path "some.path": line 3: some error`,
		},
		{
			name:     "path, line, column",
			err:      &FieldError{Path: "some.path", Line: 3, Column: 5, Message: "some error"},
			expected: `path "some.path": line 3, col 5: some error`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, tt.err.Error())
			}
		})
	}
}
