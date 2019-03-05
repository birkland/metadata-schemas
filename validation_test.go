package metadata_schemas

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/OA-PASS/metadata-schemas/lib/jsonschema"
	"github.com/xeipuuv/gojsonschema"
)

const schemaDir = "jhu"
const exampleData = "examples/jhu/full.json"

func TestSchemaValidity(t *testing.T) {
	schemaMap := loadSchemas(t)

	// Load the test data
	testdataFile, err := os.Open(exampleData)
	if err != nil {
		t.Fatalf("Could not open example data")
	}
	defer testdataFile.Close()
	body, _ := ioutil.ReadAll(testdataFile)
	testData := gojsonschema.NewBytesLoader(body)

	for id, schema := range schemaMap {
		toTest := gojsonschema.NewGoLoader(schema)

		result, err := gojsonschema.Validate(toTest, testData)
		if err != nil {
			t.Fatalf("Error validating against schema %s: %+v", id, err)
		}

		if !result.Valid() {
			for _, err := range result.Errors() {
				t.Logf("- %s\n", err)
			}
			t.Fatalf("Schema validation for %s failed!", id)
		}
	}
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

	return schemaMap
}
