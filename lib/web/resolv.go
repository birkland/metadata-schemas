package web

import (
	"fmt"
	"net/url"

	"github.com/OA-PASS/metadata-schemas/lib/jsonschema"
	"github.com/pkg/errors"
)

// HTTP headers
const (
	headerAccept    = "Accept"
	headerUserAgent = "User-Agent"
)

const (
	mediaJSONTypes = "application/json, application/ld+json"
)

// PassEntityFetcher retrieves the JSON-LD content at the given url, and
// unmarshals it into the provided struct.
//
// entityPointer is expected to be a pointer to a struct or map
// (i.e. anything encoding/json can unmarshal into).  An error will be returned otherwise.
type PassEntityFetcher interface {
	FetchEntity(url string, entityPointer interface{}) error
}

// SchemaRef is a "cut down" pass entity containing only
// an array of schema URIs.  It is a subset of the pass Repository entity.
type SchemaRef struct {
	Schemas []string `json:"schemas"`
}

// Resolve fetches and parses all schemas referenced by this SchemaRef
func (r *SchemaRef) Resolve(resolver jsonschema.Fetcher) ([]jsonschema.Instance, error) {
	var instances []jsonschema.Instance

	for _, addr := range r.Schemas {

		uri, err := url.Parse(addr)
		if err != nil {
			return instances, errors.Wrapf(err, "could not parse schema url %s", addr)
		}

		schema, ok, err := resolver.GetSchema(uri)
		if err != nil {
			return instances, errors.Wrapf(err, "could not fetch schema at %s", uri.String())
		}
		if !ok {
			return instances, fmt.Errorf("no schema at %s", uri.String())
		}

		instances = append(instances, schema)
	}

	return instances, nil
}
