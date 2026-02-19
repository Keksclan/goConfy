// Package dotenv provides .env file parsing and lookup capabilities.
// It does NOT mutate the OS environment; values are kept in an internal store.
package dotenv

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// ParseFile reads and parses a .env file into a key-value map.
func ParseFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("dotenv: open %q: %w", path, err)
	}
	defer f.Close()
	return Parse(f)
}

// Parse reads .env content from a reader and returns a key-value map.
func Parse(r io.Reader) (map[string]string, error) {
	result := make(map[string]string)
	scanner := bufio.NewScanner(r)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip blank lines and comments.
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Strip optional "export " prefix.
		line = strings.TrimPrefix(line, "export ")

		key, value, err := parseLine(line)
		if err != nil {
			return nil, fmt.Errorf("dotenv: line %d: %w", lineNum, err)
		}

		result[key] = value
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("dotenv: read: %w", err)
	}

	return result, nil
}

// parseLine parses a single KEY=VALUE line.
func parseLine(line string) (string, string, error) {
	key, rest, ok := strings.Cut(line, "=")
	if !ok {
		return "", "", fmt.Errorf("missing '=' in %q", line)
	}

	key = strings.TrimSpace(key)
	if key == "" {
		return "", "", fmt.Errorf("empty key in %q", line)
	}

	value := parseValue(rest)

	return key, value, nil
}

// parseValue extracts the value from the right-hand side of a KEY=VALUE line.
// It handles quoted and unquoted values, including inline comments.
func parseValue(raw string) string {
	trimmed := strings.TrimSpace(raw)

	// Quoted values: find matching closing quote.
	if len(trimmed) >= 2 {
		if trimmed[0] == '\'' && trimmed[len(trimmed)-1] == '\'' {
			// Single quotes: literal content, no escapes.
			return trimmed[1 : len(trimmed)-1]
		}
		if trimmed[0] == '"' && trimmed[len(trimmed)-1] == '"' {
			// Double quotes: support escapes.
			return unescapeDouble(trimmed[1 : len(trimmed)-1])
		}
	}

	// Unquoted: strip inline comments (# preceded by at least one space).
	if idx := strings.Index(trimmed, " #"); idx >= 0 {
		trimmed = strings.TrimRight(trimmed[:idx], " ")
	}

	return trimmed
}

// unescapeDouble processes escape sequences inside double-quoted values.
func unescapeDouble(s string) string {
	var b strings.Builder
	b.Grow(len(s))

	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) {
			switch s[i+1] {
			case 'n':
				b.WriteByte('\n')
				i++
			case 'r':
				b.WriteByte('\r')
				i++
			case 't':
				b.WriteByte('\t')
				i++
			case '"':
				b.WriteByte('"')
				i++
			case '\\':
				b.WriteByte('\\')
				i++
			default:
				b.WriteByte(s[i])
			}
		} else {
			b.WriteByte(s[i])
		}
	}

	return b.String()
}
