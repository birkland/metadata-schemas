// +build integration

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/OA-PASS/metadata-schemas/lib/jsonschema"
)

const j10pPath = "repositories/j10p"
const nihmsPath = "repositories/nihms"
const defaultFedoraBaseuri = "http://localhost:8080/fcrepo/rest"

func TestFedoraIntegration(t *testing.T) {
	client := &http.Client{}

	setupFedora(t, client)

	schemas := invokeSchemaService(t, client, false, fedoraURI(nihmsPath), fedoraURI(j10pPath))

	//verify we have three schemas returned (common, j10p, and nihms)
	if len(schemas) != 3 {
		t.Fatalf("Wrong number of schemas, got %d", len(schemas))
	}

	// Finally, verify the ordering of results
	expectedSchemas := []string{
		"https://oa-pass.github.io/metadata-schemas/jhu/common.json",
		"https://oa-pass.github.io/metadata-schemas/jhu/nihms.json",
		"https://oa-pass.github.io/metadata-schemas/jhu/jscholarship.json",
	}

	for i, schema := range schemas {
		if schema.ID() != expectedSchemas[i] {
			t.Fatalf("Saw %s as schema %d, but should have seen %s", schema.ID(), i, expectedSchemas[i])
		}
	}
}

func TestMergeSchemas(t *testing.T) {
	client := &http.Client{}

	setupFedora(t, client)

	schemas := invokeSchemaService(t, client, true, fedoraURI(nihmsPath), fedoraURI(j10pPath))

	//verify we have only one schema
	if len(schemas) != 1 {
		t.Fatalf("Wrong number of schemas, got %d", len(schemas))
	}
}

func invokeSchemaService(t *testing.T, client *http.Client, merge bool, repos ...string) []jsonschema.Instance {
	post, _ := http.NewRequest(http.MethodPost, schemaServiceURI(merge), requestBody(t, repos...))
	post.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(post)
	if err != nil {
		t.Fatalf("POST request failed: %s", err)
	}

	if resp.StatusCode > 299 {
		t.Fatalf("Schema service returned an error: %d", resp.StatusCode)
	}

	// Read in the body
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	// Make sure the body does not have any $ref
	if strings.Contains(string(body), "$ref") {
		t.Fatalf("Body contains a non-dereferenced $ref")
	}

	// Unmarshal into JSON and verify we have three schemas returned (common, j10p, and nihms)
	var schemas []jsonschema.Instance
	_ = json.Unmarshal(body, &schemas)
	return schemas
}

func requestBody(t *testing.T, uris ...string) io.Reader {
	body, err := json.Marshal(uris)
	if err != nil {
		t.Fatalf("Did not marshal JSON: %s", err) // should never happen
	}
	return bytes.NewReader(body)
}

func authz(r *http.Request) {
	username, ok := os.LookupEnv("PASS_FEDORA_USER")
	if !ok {
		username = "fedoraAdmin"
	}

	passwd, ok := os.LookupEnv("PASS_FEDORA_PASSWORD")
	if !ok {
		passwd = "moo"
	}

	r.SetBasicAuth(username, passwd)
}

func schemaServiceURI(doMerge bool) string {
	port, ok := os.LookupEnv("SCHEMA_SERVICE_PORT")
	if !ok {
		port = "8086"
	}

	host, ok := os.LookupEnv("SCHEMA_SERVICE_HOST")
	if !ok {
		host = "localhost"
	}

	if !doMerge {
		return fmt.Sprintf("http://%s:%s", host, port)
	}

	return fmt.Sprintf("http://%s:%s?merge=true", host, port)
}

func fedoraURI(uripath string) string {
	return fmt.Sprintf("%s/%s", fedoraBaseURI(), uripath)
}

func fedoraBaseURI() string {
	baseuri, ok := os.LookupEnv("PASS_EXTERNAL_FEDORA_BASEURL")
	if !ok {
		return defaultFedoraBaseuri
	}

	return strings.Trim(baseuri, "/")
}

func setupFedora(t *testing.T, c *http.Client) {
	j10p := fmt.Sprintf(`{
		"@context" : "https://oa-pass.github.io/pass-data-model/src/main/resources/context-3.3.jsonld",
		"@id" : "%s",
		"@type" : "Repository",
		"agreementText" : "I agree",
		"schemas": [
			"https://oa-pass.github.io/metadata-schemas/jhu/jscholarship.json",
			"https://oa-pass.github.io/metadata-schemas/jhu/common.json"
		],
		"integrationType" : "full",
		"name" : "JScholarship",
		"repositoryKey" : "jscholarship",
		"url" : "https://jscholarship.library.jhu.edu/"
	  }
	  `, fedoraURI(j10pPath))
	nihms := fmt.Sprintf(`{
		"@context" : "https://oa-pass.github.io/pass-data-model/src/main/resources/context-3.3.jsonld",
		"@id" : "%s",
		"@type" : "Repository",
		"schemas": [
			"https://oa-pass.github.io/metadata-schemas/jhu/nihms.json",
			"https://oa-pass.github.io/metadata-schemas/jhu/common.json"
		],
		"integrationType" : "one-way",
		"name" : "PubMed Central",
		"repositoryKey" : "pmc",
		"url" : "https://www.ncbi.nlm.nih.gov/pmc/"
	  }
	  `, fedoraURI(nihmsPath))

	putResource(t, c, fedoraURI(j10pPath), j10p)
	putResource(t, c, fedoraURI(nihmsPath), nihms)
}

func putResource(t *testing.T, c *http.Client, uri string, body string) {
	request, err := http.NewRequest(http.MethodPut, uri, strings.NewReader(body))
	if err != nil {
		t.Fatalf("Building request failed: %s", err)
	}

	request.Header.Set("Content-Type", "application/ld+json")
	request.Header.Set("Prefer", `handling=lenient; received="minimal"`)
	authz(request)

	resp, err := c.Do(request)
	if err != nil {
		t.Fatalf("PUT request failed: %s", err)
	}

	defer resp.Body.Close()
	io.Copy(ioutil.Discard, resp.Body)

	if resp.StatusCode > 299 {
		t.Fatalf("Could not add resource: %d", resp.StatusCode)
	}
}
