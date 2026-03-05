package profiles

import "gopkg.in/yaml.v3"

// DeepMerge merges the override mapping node into the target mapping node.
// For keys that exist in both and are mappings, it recurses. Otherwise, the
// override value replaces the target value.
func DeepMerge(target, override *yaml.Node) {
	if target.Kind != yaml.MappingNode || override.Kind != yaml.MappingNode {
		return
	}

	for i := range len(override.Content) / 2 {
		idx := i * 2
		overrideKey := override.Content[idx]
		overrideVal := override.Content[idx+1]

		found := false
		for j := range len(target.Content) / 2 {
			jdx := j * 2
			targetKey := target.Content[jdx]
			targetVal := target.Content[jdx+1]

			if targetKey.Value == overrideKey.Value {
				found = true
				if targetVal.Kind == yaml.MappingNode && overrideVal.Kind == yaml.MappingNode {
					DeepMerge(targetVal, overrideVal)
				} else {
					target.Content[jdx+1] = overrideVal
				}
				break
			}
		}

		if !found {
			target.Content = append(target.Content, overrideKey, overrideVal)
		}
	}
}
