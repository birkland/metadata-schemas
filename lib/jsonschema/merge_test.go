package jsonschema_test

import (
	"strings"
	"testing"

	"github.com/OA-PASS/metadata-schemas/lib/jsonschema"
	"github.com/go-test/deep"
)

func TestMerge(t *testing.T) {
	cases := []struct {
		name, expected string
		schemas        []string
	}{{
		name: "simple ignore preamble",
		schemas: []string{`{
			"$schema": "http://example.org/schema",
			"$id": "http://example.org/foo",
			"title": "foo",
			"description": "foo schema",
			"$comment": "one",
			"a": "1"
		}`, `{
			"$schema": "http://example.org/schema",
			"$id": "http://example.org/bar",
			"title": "bar",
			"description": "bar schema",
			"$comment": "two",
			"b": "2"
		}`},
		expected: `{
			"a": "1",
			"b": "2"
		}`,
	}, {
		name: "ignorable conflicts",
		schemas: []string{`{
			"a": {
				"title": "A",
				"description": "a letter",
				"$comment": "displays good",
				"type": "letter"
			}
		}`, `{
			"a": {
				"title": "a",
				"description": "an awesome letter",
				"$comment": "displays nicely",
				"type": "letter"
			}
		}`},
		expected: `{
			"a": {
				"title": "a",
				"$comment": "displays nicely",
				"description": "an awesome letter",
				"type": "letter"
			}
		}`,
	}, {
		name: "simple array deduplication",
		schemas: []string{`{
			"array": ["a", "b", "c"]
		}`, `{
			"array": ["b", "c", "d"]
		}`, `{
			"array": ["c", "d", "e"]
		}`},
		expected: `{
			"array": ["a", "b", "c", "d", "e"]
		}`,
	}, {
		name: "complex array deduplication",
		schemas: []string{`{
			"array": [{"a": ["b", {"c": "d"}]}, "e"]
		}`, `{
			"array": [{"a": ["b", {"c": "d"}]}, "f"]
		}`, `{
			"array": ["e", "f", {"g": "h"}]
		}`},
		expected: `{
			"array": [{"a": ["b", {"c": "d"}]}, "e", "f", {"g": "h"}]
		}`,
	}, {
		name: "object merge",
		schemas: []string{`{
			"a": "b",
			"c": ["d", "e"]
		}`, `{
			"a": "b",
			"c": ["e", "f", "g"]
		}`, `{
			"h": {
				"i": "j",
				"k": ["l", "m"],
				"n": {
					"o": "p"
				}
			}
		}`, `{
			"h": {
				"k": ["l", "m", "m'"],
				"n": {
					"q": "r"
				}
			}
		}`},
		expected: `{
			"a": "b",
			"c": ["d", "e", "f", "g"],
			"h": {
				"i": "j",
				"k": ["l", "m", "m'"],
				"n": {
					"o": "p",
					"q": "r"
				}
			}
		}`,
	}}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			var schemas []jsonschema.Instance
			for _, s := range c.schemas {
				schemas = append(schemas, parseSchema(t, s))
			}

			merged, err := jsonschema.Merge(schemas)
			if err != nil {
				t.Fatalf("Merging schemas failed: %+v", err)
			}

			if diffs := deep.Equal(merged, parseSchema(t, c.expected)); len(diffs) > 0 {
				t.Fatalf("Result does not match expected: %s", strings.Join(diffs, "\n"))
			}
		})
	}
}

func TestMergeErrors(t *testing.T) {
	cases := map[string][]string{

		"value conflict": {`{
			"a": "b"
		}`, `{
			"a": "c"
		}`},

		"nested value conflict": {`{
			"conflicting": {
				"a": "b"
			}
		}`, `{
			"conflicting": {
				"a": "c"
			}
		}`},

		"type conflict": {`{
			"a": ["b", "c"]
		}`, `{
			"a": "e"
		}`},
	}

	for name, schemas := range cases {
		schemas := schemas
		t.Run(name, func(t *testing.T) {
			var toMerge []jsonschema.Instance
			for _, s := range schemas {
				toMerge = append(toMerge, parseSchema(t, s))
			}

			_, err := jsonschema.Merge(toMerge)
			if err == nil {
				t.Fatal("expected to see an error!")
			}
		})
	}
}
