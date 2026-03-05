package yamltemplate

import "strings"

// buildComment generates the YAML comment lines for a field based on its tags.
// Each comment line is returned without the leading "# " prefix.
func buildComment(fi fieldInfo) string {
	var lines []string

	if fi.Desc != "" {
		lines = append(lines, fi.Desc)
	}
	if fi.Env != "" {
		lines = append(lines, "env: "+fi.Env)
	}
	if fi.Required {
		lines = append(lines, "required: true")
	}
	if fi.Secret {
		lines = append(lines, "secret: true")
	}
	if fi.Example != "" {
		lines = append(lines, "example: "+fi.Example)
	}
	if fi.Sep != "" {
		lines = append(lines, "sep: "+fi.Sep)
	}

	return strings.Join(lines, "\n")
}
