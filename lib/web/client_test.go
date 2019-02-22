package web_test

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/OA-PASS/metadata-schemas/lib/web"
	"github.com/go-test/deep"
)

type fakeRequester struct {
	f func(req *http.Request) (*http.Response, error)
}

func (r *fakeRequester) Do(req *http.Request) (*http.Response, error) {
	if r.f != nil {
		return r.f(req)
	}

	return nil, nil
}

type fakeBody struct {
	io.Reader
	closeFunc func() error
}

func (b *fakeBody) Close() error {
	if b.closeFunc != nil {
		return b.closeFunc()
	}
	return nil
}

func TestPrivateFetchEntity(t *testing.T) {
	publicBaseURI := "https://example.org/foo/fcrepo/rest"
	privateBaseURI := "http://127.0.0.1:8080/rest"

	privateResource := privateBaseURI + "/foo/bar"
	publicResource := publicBaseURI + "/foo/bar"

	username := "foo"
	password := "bar"

	schemaURIs := []string{"http://example.org/foo", "http://example.org/bar"}

	client := web.InternalPassClient{
		Requester: &fakeRequester{
			f: func(req *http.Request) (*http.Response, error) {

				if req.URL.String() != privateResource {
					t.Fatalf("private resource URI is incorrect")
				}

				user, pass, ok := req.BasicAuth()
				if !ok || user != username || pass != password {
					t.Fatalf("basic auth is wrong")
				}

				return &http.Response{
					Body: &fakeBody{
						Reader: strings.NewReader(fmt.Sprintf(`{
							"schemas" : [
								"%s",
								"%s"
							]
						}`, schemaURIs[0], schemaURIs[1])),
					},
				}, nil
			},
		},
		InternalBaseURI: privateBaseURI,
		ExternalBaseURI: publicBaseURI,
		Credentials: &web.Credentials{
			Username: username,
			Password: password,
		},
	}

	var ref web.SchemaRef

	err := client.FetchEntity(publicResource, &ref)
	if err != nil {
		t.Fatalf("Client fetch resulted in error %+v", err)
	}

	diffs := deep.Equal(schemaURIs, ref.Schemas)
	if len(diffs) > 0 {
		t.Fatalf("found difference in deserialized content %s", diffs)
	}

	err = client.FetchEntity("http://example.org/bad/resource", &ref)
	if err == nil {
		t.Fatalf("Should have thrown error on non whitelisted uri")
	}
}

func TestPrivateFetchEntityErrors(t *testing.T) {
	cases := []struct {
		name string
		url  string
		f    func(req *http.Request) (*http.Response, error)
	}{
		{
			name: "badURI",
			url:  "0http://bad",
		},
		{
			name: "httpError",
			url:  "http://example.org/foo",
			f: func(req *http.Request) (*http.Response, error) {
				return nil, fmt.Errorf("this is an error")
			},
		},
		{
			name: "badJSON",
			url:  "http://example.org/foo",
			f: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					Body: &fakeBody{
						Reader: strings.NewReader(`{BAD JSON-,`),
					},
				}, nil
			},
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			client := web.InternalPassClient{
				Requester: &fakeRequester{
					f: c.f,
				},
			}

			var ref web.SchemaRef

			err := client.FetchEntity(c.url, &ref)
			if err == nil {
				t.Fatalf("Should have terminated with an error")
			}
		})
	}
}
