package web_test

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/OA-PASS/metadata-schemas/lib/jsonschema"
	"github.com/OA-PASS/metadata-schemas/lib/web"
	"github.com/go-test/deep"
)

type object map[string]interface{}

func TestSchemaService(t *testing.T) {
	resourceurl1 := "http://ecample.org/repo/1"
	resourceurl2 := "http://ecample.org/repo/2"
	schemaurl1 := "http://example.org/one"
	schemaurl2 := "http://example.org/two"

	schemaMap := map[string]jsonschema.Instance{
		schemaurl1: {
			"foo": "bar",
		},
		schemaurl2: {
			"definitions": object{
				"form": object{
					"properties": object{
						"foo": "bar",
					},
				},
			},
			"baz": 8,
		},
	}

	req := &web.Request{
		Resources: []string{resourceurl1, resourceurl2},
	}

	schemaService := &web.SchemaService{
		PassClient: &staticClient{
			resultsJSON: map[string]string{
				resourceurl1: fmt.Sprintf(`{
					"schemas": [
						"%s",
						"%s"
					]
				}`, schemaurl1, schemaurl2),
				resourceurl2: fmt.Sprintf(`{
					"schema": ["%s"]
				}`, schemaurl1),
			},
		},
		SchemaFetcher: testFetcher(func(url *url.URL) (jsonschema.Instance, bool, error) {
			i, ok := schemaMap[url.String()]
			return i, ok, nil
		}),
	}

	instances, err := schemaService.Schemas(req)
	if err != nil {
		t.Fatalf("encountered an error %+v", err)
	}

	expected := []jsonschema.Instance{
		schemaMap[schemaurl2],
		schemaMap[schemaurl1]}

	diffs := deep.Equal(instances, expected)
	if len(diffs) > 0 {
		t.Fatalf("did not get expected schema instances %+v", diffs)
	}
}

func TestSchemaServiceErrors(t *testing.T) {
	resourceURL := "http://example.org/resource"
	schemaURL := "http://example.org/schema"

	cases := map[string]web.SchemaService{
		"passError": {
			PassClient: &staticClient{},
		},
		"schemaError": {
			PassClient: &staticClient{
				resultsJSON: map[string]string{
					resourceURL: fmt.Sprintf(`{
					"schemas": ["%s"]
				}`, schemaURL),
				}},
			SchemaFetcher: testFetcher(func(url *url.URL) (jsonschema.Instance, bool, error) {
				return nil, false, fmt.Errorf("This is an error")
			}),
		},
	}

	req := &web.Request{
		Resources: []string{resourceURL},
	}

	for name, svc := range cases {
		svc := svc
		t.Run(name, func(t *testing.T) {
			_, err := svc.Schemas(req)
			if err == nil {
				t.Fatalf("Should have thrown an error")
			}
		})
	}
}
