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
