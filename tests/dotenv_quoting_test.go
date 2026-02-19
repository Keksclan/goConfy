package tests

import (
	"strings"
	"testing"

	"github.com/keksclan/goConfy/internal/dotenv"
)

func TestDotenvQuotingEmptyDoubleQuoted(t *testing.T) {
	vars, err := dotenv.Parse(strings.NewReader(`KEY=""`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vars["KEY"] != "" {
		t.Errorf("expected empty string, got %q", vars["KEY"])
	}
}

func TestDotenvQuotingEmptySingleQuoted(t *testing.T) {
	vars, err := dotenv.Parse(strings.NewReader(`KEY=''`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vars["KEY"] != "" {
		t.Errorf("expected empty string, got %q", vars["KEY"])
	}
}

func TestDotenvQuotingHashInDoubleQuotes(t *testing.T) {
	vars, err := dotenv.Parse(strings.NewReader(`KEY="a#b"`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vars["KEY"] != "a#b" {
		t.Errorf("expected %q, got %q", "a#b", vars["KEY"])
	}
}

func TestDotenvQuotingHashNoSpaceUnquoted(t *testing.T) {
	// No space before #, so not an inline comment.
	vars, err := dotenv.Parse(strings.NewReader(`KEY=a#b`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vars["KEY"] != "a#b" {
		t.Errorf("expected %q, got %q", "a#b", vars["KEY"])
	}
}

func TestDotenvQuotingInlineComment(t *testing.T) {
	vars, err := dotenv.Parse(strings.NewReader(`KEY=a #comment`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vars["KEY"] != "a" {
		t.Errorf("expected %q, got %q", "a", vars["KEY"])
	}
}

func TestDotenvQuotingSpacesInDoubleQuotes(t *testing.T) {
	vars, err := dotenv.Parse(strings.NewReader(`KEY="  "`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vars["KEY"] != "  " {
		t.Errorf("expected two spaces, got %q", vars["KEY"])
	}
}

func TestDotenvQuotingSpacesInSingleQuotes(t *testing.T) {
	vars, err := dotenv.Parse(strings.NewReader(`KEY='  '`))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vars["KEY"] != "  " {
		t.Errorf("expected two spaces, got %q", vars["KEY"])
	}
}

func TestDotenvMalformedLineError(t *testing.T) {
	_, err := dotenv.Parse(strings.NewReader("NOEQUALS"))
	if err == nil {
		t.Fatal("expected error for malformed line")
	}
	if !strings.Contains(err.Error(), "line 1") {
		t.Errorf("expected error to contain line number, got: %v", err)
	}
}

func TestDotenvUnquotedTrimSpaces(t *testing.T) {
	vars, err := dotenv.Parse(strings.NewReader(`KEY=  hello  `))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if vars["KEY"] != "hello" {
		t.Errorf("expected %q, got %q", "hello", vars["KEY"])
	}
}
