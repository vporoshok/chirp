package chirp

import (
	"fmt"
	"reflect"
)

// Error of parsing field
type Error struct {
	part   string
	field  reflect.StructField
	source string
	cause  error
}

// Error implements error interface
func (err Error) Error() string {

	return fmt.Sprintf("%s[%s](%s): %s", err.part, err.Tag(), err.source, err.cause)
}

// Part of request with error (path | query | body)
func (err Error) Part() string {

	return err.part
}

// Field return field reflector
func (err Error) Field() reflect.StructField {

	return err.field
}

// Tag name or field name
//
// If error has been occurred on json unmarshaling it return "*"
func (err Error) Tag() string {
	if len(err.field.Name) == 0 {

		return "*"
	}
	part := err.part
	if part == "body" {
		part = "json"
	}
	tag, _ := tagOrFieldName(err.field, part)

	return tag
}

// Source of parse
func (err Error) Source() string {

	return err.source
}

// Cause implement errors.Causer interface
func (err Error) Cause() error {

	return err.cause
}
