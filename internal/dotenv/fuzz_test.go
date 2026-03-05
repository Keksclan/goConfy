package dotenv

import (
	"strings"
	"testing"
)

func FuzzParse(f *testing.F) {
	seeds := []string{
		"KEY=VALUE",
		"KEY=\"quoted value\"",
		"KEY='single quoted'",
		"export KEY=VALUE",
		"# comment line",
		"KEY=VALUE # inline comment",
		"KEY=\"quoted value # with hash\"",
		"EMPTY_VALUE=",
		"  KEY  =  VALUE  ",
		"KEY=\"escaped \\\" quotes\"",
		"KEY=\"escaped \\n newline\"",
		"=",
		"INVALID",
		"KEY =",
		"= VALUE",
		"\"KEY\"=VALUE",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		r := strings.NewReader(input)
		// We expect Parse never to panic
		_, _ = Parse(r)
	})
}
