package dotenv

// Store holds parsed .env key-value pairs and provides a lookup function.
type Store struct {
	vars map[string]string
}

// NewStore creates a Store from parsed key-value pairs.
func NewStore(vars map[string]string) *Store {
	return &Store{vars: vars}
}

// Lookup returns the value for the given key, if present.
func (s *Store) Lookup(key string) (string, bool) {
	val, ok := s.vars[key]
	return val, ok
}

// ChainLookup builds a combined lookup that tries primary first, then fallback.
// This is the extension point for composing OS env with dotenv lookups.
func ChainLookup(primary, fallback func(string) (string, bool)) func(string) (string, bool) {
	return func(key string) (string, bool) {
		if val, ok := primary(key); ok {
			return val, true
		}
		return fallback(key)
	}
}
