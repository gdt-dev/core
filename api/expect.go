package api

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/gdt-dev/core/parse"
)

// ExpectContains is a common struct for assertions about strings or
// collections of strings.
type ExpectContains struct {
	// ContainsAll is one or more strings that *all* must be present in the
	// subject contents
	ContainsAll *FlexStrings `yaml:"contains,omitempty"`
	// ContainsNone is one or more strings, *none of which* should be present in
	// the subject contents
	ContainsNone *FlexStrings `yaml:"contains-none-of,omitempty"`
	// ContainsOneOf is one or more strings of which *at least one* must be
	// present in the subject contents
	ContainsAny *FlexStrings `yaml:"contains-one-of,omitempty"`
}

// String returns a string representation of the ExpectContains' matching
// conditions.
func (e ExpectContains) String() string {
	parts := []string{}
	if e.ContainsAll != nil {
		parts = append(parts, fmt.Sprintf("contains-all: %s", e.ContainsAll))
	}
	if e.ContainsNone != nil {
		parts = append(parts, fmt.Sprintf("contains-none: %s", e.ContainsNone))
	}
	if e.ContainsAny != nil {
		parts = append(parts, fmt.Sprintf("contains-any: %s", e.ContainsAny))
	}
	return strings.Join(parts, " ")
}

func (e *ExpectContains) UnmarshalYAML(node *yaml.Node) error {
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
		case "all", "is", "contains", "contains-all", "contains_all":
			if valNode.Kind != yaml.ScalarNode && valNode.Kind != yaml.SequenceNode {
				return parse.ExpectedScalarOrSequenceAt(valNode)
			}
			var v FlexStrings
			if err := valNode.Decode(&v); err != nil {
				return err
			}
			e.ContainsAll = &v
		case "any", "contains-one-of", "contains-any", "contains_one_of", "contains_any":
			if valNode.Kind != yaml.ScalarNode && valNode.Kind != yaml.SequenceNode {
				return parse.ExpectedScalarOrSequenceAt(valNode)
			}
			var v FlexStrings
			if err := valNode.Decode(&v); err != nil {
				return err
			}
			e.ContainsAny = &v
		case "none", "none-of", "contains-none-of", "contains-none", "none_of", "contains_none_of", "contains_none":
			if valNode.Kind != yaml.ScalarNode && valNode.Kind != yaml.SequenceNode {
				return parse.ExpectedScalarOrSequenceAt(valNode)
			}
			var v FlexStrings
			if err := valNode.Decode(&v); err != nil {
				return err
			}
			e.ContainsNone = &v
		default:
			return parse.UnknownFieldAt(key, keyNode)
		}
	}
	return nil
}
