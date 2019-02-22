package web

import (
	"github.com/OA-PASS/metadata-schemas/lib/jsonschema"
	"github.com/pkg/errors"
)

type SchemaService struct {
	PassClient    PassEntityFetcher
	SchemaFetcher jsonschema.Fetcher
}

func (s *SchemaService) Schemas(r *Request) ([]jsonschema.Instance, error) {
	var instances []jsonschema.Instance
	fetched := make(map[string]bool)

	for _, uri := range r.Resources {
		var ref SchemaRef

		err := s.PassClient.FetchEntity(uri, &ref)
		if err != nil {
			return nil, errors.Wrapf(err, "could not fetch %s", uri)
		}

		schemas, err := ref.Resolve(s.SchemaFetcher)
		if err != nil {
			return nil, errors.Wrapf(err, "could not resolve schema")
		}

		for i, schema := range schemas {
			uri := ref.Schemas[i]
			if _, exists := fetched[uri]; !exists {
				fetched[ref.Schemas[i]] = true
				instances = append(instances, schema)
			}
		}
	}

	sorted, err := jsonschema.Sorted(instances)
	if err != nil {
		return nil, errors.Wrapf(err, "could not sort schemas")
	}
	for _, schema := range sorted {
		err := schema.Dereference(s.SchemaFetcher)
		if err != nil {
			return nil, errors.Wrapf(err, "could not dereference schema %s", schema.ID())
		}
	}

	return sorted, nil
}
