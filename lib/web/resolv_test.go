package web_test

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/OA-PASS/metadata-schemas/lib/jsonschema"
	"github.com/OA-PASS/metadata-schemas/lib/web"
	"github.com/go-test/deep"
)

type testFetcher func(url *url.URL) (jsonschema.Instance, bool, error)

func (f testFetcher) GetSchema(url *url.URL) (jsonschema.Instance, bool, error) {
	if f != nil {
		return f(url)
	}
	return nil, false, nil
}

func TestSchemaRefResolve(t *testing.T) {
	url1, _ := url.Parse("http://example.org/one")
	url2, _ := url.Parse("http://example.org/two")

	m := map[string]jsonschema.Instance{
		url1.String(): {
			"foo": "bar",
		},
		url2.String(): {
			"baz": 8,
		},
	}

	ref := &web.SchemaRef{
		Schemas: []string{url1.String(), url2.String()},
	}

	instances, err := ref.Resolve(testFetcher(func(url *url.URL) (jsonschema.Instance, bool, error) {
		i, ok := m[url.String()]
		return i, ok, nil
	}))
	if err != nil {
		t.Fatalf("error resolving: %+v", err)
	}

	diff := deep.Equal(instances, []jsonschema.Instance{m[url1.String()], m[url2.String()]})
	if len(diff) > 0 {
		t.Fatalf("Schemas are nor equal: %+v", diff)
	}
}

func TestSchemaRefResolveErrors(t *testing.T) {
	cases := map[string]struct {
		urls []string
		f    testFetcher
	}{
		"badUrl": {
			urls: []string{"0http://what?"},
		},
		"noSchema": {
			urls: []string{"http://example.org/foo"},
		},
		"errSchema": {
			urls: []string{"http://example.org/foo"},
			f: testFetcher(func(*url.URL) (jsonschema.Instance, bool, error) {
				return nil, false, fmt.Errorf("an error")
			}),
		},
	}

	for name, c := range cases {
		c := c
		t.Run(name, func(t *testing.T) {
			ref := &web.SchemaRef{
				Schemas: c.urls,
			}

			_, err := ref.Resolve(c.f)
			if err == nil {
				t.Fatalf("Should have thrown error")
			}
		})
	}
}
