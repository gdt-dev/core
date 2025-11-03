// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package exec

import (
	"github.com/gdt-dev/core/parse"
	"gopkg.in/yaml.v3"
)

type execDefaults struct{}

// Defaults is the known exec plugin defaults collection
type Defaults struct {
	execDefaults
}

// Merge merges the supplies map of key/value combinations with the set of
// handled defaults for the plugin. The supplied key/value map will NOT be
// unpacked from its top-most plugin named element. So, for example, the
// kube plugin should expect to get a map that looks like
// "kube:namespace:<namespace>" and not "namespace:<namespace>".
func (d *Defaults) Merge(map[string]any) {}

func (d *Defaults) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return parse.ExpectedMapAt(node)
	}
	// maps/structs are stored in a top-level Node.Content field which is a
	// concatenated slice of Node pointers in pairs of key/values.
	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		if keyNode.Kind != yaml.ScalarNode {
			return parse.ExpectedScalarAt(keyNode)
		}
		key := keyNode.Value
		valNode := node.Content[i+1]
		switch key {
		case "exec":
			if valNode.Kind != yaml.MappingNode {
				return parse.ExpectedMapAt(valNode)
			}
			ed := execDefaults{}
			if err := valNode.Decode(&ed); err != nil {
				return err
			}
			d.execDefaults = ed
		default:
			continue
		}
	}
	return nil
}
