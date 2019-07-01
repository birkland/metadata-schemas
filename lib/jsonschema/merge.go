package jsonschema

import (
	"reflect"

	"github.com/pkg/errors"
)

// JSON schema fields that are ignorable for the sake of merging
var ignorable = map[string]bool{"title": true, "description": true, "$id": true, "$schema": true, "$comment": true}

// Merge merges multiple schema instances into a single one, folding in properties that do not
// overlap, and
func Merge(schemas []Instance) (Instance, error) {
	var result mergeableMap = make(map[string]interface{})

	for _, toMerge := range schemas {
		for field, value := range toMerge {
			if !ignorable[field] {
				err := result.mergeIn(field, value)
				if err != nil {
					return nil, errors.Wrapf(err, "could not merge field '%s' from schema with ID '%s'", field, toMerge.ID())
				}
			}
		}
	}

	return Instance(result), nil
}

// mergeableMap is a JSON map intended for merging in values
type mergeableMap map[string]interface{}

// mergeIn a value into a given field
func (m mergeableMap) mergeIn(field string, value interface{}) error {
	_, exists := m[field]

	// Create empty map or array if none exists already, where appropriate
	if !exists {
		switch v := value.(type) {
		case []interface{}:
			m[field] = make([]interface{}, 0, len(v))
		case map[string]interface{}:
			m[field] = make(map[string]interface{}, len(v))
		}
	}

	// Check for type clash
	if m[field] != nil && reflect.TypeOf(m[field]) != reflect.TypeOf(value) {
		return errors.Errorf("type conflict for property '%s': %s vs %s", field,
			reflect.TypeOf(m[field]).Name(), reflect.TypeOf(value).Name())
	}

	// Now do the merge!
	// For arrays, copy in values if they are novel
	// For maps (JSON objects), merge each field
	// For simple values, just copy
	switch val := value.(type) {
	case []interface{}:
		for _, v := range val {
			m[field] = addIfNotPresent(v, m[field].([]interface{}))
		}
	case map[string]interface{}:
		existingMap := mergeableMap(m[field].(map[string]interface{}))
		for k, v := range val {
			err := existingMap.mergeIn(k, v)
			if err != nil {
				return errors.Wrapf(err, "could not merge into map at '%s'", field)
			}
		}
	default:
		if !ignorable[field] && m[field] != nil && m[field] != value {
			return errors.Errorf("value conflict in property '%s': %s != %s", field, m[field], val)
		}
		m[field] = value
	}

	return nil
}

// Add to a slice only if a value doesn't exist
func addIfNotPresent(value interface{}, container []interface{}) []interface{} {

	for _, existing := range container {
		if reflect.DeepEqual(existing, value) {
			return container
		}
	}

	// We don't need to defensively copy, as these are never mutated
	return append(container, value)
}
