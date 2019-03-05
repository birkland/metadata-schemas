package jsonschema

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/pkg/errors"
)

// finds references that are local to a doc
var localRefs = func(r ref) bool {
	return r.pointsIn() == ""
}

type dereferenceState struct {
	analyzed      map[string]Instance // nil entry for schemas in-progress
	fetcher       Fetcher
	localResolves int
}

func Dereference(fetcher Fetcher, schemas ...Instance) error {
	state := dereferenceState{
		analyzed: make(map[string]Instance),
		fetcher:  fetcher,
	}

	for _, schema := range schemas {
		if _, alreadyDone := state.analyzed[schema.ID()]; alreadyDone {
			continue
		}

		err := state.dereference(schema)
		if err != nil {
			return errors.Wrapf(err, "could not resolve references in %s", schema.ID())
		}
	}

	return nil
}

// Resolve all references in a given schema
func (d *dereferenceState) dereference(schema Instance) error {

	id := schema.ID()

	if _, found := d.analyzed[id]; found {
		// Should never happen, but just being safe
		return fmt.Errorf("%s has been dereferenced already", id)
	}
	d.analyzed[id] = nil

	a, err := d.analyze(schema)
	if err != nil {
		return errors.Wrapf(err, "could not analyze schema %s", schema.ID())
	}

	a, err = d.resolveLocal(a)
	if err != nil {
		return errors.Wrapf(err, "could not resolve local references")
	}

	for _, r := range a.refs {
		err := d.resolve(r, a.doc)
		if err != nil {
			return errors.Wrapf(err, "could not resolve external reference at %s", r.encounteredAt.String())
		}
	}

	d.analyzed[id] = schema
	return nil
}

// Recursively resolve all local references in a document.  We can replace any refs
// that are terminal (do not point to content that has another local $ref).  We keep replacing the
// terminal local refs until all are gone, or we detect a cycle.
func (d *dereferenceState) resolveLocal(a schemaAnalyzer) (schemaAnalyzer, error) {
	d.localResolves++
	if d.localResolves > 100 {
		return a, fmt.Errorf("schema resolution is stuck")
	}

	terminal := a.take(localTerminal(a))
	local := a.peek(localRefs)

	if len(local) == 0 && len(terminal) == 0 {
		return a, nil
	}

	if len(terminal) == 0 && len(local) > 0 {
		return a, fmt.Errorf("Cycle detected in local references in %s", a.doc.ID())
	}

	for _, r := range terminal {
		err := d.resolve(r, a.doc)
		if err != nil {
			return a, errors.Wrapf(err, "could not resolve reference %s in %s", r.encounteredAt.String(), a.doc.ID())
		}
	}

	// re-analyze and do another pass.
	nextRound, _ := d.analyze(a.doc)
	return d.resolveLocal(nextRound)
}

// Resolve a single $ref in a given schema instance
func (d *dereferenceState) resolve(r ref, inDoc Instance) (err error) {

	var src map[string]interface{}

	switch r.pointsIn() {
	case "":
		src = inDoc
	default:
		src, err = d.getSchema(r.pointsIn())
		if err != nil {
			return errors.Wrapf(err, "error resolving reference in schema %s", inDoc.ID())
		}
	}

	// Extract the value from the source (this doc, or an external one)
	if r.pointsTo.GetPointer().String() == r.encounteredAt.String() && r.pointsIn() == "" {
		return fmt.Errorf("Cycle detected: %s points to itself in %s", r.encounteredAt.String(), inDoc.ID())
	}
	value, _, err := r.pointsTo.GetPointer().Get(src)
	if err != nil {
		return errors.Wrapf(err, "could not read value at %s", r.pointsTo.GetUrl())
	}

	// Now set the value in this doc
	_, err = r.encounteredAt.Set(map[string]interface{}(inDoc), value)
	if err != nil {
		return errors.Wrapf(err, "could not insert content from %s at %s", r.pointsTo.GetUrl(), r.encounteredAt)
	}

	return nil
}

// Finds references that are local and terminal (i.e. do not contain any other $refs)
func localTerminal(a schemaAnalyzer) func(ref) bool {
	return func(r ref) bool {

		if r.pointsIn() != "" {
			return false
		}

		// Scan through all refs.  If we encounter a candidate that occurs beneath the place this
		// ref points to in the doc, then this ref is not terminal, as it contains a local ref inside it.
		refLoc := r.pointsTo.GetPointer().String()
		for _, t := range a.refs {
			candidate := t.encounteredAt.String()
			if strings.HasPrefix(candidate, refLoc) && candidate != refLoc && t.pointsIn() == "" {
				return false
			}
		}

		return true
	}
}

// analyze a schema and parse its references
func (d *dereferenceState) analyze(schema Instance) (schemaAnalyzer, error) {

	analyzer := schemaAnalyzer{
		doc: schema,
	}

	return analyzer, analyzer.findRefs()
}

// get a schema for a given URI.  If we have not encountered it yet, fetch and dereference it
// first.
func (d *dereferenceState) getSchema(schemaUri string) (Instance, error) {
	if a, ok := d.analyzed[schemaUri]; ok {
		if a == nil {
			return nil, fmt.Errorf("Cycle detected in schema that references %s", schemaUri)
		}

		return a, nil
	}

	if d.fetcher == nil {
		return nil, fmt.Errorf("No schema fetcher given")
	}

	uri, err := url.Parse(schemaUri)
	if err != nil {
		return nil, errors.Wrapf(err, "could not parse schema uri %s", schemaUri)
	}

	instance, ok, err := d.fetcher.GetSchema(uri)
	if err != nil {
		return nil, errors.Wrapf(err, "error fetching %s", schemaUri)
	}
	if !ok {
		return nil, fmt.Errorf(`could not find schema "%s"`, schemaUri)
	}

	return instance, d.dereference(instance)
}
