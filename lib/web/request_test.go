package web_test

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/OA-PASS/metadata-schemas/lib/web"
	"github.com/go-test/deep"
)

type errReader struct{}

func (e *errReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("Error reading")
}

func TestReadJSON(t *testing.T) {
	url1 := "http://example.org/foo"
	url2 := "http://example.org/bar"
	json := fmt.Sprintf(`["%s", "%s"]`, url1, url2)

	var request web.Request

	err := request.ReadJSON(strings.NewReader(json))
	if err != nil {
		t.Fatalf("Could not read JSON %+v", err)
	}

	diffs := deep.Equal(request.Resources, []string{url1, url2})
	if len(diffs) > 0 {
		t.Fatalf("encountered difference in URLs %+v", diffs)
	}
}

func TestReadJSONErrors(t *testing.T) {

	cases := map[string]io.Reader{
		"badJSON": strings.NewReader("{s-d;"),
		"badURIs": strings.NewReader(`["0http://example.org"]`),
	}

	var request web.Request

	for name, reader := range cases {
		reader := reader
		t.Run(name, func(t *testing.T) {
			err := request.ReadJSON(reader)
			if err == nil {
				t.Fatalf("Should not have seen an error")
			}
		})
	}
}

func TestReadText(t *testing.T) {
	url1 := "http://example.org/foo"
	url2 := "http://example.org/bar"
	json := fmt.Sprintf(`
	%s

	    %s
	`, url1, url2)

	var request web.Request

	err := request.ReadText(strings.NewReader(json))
	if err != nil {
		t.Fatalf("Could not read text\n %+v", err)
	}

	diffs := deep.Equal(request.Resources, []string{url1, url2})
	if len(diffs) > 0 {
		t.Fatalf("encountered difference in URLs %+v", diffs)
	}
}

func TestReadTextErrors(t *testing.T) {

	cases := map[string]io.Reader{
		"badReader": &errReader{},
		"badURIs":   strings.NewReader(`0http://example.org`),
	}

	var request web.Request

	for name, reader := range cases {
		reader := reader
		t.Run(name, func(t *testing.T) {
			err := request.ReadText(reader)
			if err == nil {
				t.Fatalf("Should not have seen an error")
			}
		})
	}
}

type staticClient struct {
	resultsJSON map[string]string
}

func (c *staticClient) FetchEntity(url string, entityPointer interface{}) error {
	v, ok := c.resultsJSON[url]
	if !ok {
		return fmt.Errorf("no value for %s", url)
	}

	return json.Unmarshal([]byte(v), entityPointer)
}
