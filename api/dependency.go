package api

import (
	"regexp"

	"github.com/Masterminds/semver/v3"
	"github.com/samber/lo"
	"gopkg.in/yaml.v3"

	"github.com/gdt-dev/core/parse"
)

var (
	ValidOSs = []string{
		"linux",
		"darwin",
		"windows",
	}
)

// Dependency describes a prerequisite binary that must be present.
type Dependency struct {
	// Name is the name of the binary that must be present.
	Name string `yaml:"name"`
	// When describes any constraining conditions that apply to this
	// Dependency.
	When *DependencyConditions `yaml:"when,omitempty"`
	// Version contains instructions for constraining and selecting the
	// dependency's version.
	Version *DependencyVersion `yaml:"version,omitempty"`
}

func (d *Dependency) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return parse.ExpectedMapAt(node)
	}
	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		if keyNode.Kind != yaml.ScalarNode {
			return parse.ExpectedScalarAt(keyNode)
		}
		key := keyNode.Value
		valNode := node.Content[i+1]
		switch key {
		case "name":
			if valNode.Kind != yaml.ScalarNode {
				return parse.ExpectedScalarAt(valNode)
			}
			d.Name = valNode.Value
		case "when":
			if valNode.Kind != yaml.MappingNode {
				return parse.ExpectedMapAt(valNode)
			}
			var when DependencyConditions
			if err := valNode.Decode(&when); err != nil {
				return err
			}
			d.When = &when
		case "version":
			if valNode.Kind != yaml.MappingNode {
				return parse.ExpectedMapAt(valNode)
			}
			var dv DependencyVersion
			if err := valNode.Decode(&dv); err != nil {
				return err
			}
			d.Version = &dv
		default:
			return parse.UnknownFieldAt(key, keyNode)
		}
	}
	return nil
}

// DependencyConditions describes constraining conditions that apply to a
// Dependency, for instance whether the dependency is only required on a
// particular OS.
type DependencyConditions struct {
	// OS indicates that the dependency only applies when the tests are run on
	// a particular operating system.
	OS string `yaml:"os,omitempty"`
}

func (c *DependencyConditions) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return parse.ExpectedMapAt(node)
	}
	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		if keyNode.Kind != yaml.ScalarNode {
			return parse.ExpectedScalarAt(keyNode)
		}
		key := keyNode.Value
		valNode := node.Content[i+1]
		switch key {
		case "os":
			if valNode.Kind != yaml.ScalarNode {
				return parse.ExpectedScalarAt(valNode)
			}
			os := valNode.Value
			if os != "" {
				if !lo.Contains(ValidOSs, os) {
					return parse.InvalidOSAt(valNode, os, ValidOSs)
				}
				c.OS = os
			}
		default:
			return parse.UnknownFieldAt(key, keyNode)
		}
	}
	return nil
}

// DependencyVersion expresses a version constraint that must be met for a
// particular dependency and instructs gdt how to get the version for a
// dependency from a binary or package manager.
type DependencyVersion struct {
	// Constraint indicates a version constraint to apply to the dependency,
	// e.g.  '>= 1.2.3' would indicate that a version of the dependency binary
	// after and including 1.2.3 must be present on the host.
	Constraint        string              `yaml:"constraint"`
	SemVerConstraints *semver.Constraints `yaml:"-"`
	// Selector provides instructions to select the version from the binary.
	Selector *DependencyVersionSelector `yaml:"selector,omitempty"`
}

func (v *DependencyVersion) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return parse.ExpectedMapAt(node)
	}
	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		if keyNode.Kind != yaml.ScalarNode {
			return parse.ExpectedScalarAt(keyNode)
		}
		key := keyNode.Value
		valNode := node.Content[i+1]
		switch key {
		case "constraint":
			if valNode.Kind != yaml.ScalarNode {
				return parse.ExpectedScalarAt(valNode)
			}
			conStr := valNode.Value
			if conStr != "" {
				con, err := semver.NewConstraint(conStr)
				if err != nil {
					return parse.InvalidVersionConstraintAt(
						valNode, conStr, err,
					)
				}
				v.Constraint = conStr
				v.SemVerConstraints = con
			}
		case "selector":
			if valNode.Kind != yaml.MappingNode {
				return parse.ExpectedMapAt(valNode)
			}
			var selector DependencyVersionSelector
			if err := valNode.Decode(&selector); err != nil {
				return err
			}
			v.Selector = &selector
		default:
			return parse.UnknownFieldAt(key, keyNode)
		}
	}
	return nil
}

// DependencyVersionSelector instructs gdt how to get the version of a binary.
type DependencyVersionSelector struct {
	// Args is the command-line to execute the dependency binary to output
	// version information, e.g. '-v' or '--version-json'.
	Args []string `yaml:"args,omitempty"`
	// Filter is an optional regex to run against the output returned by
	// Command, e.g. 'v?(\d)+\.(\d+)(\.(\d)+)?'.
	Filter      string         `yaml:"filter,omitempty"`
	FilterRegex *regexp.Regexp `yaml:"-"`
}

func (s *DependencyVersionSelector) UnmarshalYAML(node *yaml.Node) error {
	if node.Kind != yaml.MappingNode {
		return parse.ExpectedMapAt(node)
	}
	for i := 0; i < len(node.Content); i += 2 {
		keyNode := node.Content[i]
		if keyNode.Kind != yaml.ScalarNode {
			return parse.ExpectedScalarAt(keyNode)
		}
		key := keyNode.Value
		valNode := node.Content[i+1]
		switch key {
		case "args":
			if valNode.Kind != yaml.SequenceNode {
				return parse.ExpectedSequenceAt(valNode)
			}
			var args []string
			if err := valNode.Decode(&args); err != nil {
				return err
			}
			s.Args = args
		case "filter":
			if valNode.Kind != yaml.ScalarNode {
				return parse.ExpectedMapAt(valNode)
			}
			filter := valNode.Value
			if filter != "" {
				re, err := regexp.Compile(filter)
				if err != nil {
					return parse.InvalidRegexAt(valNode, filter, err)
				}
				s.Filter = filter
				s.FilterRegex = re
			}
		default:
			return parse.UnknownFieldAt(key, keyNode)
		}
	}
	return nil
}
