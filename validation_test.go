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

const schemaDir = "schemas/jhu"

func TestSchemaValidity(t *testing.T) {
	schemaMap := loadSchemas(t)

	cases := map[string]bool{
		"examples/jhu/full.json":                         true,
		"testdata/valid/nlmta-but-no-pubtype.json":       true,
		"testdata/valid/no-nlmta-but-has-a-pubtype.json": true,
		"testdata/invalid/no-nlmta-or-pubtype.json":      false,
	}

	for filename, shouldBeValid := range cases {
		filename := filename
		shouldBeValid := shouldBeValid
		t.Run(filename, func(t *testing.T) {
			var hasExpectedFailure bool
			for id, schema := range schemaMap {
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

func loadSchemas(t *testing.T) jsonschema.Map {

	dir, err := ioutil.ReadDir(schemaDir)
	if err != nil {
		t.Fatalf("Could not list schema directory!")
	}

	var schemaFiles []string
	for _, dirent := range dir {
		if dirent.Mode().IsRegular() && strings.HasSuffix(dirent.Name(), ".json") {
			schemaFiles = append(schemaFiles, filepath.Join(schemaDir, dirent.Name()))
		}
	}

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

	// Finally, add a union schema that merges them all (making sure that validations are expected against that too)
	merged, err := jsonschema.Merge(schemas)
	if err != nil {
		t.Fatalf("Error merging schemas: %+v", err)
	}
	schemaMap["merged"] = merged

	return schemaMap
}
