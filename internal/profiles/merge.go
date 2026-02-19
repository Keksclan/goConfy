package profiles

import "gopkg.in/yaml.v3"

// DeepMerge merges the override mapping node into the target mapping node.
// For keys that exist in both and are mappings, it recurses. Otherwise, the
// override value replaces the target value.
func DeepMerge(target, override *yaml.Node) {
	if target.Kind != yaml.MappingNode || override.Kind != yaml.MappingNode {
		return
	}

	for i := 0; i < len(override.Content)-1; i += 2 {
		overrideKey := override.Content[i]
		overrideVal := override.Content[i+1]

		found := false
		for j := 0; j < len(target.Content)-1; j += 2 {
			targetKey := target.Content[j]
			targetVal := target.Content[j+1]

			if targetKey.Value == overrideKey.Value {
				found = true
				if targetVal.Kind == yaml.MappingNode && overrideVal.Kind == yaml.MappingNode {
					DeepMerge(targetVal, overrideVal)
				} else {
					target.Content[j+1] = overrideVal
				}
				break
			}
		}

		if !found {
			target.Content = append(target.Content, overrideKey, overrideVal)
		}
	}
}
