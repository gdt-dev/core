// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package json

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/PaesslerAG/jsonpath"
	"gopkg.in/yaml.v3"

	"github.com/gdt-dev/core/parse"
)

var (
	// defining the JSONPath language here allows us to disaggregate parse
	// errors from runtime errors when evaluating a JSONPath expression.
	lang = jsonpath.Language()
)

// UnsupportedJSONSchemaReference returns ErrUnsupportedJSONSchemaReference for
// a supplied URL.
func UnsupportedJSONSchemaReference(url string, node *yaml.Node) error {
	return &parse.Error{
		Line:    node.Line,
		Column:  node.Column,
		Message: fmt.Sprintf("unsupported JSONSchema reference: %s", url),
	}
}

// JSONSchemaFileNotFound returns ErrJSONSchemaFileNotFound for a supplied
// path.
func JSONSchemaFileNotFound(path string, node *yaml.Node) error {
	return &parse.Error{
		Line:    node.Line,
		Column:  node.Column,
		Message: fmt.Sprintf("unable to find JSONSchema file %q", path),
	}
}

// JSONUnmarshalError returns an ErrFailure when JSON content cannot be
// decoded.
func JSONUnmarshalError(err error, node *yaml.Node) error {
	if node != nil {
		return &parse.Error{
			Line:    node.Line,
			Column:  node.Column,
			Message: fmt.Sprintf("failed to unmarshal JSON: %s", err),
		}
	}
	return &parse.Error{
		Message: fmt.Sprintf("failed to unmarshal JSON: %s", err),
	}
}

// JSONPathInvalid returns an ParseError when a JSONPath expression could not be
// parsed.
func JSONPathInvalid(path string, err error, node *yaml.Node) error {
	return &parse.Error{
		Line:    node.Line,
		Column:  node.Column,
		Message: fmt.Sprintf("JSONPath invalid: %s: %s", path, err),
	}
}

// JSONPathInvalidNoRoot returns an ErrJSONPathInvalidNoRoot when a JSONPath
// expression does not start with '$'.
func JSONPathInvalidNoRoot(path string, node *yaml.Node) error {
	return &parse.Error{
		Line:    node.Line,
		Column:  node.Column,
		Message: fmt.Sprintf("JSONPath expression %s invalid: expression must start with '$'", path),
	}
}

// UnmarshalYAML is a custom unmarshaler that ensures that JSONPath expressions
// contained in the Expect are valid.
func (e *Expect) UnmarshalYAML(node *yaml.Node) error {
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
		case "len":
			if valNode.Kind != yaml.ScalarNode {
				return parse.ExpectedScalarAt(valNode)
			}
			var v *int
			if err := valNode.Decode(&v); err != nil {
				return err
			}
			e.Len = v
		case "schema":
			if valNode.Kind != yaml.ScalarNode {
				return parse.ExpectedScalarAt(valNode)
			}
			// Ensure any JSONSchema URL specified in exponse.json.schema exists
			schemaURL := valNode.Value
			if strings.HasPrefix(schemaURL, "http://") || strings.HasPrefix(schemaURL, "https://") {
				// TODO(jaypipes): Support network lookups?
				return UnsupportedJSONSchemaReference(schemaURL, valNode)
			}
			// Convert relative filepaths to absolute filepaths rooted in the context's
			// testdir after stripping any "file://" scheme prefix
			schemaURL = strings.TrimPrefix(schemaURL, "file://")
			schemaURL, _ = filepath.Abs(schemaURL)

			f, err := os.Open(schemaURL)
			if err != nil {
				return JSONSchemaFileNotFound(schemaURL, valNode)
			}
			defer f.Close()
			if runtime.GOOS == "windows" {
				// Need to do this because of an "optimization" done in the
				// gojsonreference library:
				// https://github.com/xeipuuv/gojsonreference/blob/bd5ef7bd5415a7ac448318e64f11a24cd21e594b/reference.go#L107-L114
				e.Schema = "file:///" + schemaURL
			} else {
				e.Schema = "file://" + schemaURL
			}
		case "paths":
			if valNode.Kind != yaml.MappingNode {
				return parse.ExpectedMapAt(valNode)
			}
			paths := map[string]string{}
			if err := valNode.Decode(&paths); err != nil {
				return err
			}
			for path := range paths {
				if len(path) == 0 || path[0] != '$' {
					return JSONPathInvalidNoRoot(path, valNode)
				}
				if _, err := lang.NewEvaluable(path); err != nil {
					return JSONPathInvalid(path, err, valNode)
				}
			}
			e.Paths = paths
		case "path_formats", "path-formats":
			if valNode.Kind != yaml.MappingNode {
				return parse.ExpectedMapAt(valNode)
			}
			pathFormats := map[string]string{}
			if err := valNode.Decode(&pathFormats); err != nil {
				return err
			}
			for pathFormat := range pathFormats {
				if len(pathFormat) == 0 || pathFormat[0] != '$' {
					return JSONPathInvalidNoRoot(pathFormat, valNode)
				}
				if _, err := lang.NewEvaluable(pathFormat); err != nil {
					return JSONPathInvalid(pathFormat, err, valNode)
				}
			}
			e.PathFormats = pathFormats
		}
	}
	return nil
}
