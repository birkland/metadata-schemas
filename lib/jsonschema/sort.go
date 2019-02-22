package jsonschema

import (
	"fmt"
	"sort"

	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonpointer"
)

const formPath = "/definitions/form/properties"

type analyzed struct {
	deps   map[string]bool
	id     string
	nForms int
	schema Instance
}

// Sorted produces a sorted slice of schemas, based on the following rules:
//
// If one schema is referenced by another in a $ref, then that schema appears before the other
//
// For schemas that are independent of one another, the one with the greatest number of form properties
// (/definitions/form/properties) appears before those that have fewer.
//
// If two schemas have no dependencies and have the same number of properties, the one that appears
// first in the initial list will be first in the result.
func Sorted(schemas []Instance) ([]Instance, error) {

	sorted, err := analyze(schemas)
	if err != nil {
		return nil, errors.Wrap(err, "error analyzing schemas")
	}

	// We provide a "greater than" function for the sort, since we want the "largest" (most depended upon)
	// schemas first
	sort.Slice(sorted, func(i, j int) bool {
		// If j is a dependency of i, then i is not "greater than" j
		if _, isDep := sorted[i].deps[sorted[j].id]; isDep {
			return false
		}

		// if i is a dependency of j, then I is "greater than" j
		if _, isDep := sorted[j].deps[sorted[i].id]; isDep {
			return true
		}

		// If neither have a dependency relationship, then i >= j if i has the same or more form elements
		return sorted[i].nForms >= sorted[j].nForms
	})

	result := make([]Instance, len(schemas))
	for i, v := range sorted {
		result[i] = v.schema
	}

	return result, nil
}

func analyze(schemas []Instance) (a []analyzed, err error) {
	a = make([]analyzed, len(schemas))
	for i, schema := range schemas {

		if schema == nil {
			return a, fmt.Errorf("nil schema encountered")
		}

		a[i].id = schema.ID()
		a[i].deps, err = findDeps(schema)
		if err != nil {
			return a, errors.Wrap(err, "could not analyze schemas")
		}
		a[i].nForms = countFormProperties(schema)
		a[i].schema = schema
	}

	return a, nil
}

// countFormProperties counts the number of /definitions/form/properties in a schema
func countFormProperties(schema Instance) int {
	pointer, _ := gojsonpointer.NewJsonPointer(formPath)
	val, _, err := pointer.Get(map[string]interface{}(schema))
	if err != nil {
		return 0
	}

	if properties, ok := val.(map[string]interface{}); ok {
		return len(properties)
	}

	return 0
}

// findDeps finds all the external/dependency schema URIs from a schema
func findDeps(schema Instance) (map[string]bool, error) {
	deps := make(map[string]bool)

	analyzer := schemaAnalyzer{doc: schema}

	err := analyzer.findRefs()
	if err != nil {
		return nil, errors.Wrap(err, "could not find $refs")
	}

	for _, ref := range analyzer.refs {
		if uri := ref.schemaURI(); uri != "" {
			deps[uri] = true
		}
	}

	return deps, nil
}
