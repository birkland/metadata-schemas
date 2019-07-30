package metadata_schemas_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/OA-PASS/metadata-schemas/lib/jsonschema"
	"github.com/xeipuuv/gojsonschema"
)

const (
	jhuSchemaDir     = "schemas/jhu"
	harvardSchemaDir = "schemas/harvard"
)

// Merged schemas between institutions may conflict, so we need to do them separately
func TestJHUSchemas(t *testing.T) {
	jhuSchemas := loadSchemas(t, jhuSchemaDir, true)

	cases := map[string]bool{}

	addCases(t, cases, "examples/jhu", true)
	addCases(t, cases, "testdata/valid/jhu", true)
	addCases(t, cases, "testdata/invalid/jhu", false)

	validateSchemas(t, jhuSchemas, cases)
}

func TestHarvardSchemas(t *testing.T) {
	harvardSchemas := loadSchemas(t, harvardSchemaDir, true)

	cases := map[string]bool{}

	addCases(t, cases, "examples/harvard", true)
	addCases(t, cases, "testdata/valid/harvard", true)
	addCases(t, cases, "testdata/invalid/harvard", false)

	validateSchemas(t, harvardSchemas, cases)
}

func validateSchemas(t *testing.T, schemas jsonschema.Map, cases map[string]bool) {
	for filename, shouldBeValid := range cases {
		filename := filename
		shouldBeValid := shouldBeValid
		t.Run(filename, func(t *testing.T) {
			var hasExpectedFailure bool
			for id, schema := range schemas {
				toTest := gojsonschema.NewGoLoader(schema)

				result, err := gojsonschema.Validate(toTest, loadTestSchema(t, filename))
				if err != nil {
					t.Fatalf("Error validating against schema %s: %+v", id, err)
				}

				if shouldBeValid && !result.Valid() {
					for _, err := range result.Errors() {
						t.Logf("- %s\n", err)
					}
					t.Fatalf("Schema validation for %s failed!", id)
				} else if !shouldBeValid && !result.Valid() {
					hasExpectedFailure = true
				}
			}

			if !shouldBeValid && !hasExpectedFailure {
				t.Fatalf("schema passed validation, but should have failed")
			}
		})
	}
}

func loadTestSchema(t *testing.T, filename string) gojsonschema.JSONLoader {
	testdataFile, err := os.Open(filename)
	if err != nil {
		t.Fatalf("Could not open example data")
	}
	defer testdataFile.Close()
	body, _ := ioutil.ReadAll(testdataFile)
	return gojsonschema.NewBytesLoader(body)
}

func loadSchemas(t *testing.T, dir string, merge bool) jsonschema.Map {

	schemaFiles := findJSONDocs(t, dir)

	schemaMap, err := jsonschema.Load(schemaFiles)
	if err != nil {
		t.Fatalf("Error loading schemas: %+v", err)
	}

	var schemas []jsonschema.Instance
	for _, v := range schemaMap {
		schemas = append(schemas, v)
	}

	err = jsonschema.Dereference(schemaMap, schemas...)
	if err != nil {
		t.Fatalf("Error dereferencing schemas %+v", err)
	}

	if merge {
		// Finally, add a union schema that merges them all (making sure that validations are expected against that too)
		merged, err := jsonschema.Merge(schemas)
		if err != nil {
			t.Fatalf("Error merging schemas: %+v", err)
		}
		schemaMap["merged"] = merged
	}

	return schemaMap
}

func addCases(t *testing.T, cases map[string]bool, fromDir string, shouldValidate bool) {
	for _, path := range findJSONDocs(t, fromDir) {
		cases[path] = shouldValidate
	}
}

func findJSONDocs(t *testing.T, dir string) []string {
	var paths []string
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, "json") {
			paths = append(paths, path)
		}
		return nil
	})

	if err != nil {
		t.Fatalf("Walk for schemas failed: %+v", err)
	}

	return paths
}
