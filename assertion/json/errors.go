// Use and distribution licensed under the Apache license version 2.
//
// See the COPYING file in the root project directory for full text.

package json

import (
	"fmt"

	"github.com/gdt-dev/core/api"
)

var (
	// ErrJSONPathNotFound returns an ErrFailure when a JSONPath expression
	// could not evaluate to a found element.
	ErrJSONPathNotFound = fmt.Errorf(
		"%w: failed to find element at JSONPath", api.ErrFailure,
	)
	// ErrJSONPathConversionError returns an ErrFailure when a JSONPath
	// expression evaluated to a found element but could not be converted to a
	// string.
	ErrJSONPathConversionError = fmt.Errorf(
		"%w: JSONPath value could not be compared", api.ErrFailure,
	)
	// ErrJSONPathNotEqual returns an ErrFailure when a JSONPath
	// expression evaluated to a found element but the value did not match an
	// expected string.
	ErrJSONPathNotEqual = fmt.Errorf(
		"%w: JSONPath values not equal", api.ErrFailure,
	)
	// ErrJSONSchemaValidateError returns an ErrFailure when a JSONSchema could
	// not be parsed.
	ErrJSONSchemaValidateError = fmt.Errorf(
		"%w: failed to parse JSONSchema", api.ErrFailure,
	)
	// ErrJSONSchemaInvalid returns an ErrFailure when some content could not
	// be validated with a JSONSchema.
	ErrJSONSchemaInvalid = fmt.Errorf(
		"%w: JSON content did not adhere to JSONSchema", api.ErrFailure,
	)
	// ErrJSONFormatError returns an ErrFailure when a JSONFormat expression
	// could not evaluate to a found element.
	ErrJSONFormatError = fmt.Errorf(
		"%w: failed to determine JSON format", api.ErrFailure,
	)
	// ErrJSONFormatNotEqual returns an ErrFailure when a an element at a
	// JSONPath was not in the expected format.
	ErrJSONFormatNotEqual = fmt.Errorf(
		"%w: JSON format not equal", api.ErrFailure,
	)
)

// JSONPathNotFound returns an ErrFailure when a JSONPath expression could not
// evaluate to a found element.
func JSONPathNotFound(path string, err error) error {
	if err == nil {
		return fmt.Errorf("%w: %s", ErrJSONPathNotFound, path)
	}
	return fmt.Errorf("%w: %s: %s", ErrJSONPathNotFound, path, err)
}

// JSONPathConversionError returns an ErrFailure when a JSONPath expression
// evaluated to a found element but the expected and found value types were
// incomparable.
func JSONPathConversionError(path string, exp interface{}, got interface{}) error {
	return fmt.Errorf(
		"%w: expected value of %v could not be compared to value %v at %s",
		ErrJSONPathConversionError, exp, got, path,
	)
}

// JSONPathValueNotEqual returns an ErrFailure when a JSONPath expression
// evaluated to a found element but the value did not match an expected string.
func JSONPathNotEqual(path string, exp interface{}, got interface{}) error {
	return fmt.Errorf(
		"%w: expected %v but got %v at %s",
		ErrJSONPathNotEqual, exp, got, path,
	)
}

// JSONSchemaValidateError returns an ErrFailure when a JSONSchema could not be
// parsed.
func JSONSchemaValidateError(path string, err error) error {
	return fmt.Errorf("%w %s: %s", ErrJSONSchemaValidateError, path, err)
}

// JSONSchemaInvalid returns an ErrFailure when some content could not be
// validated with a JSONSchema.
func JSONSchemaInvalid(path string, err error) error {
	return fmt.Errorf("%w %s: %s", ErrJSONSchemaInvalid, path, err)
}

// JSONFormatError returns an ErrFailure when a JSONFormat expression could not
// evaluate to a found element.
func JSONFormatError(format string, err error) error {
	return fmt.Errorf("%w %s: %s", ErrJSONFormatError, format, err)
}

// JSONFormatNotEqual returns an ErrFailure when a an element at a JSONPath was
// not in the expected format.
func JSONFormatNotEqual(path string, exp string) error {
	return fmt.Errorf(
		"%w: element at %s was not in expected JSON format %s",
		ErrJSONFormatNotEqual, path, exp,
	)
}
