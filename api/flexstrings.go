// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package api

import (
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/gdt-dev/core/parse"
)

// FlexStrings is a struct used to parse an interface{} that can be either a
// string or a slice of strings.
type FlexStrings struct {
	values []string `yaml:"-"`
}

// Values returns the contained collection of string values.
func (f *FlexStrings) Values() []string {
	return f.values
}

// NewFlexStrings returns a new FlexStrings from zero or more string args.
func NewFlexStrings(args ...string) *FlexStrings {
	return &FlexStrings{values: args}
}

// String returns a string representation of the FlexStrings
func (f FlexStrings) String() string {
	switch len(f.values) {
	case 0:
		return ""
	case 1:
		return f.values[0]
	default:
		return "[" + strings.Join(f.values, ",") + "]"
	}
}

// UnmarshalYAML is a custom unmarshaler that understands that the value of the
// FlexStrings can be either a string or a slice of strings.
func (f *FlexStrings) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.ScalarNode && node.Kind != yaml.SequenceNode {
		return parse.ExpectedScalarOrSequenceAt(node)
	}
	var single string
	if err := node.Decode(&single); err == nil {
		f.values = []string{single}
		return nil
	}
	var many []string
	if err := node.Decode(&many); err == nil {
		f.values = many
		return nil
	}
	return parse.ExpectedScalarOrSequenceAt(node)
}
