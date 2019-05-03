package jsonschema

import (
	"io"
	"io/ioutil"
	"strings"

	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonschema"
)

// Validator validates an instance document against its schema
type Validator struct {
	schema gojsonschema.JSONLoader
}

// NewValidator creates a JSON schema validator instance for the given schema document
func NewValidator(schema io.ReadCloser) *Validator {
	var validator Validator
	defer schema.Close()
	body, _ := ioutil.ReadAll(schema)
	validator.schema = gojsonschema.NewBytesLoader(body)

	return &validator
}

// Validate validates the current schema against an instance JSON document
func (v *Validator) Validate(instance []byte) error {
	result, err := gojsonschema.Validate(v.schema, gojsonschema.NewBytesLoader(instance))
	if err != nil {
		return errors.Wrap(err, "error performing schema validation")
	}

	if !result.Valid() {
		errMessages := make([]string, 0, len(result.Errors()))
		for _, err := range result.Errors() {
			errMessages = append(errMessages, err.String())
		}
		return errors.Errorf("instance data is invalid: %s", strings.Join(errMessages, "\n"))
	}
	return nil
}
