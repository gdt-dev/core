// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package bar

import (
	"context"
	"strconv"

	"github.com/gdt-dev/core/api"
	"github.com/gdt-dev/core/parse"
	"github.com/gdt-dev/core/plugin"
	"github.com/samber/lo"
	"gopkg.in/yaml.v3"
)

var (
	// this is just for testing purposes...
	PluginRef = &Plugin{}
)

func init() {
	plugin.Register(PluginRef)
}

type Defaults struct {
	Foo string `yaml:"bar"`
}

func (d *Defaults) Merge(map[string]any) {}

func (d *Defaults) UnmarshalYAML(node *yaml.Node) error {
	return nil
}

type Spec struct {
	api.Spec
	Bar int `yaml:"bar"`
}

func (s *Spec) SetBase(b api.Spec) {
	s.Spec = b
}

func (s *Spec) Base() *api.Spec {
	return &s.Spec
}

func (s *Spec) Retry() *api.Retry {
	return api.NoRetry
}

func (s *Spec) Timeout() *api.Timeout {
	return nil
}

func (s *Spec) Eval(context.Context) (*api.Result, error) {
	return api.NewResult(), nil
}

func (s *Spec) UnmarshalYAML(node *yaml.Node) error {
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
		case "bar":
			if valNode.Kind != yaml.ScalarNode {
				return parse.ExpectedScalarAt(valNode)
			}
			if v, err := strconv.Atoi(valNode.Value); err != nil {
				return parse.ExpectedIntAt(valNode)
			} else {
				s.Bar = v
			}
		default:
			if lo.Contains(api.BaseSpecFields, key) {
				continue
			}
			return parse.UnknownFieldAt(key, keyNode)
		}
	}
	return nil
}

type Plugin struct{}

func (p *Plugin) Info() api.PluginInfo {
	return api.PluginInfo{
		Name: "bar",
	}
}

func (p *Plugin) Defaults() api.DefaultsHandler {
	return &Defaults{}
}

func (p *Plugin) Specs() []api.Evaluable {
	return []api.Evaluable{&Spec{}}
}
