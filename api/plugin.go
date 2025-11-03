// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package api

import "gopkg.in/yaml.v3"

// PluginInfo contains basic information about the plugin and what type of
// tests it can handle.
type PluginInfo struct {
	// Name is the primary name of the plugin
	Name string
	// Aliases is an optional set of aliased names for the plugin
	Aliases []string
	// Description describes what types of tests the plugin can handle.
	Description string
	// Timeout is a Timeout that should be used by default for test specs of
	// this plugin.
	Timeout *Timeout
	// Retry is a Retry that should be used by default for test specs of this
	// plugin.
	Retry *Retry
}

type DefaultsHandler interface {
	yaml.Unmarshaler
	// Merge merges the supplies map of key/value combinations with the set of
	// handled defaults for the plugin. The supplied key/value map will NOT be
	// unpacked from its top-most plugin named element. So, for example, the
	// kube plugin should expect to get a map that looks like
	// "kube:namespace:<namespace>" and not "namespace:<namespace>".
	Merge(map[string]any)
}

// Plugin is the driver interface for different types of gdt tests.
type Plugin interface {
	// Info returns a struct that describes what the plugin does
	Info() PluginInfo
	// Defaults returns a YAML Unmarshaler types that the plugin knows how
	// to parse its defaults configuration with.
	Defaults() DefaultsHandler
	// Specs returns a list of YAML Unmarshaler types that the plugin knows
	// how to parse.
	Specs() []Evaluable
}
