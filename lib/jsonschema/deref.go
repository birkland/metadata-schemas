package jsonschema

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonpointer"
	"github.com/xeipuuv/gojsonreference"
)

const refKey = "$ref"

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

// Dereference replaces all $ref references with their dereferenced content
func (s Instance) Dereference(pool Fetcher) error {

	proc := schemaAnalyzer{
		doc: s,
	}

	err := proc.findRefs()
	if err != nil {
		return errors.Wrapf(err, "could not find references in schema")
	}

	var self map[string]interface{} = s
	for _, ref := range proc.refs {
		var src map[string]interface{}
		if ref.what.HasFragmentOnly {
			src = s
		} else if pool != nil {
			var ok bool
			src, ok, err = pool.GetSchema(ref.what.GetUrl())
			if err != nil {
				return errors.Wrapf(err, "error fetching schema %s", ref.what.GetUrl())
			}
			if !ok {
				return fmt.Errorf("no schema found %s", ref.what.GetUrl())
			}
		} else {
			return fmt.Errorf("external reference found, but no schema fetcher provided: %s", ref.what.GetUrl())
		}

		// Extract the value from the source (this doc, or an external one)
		value, _, err := ref.what.GetPointer().Get(src)
		if err != nil {
			return errors.Wrapf(err, "could not read value at %s", ref.what.GetUrl())
		}

		// Now set the value in this doc
		_, err = ref.where.Set(self, value)
		if err != nil {
			return errors.Wrapf(err, "could not insert content from %s at %s", ref.what.GetUrl(), ref.where)
		}
	}

	return nil
}

type errs []error

func (e errs) Error() string {
	return fmt.Sprintf("%s", []error(e))
}

type jsonPointerPath []string

func (p jsonPointerPath) next(segment string) jsonPointerPath {
	return append(p, segment)
}

func (p jsonPointerPath) pointer() gojsonpointer.JsonPointer {
	ptr, _ := gojsonpointer.NewJsonPointer(fmt.Sprintf("/%s", strings.Join(p, "/")))
	return ptr
}

type ref struct {
	what  gojsonreference.JsonReference
	where gojsonpointer.JsonPointer
}

func (r *ref) schemaURI() string {
	if r.what.HasFullUrl {
		return strings.Split(r.what.GetUrl().String(), "#")[0]
	}
	return ""
}

type schemaAnalyzer struct {
	doc    Instance
	refs   []ref
	errors errs
}

func (c *schemaAnalyzer) findRefs() error {
	c.scanObj(nil, c.doc)

	if len(c.errors) > 0 {
		return c.errors
	}

	return nil
}

func (c *schemaAnalyzer) scanObj(p jsonPointerPath, obj map[string]interface{}) {
	for k, v := range obj {
		if k == refKey {
			c.addRef(p, v)
		} else {
			c.scan(p.next(k), v)
		}
	}
}

func (c *schemaAnalyzer) scan(p jsonPointerPath, something interface{}) {
	if something == nil {
		return
	}

	switch v := something.(type) {
	case map[string]interface{}:
		c.scanObj(p, v)
	case []interface{}:
		c.scanList(p, v)
	default:
	}
}

func (c *schemaAnalyzer) scanList(p jsonPointerPath, list []interface{}) {
	for i, v := range list {
		c.scan(p.next(strconv.Itoa(i)), v)
	}
}

func (c *schemaAnalyzer) addRef(p jsonPointerPath, refContent interface{}) {

	refString, ok := refContent.(string)
	if !ok {
		c.errors = append(c.errors, fmt.Errorf("expected to find a string at %s, instead found %s", p.pointer(), refContent))
		return
	}

	jsonref, err := gojsonreference.NewJsonReference(refString)
	if err != nil {
		c.errors = append(c.errors,
			errors.Wrapf(err, "could not parse json reference at %s with content %s", p.pointer(), refString))
		return
	}

	// Ref is of the form foo.json#/path.  So we need to make foo.json absolute by taking the base
	// of $id
	if !jsonref.HasFullUrl && jsonref.HasUrlPathOnly && !jsonref.HasFragmentOnly {
		id, ok := c.doc["$id"]
		if !ok {
			c.errors = append(c.errors, fmt.Errorf("found relative reference %s, but doc does not have an ID", jsonref.GetUrl()))
			return
		}

		if _, ok := id.(string); !ok {
			c.errors = append(c.errors, fmt.Errorf("$id is not a string"))
			return
		}

		uri, err := url.Parse(id.(string))
		if err != nil {
			c.errors = append(c.errors, errors.Wrap(err, "$id is not a well-formed url"))
			return
		}
		uri.Path = path.Join(path.Dir(uri.Path), jsonref.GetUrl().Path)

		absoluteRef, _ := gojsonreference.NewJsonReference(fmt.Sprintf("%s#%s", uri.String(), jsonref.GetUrl().Fragment))

		c.refs = append(c.refs, ref{what: absoluteRef, where: p.pointer()})
		return
	}

	c.refs = append(c.refs, ref{what: jsonref, where: p.pointer()})
}
