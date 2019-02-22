package jsonschema_test

import (
	"testing"
)

func TestSchemaID(t *testing.T) {
	cases := []struct {
		name     string
		schema   string
		expected string
	}{
		{"normal", `{"$id": "http://example.org/"}`, "http://example.org/"},
		{"missing", `{"foo": "bar"}`, ""},
		{"corrupt", `{"$id": {"foo": "bar"}}`, ""},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			schema := parseSchema(t, c.schema)
			if schema.ID() != c.expected {
				t.Fatalf("Did not find expected ID %s", c.expected)
			}
		})
	}
}
