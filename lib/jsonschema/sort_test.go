package jsonschema_test

import (
	"strings"
	"testing"

	"github.com/OA-PASS/metadata-schemas/lib/jsonschema"
	"github.com/go-test/deep"
)

func TestSort(t *testing.T) {

	one := parseSchema(t, `{
		"$id": "http://example.org/schemas/one.json",
		"definitions": {
			"form": {
				"properties": {
					"foo": "bar"
				}
			}
		}
	}`)
	two := parseSchema(t, `{
		"$id": "http://example.org/schemas/two.json",
		"definitions": {
			"form": {
				"properties": {
					"foo": {"$ref": "one.json#/definitions/form/properties/foo"},
					"bar": "baz",
					"baz": {"$ref": "#/definitions/form/properties/bar"}
				}
			}
		}
	}`)
	three := parseSchema(t, `{
		"$id": "http://example.org/schemas/three.json",
		"definitions": {
			"form": {
				"properties": {
					"foo": {"$ref": "one.json#/definitions/form/properties/foo"},
					"bar": {"$ref": "two.json#/definitions/form/properties/foo"},
					"baz": "value"
				}
			}
		}
	}`)
	four := parseSchema(t, `{
		"$id": "http://example.org/schemas/four.json",
		"definitions": {
			"form": {
				"properties": {
					"foo2": {"$ref": "one.json#/definitions/form/properties/foo"},
					"bar2": {"$ref": "two.json#/definitions/form/properties/foo"},
					"baz": "value"
				}
			}
		}
	}`)
	five := parseSchema(t, `{
		"$id": "http://example.org/schemas/five.json",
		"definitions": {
			"form": {
				"properties": {
					"one": 1,
					"two": 2
				}
			}
		}
	}`)
	six := parseSchema(t, `{
		"$id": "http://example.org/schemas/six.json",
		"definitions": {
			"form": {
				"properties": {
					"one": 1
				}
			}
		}
	}`)
	seven := parseSchema(t, `{
		"$id": "http://example.org/schemas/seven.json"
	}`)

	sorted, err := jsonschema.Sorted([]jsonschema.Instance{five, two, seven, one, six, three, four})
	if err != nil {
		t.Fatalf("Error sorting schemas %+v", err)
	}

	expected := []jsonschema.Instance{one, two, three, four, five, six, seven}

	diffs := deep.Equal(sorted, expected)
	if len(diffs) > 0 {
		t.Fatalf("Found differences in sorted schemas %s", strings.Join(diffs, "\n"))
	}
}

func TestSortErrors(t *testing.T) {
	cases := map[string][]jsonschema.Instance{
		"nil": {nil},
		"refError": {parseSchema(t, `{
			"$id": "0http://example.org/bad.json",
			"foo": {"$ref": "something.json#/foo/bar"}
		}`)},
	}

	for name, schemas := range cases {
		schemas := schemas
		t.Run(name, func(t *testing.T) {
			_, err := jsonschema.Sorted(schemas)
			if err == nil {
				t.Fatalf("Should have seen an error")
			}
		})
	}
}
