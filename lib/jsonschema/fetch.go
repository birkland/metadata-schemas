package jsonschema

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

// Fetcher fetches a parsed schema, given a URL.
// The URL may be unclean (have hash fragments)
type Fetcher interface {
	GetSchema(uri *url.URL) (s Instance, ok bool, err error) // Retrieve a schema
}

// Map is a simple map of schema URIs to schema instances, serving as a static
// schema Fetcher
type Map map[string]Instance

// Add an un-parsed schema to the map.  This will parse it, and place it
// into the map based on the $id present in the schema
func (m Map) Add(src ...io.Reader) error {
	for _, reader := range src {
		schema := make(map[string]interface{})
		err := json.NewDecoder(reader).Decode(&schema)
		if err != nil {
			return errors.Wrapf(err, "could not decode schema instance")
		}

		i, ok := schema[idKey]
		if !ok {
			return fmt.Errorf("schema does not have an $id")
		}

		id, ok := i.(string)
		if !ok {
			return fmt.Errorf("$id is not a string")
		}

		log.Printf("Loaded schema %s", id)
		m[id] = schema
	}

	return nil
}

// GetSchema retrieves a schema from the map, given a possibly unclean URL.
func (m Map) GetSchema(url *url.URL) (Instance, bool, error) {
	normalized := strings.Split(url.String(), "#")[0]
	s, ok := m[normalized]
	return s, ok, nil
}
