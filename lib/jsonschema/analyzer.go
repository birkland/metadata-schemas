package jsonschema

import (
	"fmt"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/pkg/errors"
	"github.com/xeipuuv/gojsonpointer"
	"github.com/xeipuuv/gojsonreference"
)

const refKey = "$ref"

type errs []error

func (e errs) Error() string {
	return fmt.Sprintf("%s", []error(e))
}

type ref struct {
	encounteredAt gojsonpointer.JsonPointer     // The location of the $ref in the schema in which it occurs
	pointsTo      gojsonreference.JsonReference // The json pointer the $ref points to
}

// returns the schema URI of the schema the $ref points to
func (r *ref) pointsIn() string {
	if r.pointsTo.HasFullUrl {
		return strings.Split(r.pointsTo.GetUrl().String(), "#")[0]
	}
	return ""
}

type pointer []string

func (p pointer) push(segment string) pointer {
	return append(p, segment)
}

func (p pointer) toGojsonPointer() gojsonpointer.JsonPointer {
	ptr, _ := gojsonpointer.NewJsonPointer(fmt.Sprintf("/%s", strings.Join(p, "/")))
	return ptr
}

type schemaAnalyzer struct {
	doc    Instance
	refs   []ref
	errors errs
}

func (c *schemaAnalyzer) findRefs() error {
	c._scanObj(nil, c.doc)

	if len(c.errors) > 0 {
		return c.errors
	}

	return nil
}

func (c *schemaAnalyzer) take(filter func(ref) bool) []ref {
	var keep, taken []ref

	for _, r := range c.refs {
		switch filter(r) {
		case true:
			taken = append(taken, r)
		case false:
			keep = append(keep, r)
		}
	}

	c.refs = keep
	return taken
}

func (c *schemaAnalyzer) peek(filter func(ref) bool) []ref {
	var matches []ref
	for _, r := range c.refs {
		if filter(r) {
			matches = append(matches, r)
		}
	}
	return matches
}

func (c *schemaAnalyzer) _scanObj(p pointer, obj map[string]interface{}) {
	for k, v := range obj {
		if k == refKey {
			c._addRef(p, v)
		} else {
			c._scan(p.push(k), v)
		}
	}
}

func (c *schemaAnalyzer) _scan(p pointer, something interface{}) {
	if something == nil {
		return
	}

	switch v := something.(type) {
	case map[string]interface{}:
		c._scanObj(p, v)
	case []interface{}:
		c._scanList(p, v)
	default:
	}
}

func (c *schemaAnalyzer) _scanList(p pointer, list []interface{}) {
	for i, v := range list {
		c._scan(p.push(strconv.Itoa(i)), v)
	}
}

// we've found a {"$ref": <refContent> }, where <refContent> is hopefully a http URI
// like "#/path/to/foo" or "whatever.json#/path/to/foo", or "http://example.org/whatever.json#/path/to/foo"
func (c *schemaAnalyzer) _addRef(p pointer, refContent interface{}) {

	refString, ok := refContent.(string)
	if !ok {
		c.errors = append(c.errors,
			fmt.Errorf("expected to find a string in $ref at %s, instead found %s", p.toGojsonPointer(), refContent))
		return
	}

	jsonref, err := gojsonreference.NewJsonReference(refString)
	if err != nil {
		c.errors = append(c.errors,
			errors.Wrapf(err, "could not parse json reference at %s with content %s", p.toGojsonPointer(), refString))
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

		c.refs = append(c.refs, ref{pointsTo: absoluteRef, encounteredAt: p.toGojsonPointer()})
		return
	}

	c.refs = append(c.refs, ref{pointsTo: jsonref, encounteredAt: p.toGojsonPointer()})
}
