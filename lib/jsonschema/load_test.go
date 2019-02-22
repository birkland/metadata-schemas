package jsonschema_test

import (
	"testing"

	"github.com/OA-PASS/metadata-schemas/lib/jsonschema"
)

func TestLoad(t *testing.T) {
	m, err := jsonschema.Load([]string{"testdata/good/schema1.json", "testdata/good/dir"})
	if err != nil {
		t.Fatalf("got error %+v", err)
	}

	if len(m) != 2 {
		t.Fatalf("Did not find the right number of schemas")
	}
}

func TestLoadErrors(t *testing.T) {
	cases := map[string]string{
		"badSchema":    "testdata/bad.json",
		"doesNotExist": "DOES_NOT_EXIST",
	}

	for name, path := range cases {
		path := path
		t.Run(name, func(t *testing.T) {
			_, err := jsonschema.Load([]string{path})
			if err == nil {
				t.Fatalf("shoud have thrown an error")
			}
		})
	}
}
