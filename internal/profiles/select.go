package profiles

import (
	"os"

	"gopkg.in/yaml.v3"
)

// SelectProfile determines which profile to use. It checks, in order:
// 1. An explicitly provided profile name
// 2. The value of the given environment variable
// Returns the profile name or empty string if none selected.
func SelectProfile(explicit string, envVar string) string {
	if explicit != "" {
		return explicit
	}
	if envVar != "" {
		return os.Getenv(envVar)
	}
	return ""
}

// ApplyProfile looks for a "profiles" mapping in the root node, extracts the
// selected profile's overrides, deep-merges them into the root, and removes
// the "profiles" key from the tree.
func ApplyProfile(root *yaml.Node, profileName string) error {
	if root == nil || profileName == "" {
		return removeProfilesKey(root)
	}

	mapping := root
	if root.Kind == yaml.DocumentNode && len(root.Content) > 0 {
		mapping = root.Content[0]
	}

	if mapping.Kind != yaml.MappingNode {
		return nil
	}

	// Find "profiles" key
	profileIdx := -1
	for i := range len(mapping.Content) / 2 {
		idx := i * 2
		if mapping.Content[idx].Value == "profiles" {
			profileIdx = idx
			break
		}
	}

	if profileIdx < 0 {
		return nil
	}

	profilesNode := mapping.Content[profileIdx+1]
	if profilesNode.Kind != yaml.MappingNode {
		return removeProfilesKey(root)
	}

	// Find the selected profile
	var overrideNode *yaml.Node
	for i := range len(profilesNode.Content) / 2 {
		idx := i * 2
		if profilesNode.Content[idx].Value == profileName {
			overrideNode = profilesNode.Content[idx+1]
			break
		}
	}

	// Remove "profiles" key before merging
	mapping.Content = append(mapping.Content[:profileIdx], mapping.Content[profileIdx+2:]...)

	if overrideNode != nil && overrideNode.Kind == yaml.MappingNode {
		DeepMerge(mapping, overrideNode)
	}

	return nil
}

func removeProfilesKey(root *yaml.Node) error {
	if root == nil {
		return nil
	}

	mapping := root
	if root.Kind == yaml.DocumentNode && len(root.Content) > 0 {
		mapping = root.Content[0]
	}

	if mapping.Kind != yaml.MappingNode {
		return nil
	}

	for i := range len(mapping.Content) / 2 {
		idx := i * 2
		if mapping.Content[idx].Value == "profiles" {
			mapping.Content = append(mapping.Content[:idx], mapping.Content[idx+2:]...)
			return nil
		}
	}

	return nil
}
