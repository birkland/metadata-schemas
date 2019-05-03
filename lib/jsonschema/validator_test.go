package jsonschema_test

import (
	"io/ioutil"
	"testing"

	"github.com/OA-PASS/metadata-schemas/lib/jsonschema"
	"github.com/OA-PASS/metadata-schemas/lib/schemas"
)

func TestValidate(t *testing.T) {
	schema, err := schemas.Load("jhu/global.json")
	if err != nil {
		t.Fatalf("Could not load schema: %+v", err)
	}

	validator := jsonschema.NewValidator(schema)

	instance, err := ioutil.ReadFile("../../examples/jhu/full.json")
	if err != nil {
		t.Fatalf("could not open test instance: %+v", err)
	}

	err = validator.Validate(instance)
	if err != nil {
		t.Fatalf("schema validation failed: %+v", err)
	}
}

func TestValidateInvalid(t *testing.T) {
	schema, err := schemas.Load("jhu/global.json")
	if err != nil {
		t.Fatalf("Could not load schema: %+v", err)
	}

	validator := jsonschema.NewValidator(schema)

	instance := []byte(`{
		"foo": "bar"
	}`)

	err = validator.Validate(instance)
	if err == nil {
		t.Fatalf("Was expecting error")
	}

}
