package profiles

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func FuzzDeepMerge(f *testing.F) {
	seeds := []struct {
		target   string
		override string
	}{
		{"k1: v1", "k1: v2"},
		{"k1: v1", "k2: v2"},
		{"k1: {s1: v1}", "k1: {s1: v2}"},
		{"k1: {s1: v1}", "k1: {s2: v2}"},
		{"k1: v1", "k1: {s1: v1}"},
		{"k1: {s1: v1}", "k1: v1"},
		{"k1: [v1]", "k1: [v2]"},
		{"", "k1: v1"},
		{"k1: v1", ""},
	}

	for _, seed := range seeds {
		f.Add(seed.target, seed.override, "dev")
	}

	f.Fuzz(func(t *testing.T, targetYaml, overrideYaml string, profileName string) {
		var targetNode, overrideNode yaml.Node
		if err := yaml.Unmarshal([]byte(targetYaml), &targetNode); err != nil {
			return
		}
		if err := yaml.Unmarshal([]byte(overrideYaml), &overrideNode); err != nil {
			return
		}

		// Ensure we are dealing with mappings at the root of the document
		if len(targetNode.Content) == 0 || targetNode.Content[0].Kind != yaml.MappingNode {
			return
		}
		if len(overrideNode.Content) == 0 || overrideNode.Content[0].Kind != yaml.MappingNode {
			return
		}

		// Test DeepMerge directly
		DeepMerge(targetNode.Content[0], overrideNode.Content[0])

		// Test ApplyProfile
		var rootNode yaml.Node
		if err := yaml.Unmarshal([]byte(targetYaml), &rootNode); err == nil {
			_ = ApplyProfile(&rootNode, profileName)
		}
	})
}
