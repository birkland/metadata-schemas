package jsonschema_test

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/OA-PASS/metadata-schemas/lib/jsonschema"
	"github.com/go-test/deep"
)

func TestMapSchemaAddGet(t *testing.T) {
	id1 := "http://example.org/foo"
	id2 := "http://example.org/bar"

	data := []struct {
		id   string
		json string
	}{
		{id1, fmt.Sprintf(`{ "$id": "%s", "foo": "bar"}`, id1)},
		{id2, fmt.Sprintf(`{ "$id": "%s", "foo": "bar"}`, id2)},
	}

	m := jsonschema.Map(make(map[string]jsonschema.Instance))
	err := m.Add(strings.NewReader(data[0].json), strings.NewReader(data[1].json))
	if err != nil {
		t.Fatalf("Adding schema produced an error %+v", err)
	}

	if len(m) != 2 {
		t.Fatalf("Should have mapped two schemas")
	}

	for _, d := range data {
		_, ok := m[d.id]
		if !ok {
			t.Fatalf("Did not map id %s", d.id)
		}

		url, _ := url.Parse(d.id)

		roundtripped, ok, err := m.GetSchema(url)
		if !ok || err != nil {
			t.Errorf("Could not get schema %s", d.id)
		}

		original := jsonschema.Instance(make(map[string]interface{}))
		_ = json.Unmarshal([]byte(d.json), &original)

		diffs := deep.Equal(roundtripped, original)
		if len(diffs) > 0 {
			t.Fatalf("Differences found between roundtripped and original json: %s", diffs)
		}
	}

}

func TestMapSchemaErrors(t *testing.T) {
	cases := map[string]string{
		"oId":          `{"foo": "bar"}`,
		"idNotAString": `{"$id": {"foo": "bar"}}`,
		"badJSON":      "{{--,",
	}

	m := jsonschema.Map(make(map[string]jsonschema.Instance))

	for name, json := range cases {
		json := json
		t.Run(name, func(t *testing.T) {
			err := m.Add(strings.NewReader(json))
			if err == nil {
				t.Fatalf("Should have thrown an error")
			}
		})
	}
}

func TestDereference(t *testing.T) {

	expected := parseSchema(t, `{
		"$id": "http://example.org/cows/test.json",
		"foo": null,
		"cows": [
			"gladys",
			"gertrude",
			"bessie"
		],
		"stats": {
			"gladys": {
				"weight": 1024,
				"lactating": true
			},
			"bessie": {
				"stomach": "empty"
			}
		},
		"definitions": {
			"stomachStats": [
				"empty",
				"full"
			]
		}
	}`)

	toTest := parseSchema(t, `{
		"$id": "http://example.org/cows/test.json",
		"foo": null,
		"cows": [
			"gladys",
			"gertrude",
			{"$ref": "external.json#/definitions/cows/names/2"}
		],
		"stats": {
			"gladys": {"$ref": "http://example.org/cows/external.json#/definitions/stats/gladysStats"},
			"bessie": {
				"stomach": {"$ref": "#/definitions/stomachStats/0"}
			}
		},
		"definitions": {
			"stomachStats": [
				"empty",
				"full"
			]
		}
	}`)

	externals := jsonschema.Map(make(map[string]jsonschema.Instance))
	err := externals.Add(strings.NewReader(`{
		"$id": "http://example.org/cows/external.json",
		"definitions": {
			"cows": {
				"names": [
					"gladys",
					"gertrude",
					"bessie"
				]
			},
			"stats": {
				"gladysStats": {
					"weight": 1024,
					"lactating": true
				}
			}
		}
	}`))
	if err != nil {
		t.Fatalf("bad test json for external schema: %+v", err)
	}

	err = toTest.Dereference(externals)
	if err != nil {
		t.Fatalf("%+v", err)
	}

	diffs := deep.Equal(toTest, expected)
	if len(diffs) > 0 {
		t.Fatalf("Did not dereference to expected schema %s", diffs)
	}
}

func TestDereferenceParseErrors(t *testing.T) {
	cases := map[string]string{
		"noId":           `{"foo": {"$ref": "what.json/foo"}}`,
		"refIsNotString": `{"$id": "http://example.org/", "$ref": {"foo": "bar"}}`,
		"badJsonRef":     `{"$id": "http://example.org/", "$ref": "0http://foo/bar"}`,
		"badId":          `{"$id": "0http://example.org/foo", "foo": {"$ref": "what.json/foo"}}`,
		"nonStringID":    `{"$id": 42, "foo": {"$ref": "what.json/foo"}}`,
	}

	var m jsonschema.Map

	for name, body := range cases {
		body := body
		t.Run(name, func(t *testing.T) {
			s := parseSchema(t, body)
			err := s.Dereference(m)
			if err == nil {
				t.Fatalf("Should have thrown an error")
			}
			if len(err.Error()) == 0 {
				t.Fatalf("Should have printed an error")
			}
		})
	}
}

type testFetcher func(url *url.URL) (jsonschema.Instance, bool, error)

func (f testFetcher) GetSchema(url *url.URL) (jsonschema.Instance, bool, error) {
	return f(url)
}
func TestDereferenceResolveErrors(t *testing.T) {
	toTest := parseSchema(t, `{
		"$id": "http://example.org/cows/test.json",
		"foo": {"$ref": "foo.json/there"}
	}`)

	cases := map[string]jsonschema.Fetcher{
		"nil": nil,
		"notFound": testFetcher(func(uri *url.URL) (jsonschema.Instance, bool, error) {
			return nil, false, nil
		}),
		"lookupError": testFetcher(func(uri *url.URL) (jsonschema.Instance, bool, error) {
			return nil, false, fmt.Errorf("This is an error")
		}),
	}

	for name, fetcher := range cases {
		fetcher := fetcher
		t.Run(name, func(t *testing.T) {
			err := toTest.Dereference(fetcher)
			if err == nil {
				t.Fatalf("Should have thrown an error!")
			}
		})
	}
}

func TestDereferencePointerError(t *testing.T) {
	bad := parseSchema(t, `{
		"$id": "http://example.org/cows/test.json",
		"foo": {"$ref": "#/does/not/exist"}
	}`)

	good := parseSchema(t, `{
		"$id": "http://example.org/cows/test.json",
		"foo": {"$ref": "#/foo"},
		"bar": "baz"
	}`)

	err := good.Dereference(nil)
	if err != nil {
		t.Fatalf("Got unexpected error %+v", err)
	}

	err = bad.Dereference(nil)
	if err == nil {
		t.Fatalf("Should have seen an error")
	}
}

func parseSchema(t *testing.T, jsonString string) jsonschema.Instance {
	instance := jsonschema.Instance(make(map[string]interface{}))
	err := json.Unmarshal([]byte(jsonString), &instance)
	if err != nil {
		t.Fatalf("could not parse json: %+v", err)
	}

	return instance
}
