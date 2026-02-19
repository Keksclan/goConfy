package envmacro

import "os"

// LookupFunc defines how to look up an environment variable.
type LookupFunc func(string) (string, bool)

// DefaultLookup is the standard environment lookup using os.LookupEnv.
func DefaultLookup(key string) (string, bool) {
	return os.LookupEnv(key)
}
