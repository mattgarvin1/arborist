package arborist

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// Return the list of JSON tags which are defined in this struct.
//
// **Example**
//
// ```go
// type City struct {
//     Name       string `json:"name"`
//     Population int    `json:"population,omitempty"`
// }
//
// c := City{"Chicago", 2700000}
// structJSONFields(c)
// // => {"name", "population,omitempty"}
// ```
func structJSONFields(x interface{}) map[string]struct{} {
	var structValue reflect.Value = reflect.ValueOf(x)
	if structValue.Kind() == reflect.Ptr {
		structValue = structValue.Elem()
	}
	var structType reflect.Type = structValue.Type()
	result := make(map[string]struct{})
	for i := 0; i < structValue.NumField(); i++ {
		field := structType.Field(i)
		jsonTag := field.Tag.Get("json")
		result[jsonTag] = struct{}{}
	}
	return result
}

// validateJSON checks that the input struct `x` has fields with JSON tags
// that exactly match the given content. If there are any fields in one and
// not the other an error is returned.
//
// Use this function to deserialize JSON when the JSON must contain exactly the
// fields specified in a given struct, by first unmarshalling some bytes to a
// `map[string]interface{}`, then calling this function on the struct in
// question and the map, and then finally assigning fields on the struct
// directly from the map.
func validateJSON(
	structName string,
	x interface{},
	content map[string]interface{},
	optionalFields map[string]struct{},
) error {
	if structName == "" {
		structName = reflect.ValueOf(x).Elem().Type().Name()
	}
	if optionalFields == nil {
		optionalFields = make(map[string]struct{})
	}

	expectFields := structJSONFields(x)
	// Because the fields might contain extra stuff like `omitempty`, we have
	// to clean these up to make sure it's just the tag names.
	for field := range expectFields {
		// If there's a field like `"tag,omitempty"` then delete that from
		// `expectFields`, and insert just `"tag"` back.
		split := strings.Split(field, ",")
		if len(split) > 1 {
			delete(expectFields, field)
		}
		expectFields[split[0]] = struct{}{}
	}

	// First, check that the content contains an entry for every field in the
	// input with a JSON tag.
	missingFields := []string{}
	for field := range expectFields {
		_, exists := content[field]
		_, optional := optionalFields[field]
		if !exists && !optional {
			missingFields = append(missingFields, field)
		}
	}
	if len(missingFields) > 0 {
		return missingRequiredFields(structName, missingFields)
	}

	// Now, check that the content does not contain any unexpected fields.
	unexpectedFields := []string{}
	for field := range content {
		if _, exists := expectFields[field]; !exists {
			unexpectedFields = append(unexpectedFields, field)
		}
	}
	if len(unexpectedFields) > 0 {
		return containsUnexpectedFields(structName, unexpectedFields)
	}

	return nil
}

func unmarshal(body []byte, x interface{}) *ErrorResponse {
	var structValue reflect.Value = reflect.ValueOf(x)
	if structValue.Kind() == reflect.Ptr {
		structValue = structValue.Elem()
	}
	var structType reflect.Type = structValue.Type()
	err := json.Unmarshal(body, x)
	if err != nil {
		msg := fmt.Sprintf(
			"could not parse %s from JSON; make sure input has correct types",
			structType,
		)
		response := newErrorResponse(msg, 400, &err)
		response.log.Info(
			"tried to create %s but input was invalid; offending JSON: %s",
			structType,
			loggableJSON(body),
		)
		return response
	}
	return nil
}
