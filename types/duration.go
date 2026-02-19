package types

import (
	"fmt"
	"time"

	"gopkg.in/yaml.v3"
)

// Duration is a wrapper around time.Duration that supports YAML marshaling/unmarshaling.
type Duration time.Duration

// UnmarshalYAML implements yaml.Unmarshaler.
func (d *Duration) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.ScalarNode {
		return fmt.Errorf("duration must be a scalar, got %v", node.Kind)
	}

	dur, err := time.ParseDuration(node.Value)
	if err != nil {
		return fmt.Errorf("failed to parse duration: %w", err)
	}

	*d = Duration(dur)
	return nil
}

// MarshalYAML implements yaml.Marshaler.
func (d Duration) MarshalYAML() (any, error) {
	return time.Duration(d).String(), nil
}

// String returns the string representation of the duration.
func (d Duration) String() string {
	return time.Duration(d).String()
}
