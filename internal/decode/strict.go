package decode

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v3"
)

// Strict decodes YAML bytes into the target struct using strict mode,
// which rejects unknown fields.
func Strict(data []byte, target any) error {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(target); err != nil {
		return fmt.Errorf("strict YAML decode: %w", err)
	}
	return nil
}

// Relaxed decodes YAML bytes into the target struct without strict mode.
func Relaxed(data []byte, target any) error {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(target); err != nil {
		return fmt.Errorf("YAML decode: %w", err)
	}
	return nil
}
