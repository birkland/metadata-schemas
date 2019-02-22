package jsonschema

const idKey = "$id"

// Instance is an instance of a parsed JSON schema (in generic map[string]interface{} form)
type Instance map[string]interface{}

// ID returns the identifier ("$id") of the schema, if present
func (s Instance) ID() string {
	if id, ok := s[idKey]; ok {
		if v, ok := id.(string); ok {
			return v
		}
	}
	return ""
}
